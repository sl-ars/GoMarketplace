package auth

import (
	"github.com/gorilla/mux"
)

// RegisterRoutes registers all authentication routes
func RegisterRoutes(r *mux.Router, handler *Handler) {
	// All auth routes are public (no authentication required)
	r.HandleFunc("/auth/register", handler.Register).Methods("POST")
	r.HandleFunc("/auth/login", handler.Login).Methods("POST")
	r.HandleFunc("/auth/refresh", handler.Refresh).Methods("POST")
	r.HandleFunc("/auth/verify", handler.Verify).Methods("GET")

	// Email verification
	r.HandleFunc("/auth/verify-email", handler.VerifyEmail).Methods("POST", "GET")
	r.HandleFunc("/auth/resend-verification", handler.ResendVerification).Methods("POST")

	// Password reset
	r.HandleFunc("/auth/forgot-password", handler.ForgotPassword).Methods("POST")
	r.HandleFunc("/auth/reset-password", handler.ResetPassword).Methods("POST")
}
