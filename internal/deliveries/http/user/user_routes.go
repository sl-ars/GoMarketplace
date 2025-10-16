package user

import (
	"github.com/gorilla/mux"
	"go-app-marketplace/internal/middleware"
	"go-app-marketplace/internal/services"
	"go-app-marketplace/pkg/logger"
)

func RegisterUserRoutes(r *mux.Router, service *services.UserService, jwtSecret []byte, log *logger.Logger) {
	// Public routes
	r.HandleFunc("/register", RegisterHandler(service, log)).Methods("POST")
	r.HandleFunc("/login", LoginHandler(service, log)).Methods("POST")
	r.HandleFunc("/refresh", RefreshHandler(service, log)).Methods("POST")
	r.HandleFunc("/verify", VerifyHandler(service, log)).Methods("GET")

	// Protected routes
	protected := r.PathPrefix("/").Subrouter()
	protected.Use(middleware.AuthMiddleware(jwtSecret, log))
	protected.HandleFunc("/me", GetCurrentUserHandler(service, log)).Methods("GET")
}
