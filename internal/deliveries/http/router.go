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
}

func NewRouter(s *Services) http.Handler {
	r := mux.NewRouter()

	// Add logging middleware
	r.Use(middleware.LoggingMiddleware(s.Logger))

	// API routes
	api := r.PathPrefix("/api").Subrouter()

	// Healthcheck
	api.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Get JWT key from auth service
	jwtKey := s.Auth.GetJWTKey()

	// Auth routes (public)
	authHandler := auth.NewHandler(s.Auth, s.Logger)
	auth.RegisterRoutes(api.PathPrefix("/").Subrouter(), authHandler)

	// User routes (protected)
	userHandler := user.NewHandler(s.User, s.Logger)
	user.RegisterRoutes(api.PathPrefix("/").Subrouter(), userHandler, jwtKey, s.Logger)

	// Cart routes
	cartHandler := cart.NewCartHandler(s.Cart, s.Logger)
	cart.RegisterCartRoutes(api.PathPrefix("/").Subrouter(), cartHandler, jwtKey, s.Logger)

	// Product routes
	productHandler := product.NewProductHandler(s.Product, s.Offer, s.Logger)
	product.RegisterProductRoutes(api.PathPrefix("/").Subrouter(), productHandler, jwtKey, s.Logger)

	// Offer routes
	offerHandler := offer.NewOfferHandler(s.Offer, s.Logger)
	offer.RegisterOfferRoutes(api.PathPrefix("/").Subrouter(), offerHandler, jwtKey, s.Logger)

	// Order routes
	orderHandler := order.NewOrderHandler(s.Order, s.Logger)
	order.RegisterOrderRoutes(api.PathPrefix("/").Subrouter(), orderHandler, jwtKey, s.Logger)

	// Refund routes
	refundHandler := refund.NewHandler(s.Refund, s.Logger)
	refund.Register(api.PathPrefix("/").Subrouter(), refundHandler, jwtKey, s.Logger)

	// Stripe Webhook Handler
	stripeWebhookHandler := webhook.NewStripeWebhookHandler(s.Order, s.Payment.GetWebhookSecret())
	r.HandleFunc("/api/webhook/stripe", stripeWebhookHandler.HandleWebhook).Methods("POST")

	// Swagger
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	return r
}
