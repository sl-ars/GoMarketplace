package cart

import (
	"encoding/json"
	"go-app-marketplace/internal/services"
	"go-app-marketplace/pkg/httpx"
	"go-app-marketplace/pkg/reqresp"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type CartHandler struct {
	cartService *services.CartService
}

func NewCartHandler(cartService *services.CartService) *CartHandler {
	return &CartHandler{cartService: cartService}
}

var validate = validator.New()

// AddItemToCart adds an item to the user's cart
// @Summary Add item to cart
// @Tags Cart
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param input body reqresp.AddItemToCartRequest true "Offer ID and Quantity"
// @Success 200 {object} reqresp.StandardResponse
// @Failure 400 {object} reqresp.StandardResponse
// @Failure 500 {object} reqresp.StandardResponse
// @Router /api/cart/add [post]
func (h *CartHandler) AddItemToCart(w http.ResponseWriter, r *http.Request) {
	var req reqresp.AddItemToCartRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if err := validate.Struct(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	userID := r.Context().Value("user_id").(int64)

	err := h.cartService.AddItem(r.Context(), userID, req.OfferID, req.Quantity)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to add item to cart", err.Error())
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, "Item added to cart successfully", nil)
}

// GetCart returns all items in the user's cart
// @Summary Get cart items
// @Tags Cart
// @Security BearerAuth
// @Produce json
// @Success 200 {object} reqresp.StandardResponse
// @Failure 500 {object} reqresp.StandardResponse
// @Router /api/cart [get]
func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int64)

	cartItems, err := h.cartService.GetCart(r.Context(), userID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to retrieve cart", err.Error())
		return
	}

	var resp []reqresp.CartItemResponse
	for _, item := range cartItems {
		resp = append(resp, reqresp.CartItemResponse{
			OfferID:  item.OfferID,
			Quantity: item.Quantity,
		})
	}

	httpx.WriteSuccess(w, http.StatusOK, "Cart fetched successfully", resp)
}

// RemoveItemFromCart removes an item from the user's cart
// @Summary Remove item from cart
// @Tags Cart
// @Security BearerAuth
// @Produce json
// @Param offer_id path int true "Offer ID"
// @Success 200 {object} reqresp.StandardResponse
// @Failure 400 {object} reqresp.StandardResponse
// @Failure 500 {object} reqresp.StandardResponse
// @Router /api/cart/remove/{offer_id} [delete]
func (h *CartHandler) RemoveItemFromCart(w http.ResponseWriter, r *http.Request) {
	offerIDStr := mux.Vars(r)["offer_id"]
	offerID, err := strconv.ParseInt(offerIDStr, 10, 64)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "Invalid offer ID", err.Error())
		return
	}

	userID := r.Context().Value("user_id").(int64)

	err = h.cartService.RemoveItem(r.Context(), userID, offerID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to remove item from cart", err.Error())
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, "Item removed from cart successfully", nil)
}

// ClearCart clears all items from the cart
// @Summary Clear cart
// @Tags Cart
// @Security BearerAuth
// @Produce json
// @Success 200 {object} reqresp.StandardResponse
// @Failure 500 {object} reqresp.StandardResponse
// @Router /api/cart/clear [delete]
func (h *CartHandler) ClearCart(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int64)

	err := h.cartService.ClearCart(r.Context(), userID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to clear cart", err.Error())
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, "Cart cleared successfully", nil)
}
