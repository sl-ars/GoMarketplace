package admin

import (
	"go-app-marketplace/internal/middleware"
	"go-app-marketplace/pkg/domain"
	"go-app-marketplace/pkg/logger"

	"github.com/gorilla/mux"
)

// RegisterRoutes registers all admin routes
func RegisterRoutes(r *mux.Router, handler *Handler, jwtSecret []byte, log *logger.Logger) {
	// All admin routes require authentication and admin role
	admin := r.PathPrefix("/admin").Subrouter()
	admin.Use(middleware.AuthMiddleware(jwtSecret, log))
	admin.Use(middleware.RequireRoles(domain.UserRoleAdmin))

	// =========================================================================
	// Dashboard
	// =========================================================================
	admin.HandleFunc("/dashboard", handler.GetDashboard).Methods("GET")
	admin.HandleFunc("/stats/users", handler.GetUserStats).Methods("GET")

	// =========================================================================
	// User Management
	// =========================================================================
	admin.HandleFunc("/users", handler.ListUsers).Methods("GET")
	admin.HandleFunc("/users/{id}", handler.GetUser).Methods("GET")
	admin.HandleFunc("/users/{id}", handler.DeleteUser).Methods("DELETE")
	admin.HandleFunc("/users/{id}/ban", handler.BanUser).Methods("POST")
	admin.HandleFunc("/users/{id}/unban", handler.UnbanUser).Methods("POST")
	admin.HandleFunc("/users/{id}/role", handler.UpdateUserRole).Methods("PUT")

	// =========================================================================
	// Product Management
	// =========================================================================
	admin.HandleFunc("/products", handler.CreateProduct).Methods("POST")
	admin.HandleFunc("/products", handler.ListProducts).Methods("GET")
	admin.HandleFunc("/products/{id}", handler.GetProduct).Methods("GET")
	admin.HandleFunc("/products/{id}", handler.UpdateProduct).Methods("PUT")
	admin.HandleFunc("/products/{id}", handler.DeleteProduct).Methods("DELETE")
}
