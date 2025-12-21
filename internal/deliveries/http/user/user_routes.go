package user

import (
	"github.com/gorilla/mux"

	"go-app-marketplace/internal/middleware"
	"go-app-marketplace/pkg/logger"
)

// RegisterRoutes registers user routes (all protected)
func RegisterRoutes(r *mux.Router, handler *Handler, jwtSecret []byte, log *logger.Logger) {
	// All user routes require authentication
	protected := r.PathPrefix("/users").Subrouter()
	protected.Use(middleware.AuthMiddleware(jwtSecret, log))
	protected.HandleFunc("/me", handler.GetCurrentUser).Methods("GET")
}
