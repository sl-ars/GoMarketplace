package app

import (
	"go-app-marketplace/internal/app/config"
	"go-app-marketplace/internal/app/connections"
	"go-app-marketplace/internal/app/start"
	"go-app-marketplace/internal/deliveries/http"
	"go-app-marketplace/internal/repositories"
	"go-app-marketplace/internal/services"
	"go-app-marketplace/internal/usecases"
	"log"
)

func Run(configFiles ...string) {
	// Load env
	cfg, err := config.NewConfig(configFiles...)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize DB and external connections
	conns, err := connections.NewConnections(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize connections: %v", err)
	}
	defer conns.Close()

	// Dependency injection
	userRepo := repositories.NewUserPostgresRepo(conns.DB)
	userUC := usecases.NewUserUseCase(userRepo)
	userService := services.NewUserService(userUC, cfg.JWTSecret)

	productRepo := repositories.NewProductRepository(conns.DB)
	productUC := usecases.NewProductUseCase(productRepo)
	productService := services.NewProductService(productUC)

	offerRepo := repositories.NewOfferRepository(conns.DB)
	offerUC := usecases.NewOfferUseCase(offerRepo)
	offerService := services.NewOfferService(offerUC)

	cartRepo := repositories.NewCartRepository(conns.DB)
	cartUC := usecases.NewCartUseCase(cartRepo, offerRepo)
	cartService := services.NewCartService(cartUC)

	orderRepo := repositories.NewOrderRepository(conns.DB)
	orderUC := usecases.NewOrderUsecase(orderRepo, cartRepo, offerRepo)
	orderService := services.NewOrderService(orderUC)

	// Stripe Payment Service
	paymentService := services.NewPaymentService(cfg.StripeSecretKey, cfg.StripeWebhookSecret)

	// Set the payment service on the order service to avoid circular dependency
	orderService.SetPaymentService(paymentService)

	// refund
	refundRepo := repositories.NewRefundRepository(conns.DB)
	refundUC := usecases.NewRefundUsecase(refundRepo, orderRepo)
	refundService := services.NewRefundService(refundUC)
	// Wrap services
	svc := &http.Services{
		User:    userService,
		Cart:    cartService,
		Product: productService,
		Offer:   offerService,
		Order:   orderService,
		Payment: paymentService,
		Refund:  refundService,
		JWTKey:  []byte(cfg.JWTSecret),
	}

	// Router
	router := http.NewRouter(svc)

	// Start the server
	start.StartHTTPServer(cfg.HTTPServer.Port, router)
}
