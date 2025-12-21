package product

import (
	"go-app-marketplace/pkg/logger"

	"github.com/gorilla/mux"
)

func RegisterProductRoutes(r *mux.Router, handler *ProductHandler, jwtSecret []byte, log *logger.Logger) {
	// Public endpoints (read-only)
	public := r.PathPrefix("/products").Subrouter()
	public.HandleFunc("", handler.ListProducts).Methods("GET")
	public.HandleFunc("/search", handler.SearchProducts).Methods("GET")
	public.HandleFunc("/{id}", handler.GetProduct).Methods("GET")

	// Note: Admin product CRUD endpoints are now in /api/admin/products (see admin_routes.go)
}
