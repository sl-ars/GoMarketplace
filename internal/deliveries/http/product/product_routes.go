package product

import (
	"github.com/gorilla/mux"
	"go-app-marketplace/internal/middleware"
	"go-app-marketplace/pkg/domain"
)

func RegisterProductRoutes(r *mux.Router, handler *ProductHandler, jwtSecret []byte) {
	// Public endpoints
	public := r.PathPrefix("/products").Subrouter()
	public.HandleFunc("", handler.ListProducts).Methods("GET")
	public.HandleFunc("/{id}", handler.GetProduct).Methods("GET")

	// Admin endpoints
	admin := r.PathPrefix("/admin/products").Subrouter()
	admin.Use(middleware.AuthMiddleware(jwtSecret))
	admin.Use(middleware.RequireRoles(domain.UserRoleAdmin))

	admin.HandleFunc("", handler.CreateProduct).Methods("POST")
}
