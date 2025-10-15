// internal/deliveries/http/refund/refund_routes.go
package refund

import (
	"github.com/gorilla/mux"
	"go-app-marketplace/internal/middleware"
	"go-app-marketplace/pkg/domain"
)

func Register(r *mux.Router, h *Handler, jwt []byte) {
	// /api/refunds
	sub := r.PathPrefix("/refunds").Subrouter()
	sub.Use(middleware.AuthMiddleware(jwt))

	// Customer
	sub.HandleFunc("/{item_id:[0-9]+}", h.Request).Methods("POST")

	// Seller confirms
	seller := sub.PathPrefix("").Subrouter()
	seller.Use(middleware.RequireRoles(domain.UserRoleSeller))
	seller.HandleFunc("/{refund_id:[0-9]+}/decide", h.Decide).Methods("PATCH")

}
