package app

import (
	"context"
	"log"
	"time"

	"go-app-marketplace/internal/app/config"
	"go-app-marketplace/internal/app/connections"
	"go-app-marketplace/internal/app/start"
	"go-app-marketplace/internal/deliveries/http"
	"go-app-marketplace/internal/elasticsearch"
	"go-app-marketplace/internal/messagebus"
	"go-app-marketplace/internal/middleware"
	"go-app-marketplace/internal/redisdb"
	"go-app-marketplace/internal/repositories"
	"go-app-marketplace/internal/services"
	"go-app-marketplace/internal/usecases"
	"go-app-marketplace/pkg/logger"
	"go-app-marketplace/pkg/ratelimit"
)

func Run(configFiles ...string) {
	// Load env
	cfg, err := config.NewConfig(configFiles...)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	appLogger := logger.New(cfg.Logger)
	appLogger.Info("Starting application")

	// Initialize DB and Redis connections
	conns, err := connections.NewConnections(cfg)
	if err != nil {
		appLogger.WithError(err).Fatal("Failed to initialize connections")
	}
	defer conns.Close()

	// Set Redis client for cache utilities
	redisdb.SetClient(conns.Redis)
	appLogger.Info("Database and Redis connections established")

	// Initialize rate limiter
	var rateLimiter *ratelimit.RedisLimiter
	if cfg.RateLimit.Enabled {
		rateLimiter = ratelimit.NewRedisLimiter(conns.Redis, "marketplace:ratelimit")
		appLogger.Info("Rate limiter initialized")
	}

	// Initialize trusted proxies for secure IP extraction
	var trustedProxies *middleware.TrustedProxies
	if len(cfg.RateLimit.TrustedProxies) > 0 {
		trustedProxies = middleware.NewTrustedProxies(cfg.RateLimit.TrustedProxies)
		appLogger.WithField("proxies", cfg.RateLimit.TrustedProxies).Info("Trusted proxies configured")
	} else {
		appLogger.Warn("No trusted proxies configured - X-Forwarded-For headers will be ignored")
	}

	// Build rate limit settings from config
	rateLimitSettings := http.RateLimitSettings{
		Enabled: cfg.RateLimit.Enabled,
		Auth: ratelimit.Config{
			Requests: cfg.RateLimit.AuthRequests,
			Window:   time.Minute,
		},
		Public: ratelimit.Config{
			Requests: cfg.RateLimit.PublicRequests,
			Window:   time.Minute,
		},
		Standard: ratelimit.Config{
			Requests: cfg.RateLimit.StandardRequests,
			Window:   time.Minute,
		},
		TrustedProxies: trustedProxies,
	}

	// --- RabbitMQ connection ---
	rmqConn, rmqCh, err := connections.NewRabbitMQConn(cfg.RabbitMQURL)
	if err != nil {
		appLogger.WithError(err).Fatal("failed to connect to RabbitMQ")
	}
	defer rmqConn.Close()
	defer rmqCh.Close()

	orderPublisher, err := messagebus.NewRabbitMQOrderPublisher(
		rmqCh,
		"app.exchange",   // exchange name
		"orders.created", // routing key
	)
	if err != nil {
		appLogger.WithError(err).Fatal("failed to create order publisher")
	}

	// Initialize email publisher (for async email sending via RabbitMQ)
	var emailPublisher *messagebus.EmailPublisher
	if cfg.Email.Enabled {
		emailPublisher, err = messagebus.NewEmailPublisher(rmqCh, "emails.exchange")
		if err != nil {
			appLogger.WithError(err).Fatal("failed to create email publisher")
		}
		appLogger.Info("Email publisher initialized (via RabbitMQ)")
	} else {
		appLogger.Warn("Email sending is disabled - set EMAIL_ENABLED=true to enable")
	}

	// Dependency injection
	userRepo := repositories.NewUserPostgresRepo(conns.DB, appLogger)

	// Auth service (handles registration, login, token refresh/verify, email verification, password reset)
	authUC := usecases.NewAuthUseCase(userRepo)
	authService := services.NewAuthService(authUC, emailPublisher, services.AuthServiceConfig{
		JWTSecret:       cfg.JWTSecret,
		BaseURL:         cfg.Auth.BaseURL,
		RequireVerified: cfg.Auth.RequireEmailVerified,
	}, appLogger)

	// User service (handles user profile operations)
	userUC := usecases.NewUserUseCase(userRepo)
	userService := services.NewUserService(userUC, appLogger)

	productRepo := repositories.NewProductRepository(conns.DB)
	productUC := usecases.NewProductUseCase(productRepo)
	productService := services.NewProductService(productUC)

	// Initialize Elasticsearch
	var productSearchService *elasticsearch.ProductSearchService
	if cfg.Elasticsearch.Enabled {
		esClient, err := elasticsearch.NewClient(
			cfg.Elasticsearch.Addresses,
			cfg.Elasticsearch.Username,
			cfg.Elasticsearch.Password,
			appLogger.Logger,
		)
		if err != nil {
			appLogger.WithError(err).Warn("Failed to initialize Elasticsearch, search will be disabled")
		} else {
			productSearchService = elasticsearch.NewProductSearchService(esClient)
			if err := productSearchService.InitializeIndex(context.Background()); err != nil {
				appLogger.WithError(err).Warn("Failed to initialize product search index")
			} else {
				appLogger.Info("Elasticsearch product search initialized")
				productService.SetSearchService(productSearchService)
			}
		}
	}

	offerRepo := repositories.NewOfferRepository(conns.DB)
	offerUC := usecases.NewOfferUseCase(offerRepo)
	offerService := services.NewOfferService(offerUC)

	cartRepo := repositories.NewCartRepository(conns.DB)
	cartUC := usecases.NewCartUseCase(cartRepo, offerRepo)
	cartService := services.NewCartService(cartUC)

	orderRepo := repositories.NewOrderRepository(conns.DB)
	orderUC := usecases.NewOrderUsecase(orderRepo, cartRepo, offerRepo, orderPublisher)
	orderService := services.NewOrderService(orderUC)

	// Stripe Payment Service
	paymentService := services.NewPaymentService(cfg.StripeSecretKey, cfg.StripeWebhookSecret)

	// Set the payment service on the order service to avoid circular dependency
	orderService.SetPaymentService(paymentService)

	// Refund service
	refundRepo := repositories.NewRefundRepository(conns.DB)
	refundUC := usecases.NewRefundUsecase(refundRepo, orderRepo)
	refundService := services.NewRefundService(refundUC)

	// Wrap services
	svc := &http.Services{
		Auth:            authService,
		User:            userService,
		Cart:            cartService,
		Product:         productService,
		Offer:           offerService,
		Order:           orderService,
		Payment:         paymentService,
		Refund:          refundService,
		Logger:          appLogger,
		RateLimiter:     rateLimiter,
		RateLimitConfig: rateLimitSettings,
	}

	// Router
	router := http.NewRouter(svc)

	// Start the server
	appLogger.WithField("port", cfg.HTTPServer.Port).Info("Starting HTTP server")
	start.StartHTTPServer(cfg.HTTPServer.Port, router)
}
