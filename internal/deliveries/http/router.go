package http

import (
	"net/http"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "go-app-marketplace/docs"
	"go-app-marketplace/internal/deliveries/http/auth"
	"go-app-marketplace/internal/deliveries/http/cart"
	"go-app-marketplace/internal/deliveries/http/offer"
	"go-app-marketplace/internal/deliveries/http/order"
	"go-app-marketplace/internal/deliveries/http/product"
	"go-app-marketplace/internal/deliveries/http/refund"
	"go-app-marketplace/internal/deliveries/http/user"
	"go-app-marketplace/internal/deliveries/http/webhook"
	"go-app-marketplace/internal/middleware"
	"go-app-marketplace/internal/services"
	"go-app-marketplace/pkg/logger"
	"go-app-marketplace/pkg/ratelimit"
)

type Services struct {
	Auth    *services.AuthService
	User    *services.UserService
	Cart    *services.CartService
	Product *services.ProductService
	Offer   *services.OfferService
	Order   *services.OrderService
	Payment *services.PaymentService
	Refund  *services.RefundService
	Logger  *logger.Logger

	// Rate limiting
	RateLimiter     *ratelimit.RedisLimiter
	RateLimitConfig RateLimitSettings
}

// RateLimitSettings holds rate limit configuration
type RateLimitSettings struct {
	Enabled  bool
	Auth     ratelimit.Config
	Public   ratelimit.Config
	Standard ratelimit.Config
}

func NewRouter(s *Services) http.Handler {
	r := mux.NewRouter()

	// Add logging middleware
	r.Use(middleware.LoggingMiddleware(s.Logger))

	// Rate limit config for middleware
	rlConfig := &middleware.RateLimitConfig{
		Limiter: s.RateLimiter,
		Logger:  s.Logger,
	}

	// API routes
	api := r.PathPrefix("/api").Subrouter()

	// Apply global rate limiting if enabled
	if s.RateLimitConfig.Enabled && s.RateLimiter != nil {
		api.Use(middleware.RateLimitMiddleware(rlConfig, s.RateLimitConfig.Standard))
	}

	// Healthcheck (no rate limit)
	api.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Get JWT key from auth service
	jwtKey := s.Auth.GetJWTKey()

	// =========================================================================
	// Auth routes (public, but with STRICT rate limiting for security)
	// =========================================================================
	authRouter := api.PathPrefix("/auth").Subrouter()
	if s.RateLimitConfig.Enabled && s.RateLimiter != nil {
		// Override with stricter auth rate limit
		authRouter.Use(middleware.RateLimitByEndpoint(rlConfig, s.RateLimitConfig.Auth))
	}
	authHandler := auth.NewHandler(s.Auth, s.Logger)
	authRouter.HandleFunc("/register", authHandler.Register).Methods("POST")
	authRouter.HandleFunc("/login", authHandler.Login).Methods("POST")
	authRouter.HandleFunc("/refresh", authHandler.Refresh).Methods("POST")
	authRouter.HandleFunc("/verify", authHandler.Verify).Methods("GET")

	// =========================================================================
	// User routes (protected)
	// =========================================================================
	userHandler := user.NewHandler(s.User, s.Logger)
	user.RegisterRoutes(api.PathPrefix("/").Subrouter(), userHandler, jwtKey, s.Logger)

	// =========================================================================
	// Cart routes (protected)
	// =========================================================================
	cartHandler := cart.NewCartHandler(s.Cart, s.Logger)
	cart.RegisterCartRoutes(api.PathPrefix("/").Subrouter(), cartHandler, jwtKey, s.Logger)

	// =========================================================================
	// Product routes (public read, admin write)
	// =========================================================================
	productHandler := product.NewProductHandler(s.Product, s.Offer, s.Logger)
	product.RegisterProductRoutes(api.PathPrefix("/").Subrouter(), productHandler, jwtKey, s.Logger)

	// =========================================================================
	// Offer routes (seller only)
	// =========================================================================
	offerHandler := offer.NewOfferHandler(s.Offer, s.Logger)
	offer.RegisterOfferRoutes(api.PathPrefix("/").Subrouter(), offerHandler, jwtKey, s.Logger)

	// =========================================================================
	// Order routes (protected)
	// =========================================================================
	orderHandler := order.NewOrderHandler(s.Order, s.Logger)
	order.RegisterOrderRoutes(api.PathPrefix("/").Subrouter(), orderHandler, jwtKey, s.Logger)

	// =========================================================================
	// Refund routes (protected)
	// =========================================================================
	refundHandler := refund.NewHandler(s.Refund, s.Logger)
	refund.Register(api.PathPrefix("/").Subrouter(), refundHandler, jwtKey, s.Logger)

	// =========================================================================
	// Stripe Webhook (no rate limit - Stripe handles retries)
	// =========================================================================
	stripeWebhookHandler := webhook.NewStripeWebhookHandler(s.Order, s.Payment.GetWebhookSecret())
	r.HandleFunc("/api/webhook/stripe", stripeWebhookHandler.HandleWebhook).Methods("POST")

	// Swagger
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	return r
}
