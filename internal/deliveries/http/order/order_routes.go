package order

import (
	"github.com/gorilla/mux"
	"go-app-marketplace/internal/middleware"
	"go-app-marketplace/pkg/domain"
	"net/http"
)

func RegisterOrderRoutes(r *mux.Router, h *OrderHandler, jwtKey []byte) {

	buyer := r.PathPrefix("/orders").Subrouter()
	buyer.Use(middleware.AuthMiddleware(jwtKey))

	buyer.HandleFunc("/checkout", h.Checkout).Methods(http.MethodPost)

	buyer.HandleFunc("/checkout/{id:[0-9]+}", h.CheckoutExistingOrder).Methods(http.MethodPost)

	buyer.HandleFunc("/{id:[0-9]+}/cancel", h.CancelOrderItem).Methods(http.MethodPost)

	buyer.HandleFunc("/{id:[0-9]+}", h.GetOrder).Methods(http.MethodGet)

	buyer.HandleFunc("", h.ListOrders).Methods(http.MethodGet)

	seller := r.PathPrefix("/seller").Subrouter()
	seller.Use(middleware.AuthMiddleware(jwtKey))
	seller.Use(middleware.RequireRoles(domain.UserRoleSeller))

	seller.HandleFunc("/orders", h.ListSellerOrderItems).Methods(http.MethodGet)

	seller.HandleFunc("/orders/items/{id:[0-9]+}/status",
		h.UpdateOrderItemStatus).Methods(http.MethodPatch)
}
