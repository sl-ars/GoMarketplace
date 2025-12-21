package offer

import (
	"github.com/gorilla/mux"
	"go-app-marketplace/internal/middleware"
	"go-app-marketplace/pkg/domain"
	"go-app-marketplace/pkg/logger"
)

func RegisterOfferRoutes(r *mux.Router, handler *OfferHandler, jwtSecret []byte, log *logger.Logger) {
	offerRouter := r.PathPrefix("/offers").Subrouter()
	offerRouter.Use(middleware.AuthMiddleware(jwtSecret, log))
	offerRouter.Use(middleware.RequireRoles(domain.UserRoleSeller))

	offerRouter.HandleFunc("", handler.CreateOffer).Methods("POST")
	offerRouter.HandleFunc("/{id:[0-9]+}", handler.GetOffer).Methods("GET")
	offerRouter.HandleFunc("/{id:[0-9]+}", handler.UpdateOffer).Methods("PUT")
	offerRouter.HandleFunc("/{id:[0-9]+}", handler.DeleteOffer).Methods("DELETE")
	offerRouter.HandleFunc("/me", handler.ListMyOffers).Methods("GET")
}
