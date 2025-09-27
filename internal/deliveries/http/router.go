package http

import (
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	_ "go-app-marketplace/docs"
	"go-app-marketplace/internal/deliveries/http/user"
	"go-app-marketplace/internal/services"
	"net/http"
)

type Services struct {
	User *services.UserService
}

func NewRouter(s *Services) http.Handler {
	r := mux.NewRouter()

	// Healthcheck
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	}).Methods("GET")

	// User routes
	user.RegisterUserRoutes(r.PathPrefix("/").Subrouter(), s.User)

	// Swagger
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	return r
}
