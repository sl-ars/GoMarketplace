package user

import (
	"github.com/gorilla/mux"
	"go-app-marketplace/internal/middleware"
	"go-app-marketplace/internal/services"
)

func RegisterUserRoutes(r *mux.Router, service *services.UserService, jwtSecret []byte) {
	// Public routes
	r.HandleFunc("/register", RegisterHandler(service)).Methods("POST")
	r.HandleFunc("/login", LoginHandler(service)).Methods("POST")
	r.HandleFunc("/refresh", RefreshHandler(service)).Methods("POST")
	r.HandleFunc("/verify", VerifyHandler(service)).Methods("GET")

	// Protected routes
	protected := r.PathPrefix("/").Subrouter()
	protected.Use(middleware.AuthMiddleware(jwtSecret))
	protected.HandleFunc("/me", GetCurrentUserHandler(service)).Methods("GET")
}
