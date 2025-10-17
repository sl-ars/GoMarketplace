package http

import (
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	_ "go-app-marketplace/docs"
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
	"net/http"
)

type Services struct {
	User    *services.UserService
	Cart    *services.CartService
	Product *services.ProductService
	Offer   *services.OfferService
	Order   *services.OrderService
	Payment *services.PaymentService
	Refund  *services.RefundService
	JWTKey  []byte
	Logger  *logger.Logger
}

func NewRouter(s *Services) http.Handler {
	r := mux.NewRouter()

	// Добавляем middleware логирования
	r.Use(middleware.LoggingMiddleware(s.Logger))

	// API routes
	api := r.PathPrefix("/api").Subrouter()

	// Healthcheck
	api.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	}).Methods("GET")

	// User routes
	user.RegisterUserRoutes(api.PathPrefix("/").Subrouter(), s.User, s.JWTKey, s.Logger)

	// Cart routes
	cartHandler := cart.NewCartHandler(s.Cart)
	cart.RegisterCartRoutes(api.PathPrefix("/").Subrouter(), cartHandler, s.JWTKey, s.Logger)

	// Product routes
	productHandler := product.NewProductHandler(s.Product, s.Offer)
	product.RegisterProductRoutes(api.PathPrefix("/").Subrouter(), productHandler, s.JWTKey, s.Logger)

	// Offer routes
	offerHandler := offer.NewOfferHandler(s.Offer)
	offer.RegisterOfferRoutes(api.PathPrefix("/").Subrouter(), offerHandler, s.JWTKey, s.Logger)

	// Order routes
	orderHandler := order.NewOrderHandler(s.Order)
	order.RegisterOrderRoutes(api.PathPrefix("/").Subrouter(), orderHandler, s.JWTKey, s.Logger)

	// Refund routes
	refundHandler := refund.NewHandler(s.Refund)
	refund.Register(api.PathPrefix("/").Subrouter(), refundHandler, s.JWTKey, s.Logger)

	// Stripe Webhook Handler
	stripeWebhookHandler := webhook.NewStripeWebhookHandler(s.Order, s.Payment.GetWebhookSecret())
	r.HandleFunc("/api/webhook/stripe", stripeWebhookHandler.HandleWebhook).Methods("POST")

	// Swagger
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	return r
}
