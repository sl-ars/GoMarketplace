package product

import (
	"go-app-marketplace/internal/middleware"
	"go-app-marketplace/pkg/domain"
	"go-app-marketplace/pkg/logger"

	"github.com/gorilla/mux"
)

func RegisterProductRoutes(r *mux.Router, handler *ProductHandler, jwtSecret []byte, log *logger.Logger) {
	// Public endpoints
	public := r.PathPrefix("/products").Subrouter()
	public.HandleFunc("", handler.ListProducts).Methods("GET")
	public.HandleFunc("/search", handler.SearchProducts).Methods("GET")
	public.HandleFunc("/{id}", handler.GetProduct).Methods("GET")

	// Admin endpoints
	admin := r.PathPrefix("/admin/products").Subrouter()
	admin.Use(middleware.AuthMiddleware(jwtSecret, log))
	admin.Use(middleware.RequireRoles(domain.UserRoleAdmin))

	admin.HandleFunc("", handler.CreateProduct).Methods("POST")
}
