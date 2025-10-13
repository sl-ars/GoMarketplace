package cart

import (
	"github.com/gorilla/mux"
	"go-app-marketplace/internal/middleware"
	"net/http"
)

func RegisterCartRoutes(r *mux.Router, h *CartHandler, jwtSecret []byte) {
	cartRouter := r.PathPrefix("/cart").Subrouter()

	cartRouter.Use(middleware.AuthMiddleware(jwtSecret))

	cartRouter.HandleFunc("/add", h.AddItemToCart).Methods(http.MethodPost)
	cartRouter.HandleFunc("", h.GetCart).Methods(http.MethodGet)
	cartRouter.HandleFunc("/remove/{offer_id}", h.RemoveItemFromCart).Methods(http.MethodDelete)
	cartRouter.HandleFunc("/clear", h.ClearCart).Methods(http.MethodDelete)
}
