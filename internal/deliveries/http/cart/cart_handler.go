package cart

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"

	"go-app-marketplace/internal/services"
	"go-app-marketplace/pkg/apperror"
	"go-app-marketplace/pkg/httpx"
	"go-app-marketplace/pkg/logger"
	"go-app-marketplace/pkg/reqresp"
)

type CartHandler struct {
	cartService *services.CartService
	logger      *logger.Logger
}

func NewCartHandler(cartService *services.CartService, log *logger.Logger) *CartHandler {
	return &CartHandler{
		cartService: cartService,
		logger:      log,
	}
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
		h.logger.WithError(err).Error("Failed to decode add to cart request")
		httpx.WriteError(w, http.StatusBadRequest, apperror.ErrBadRequest, apperror.ErrInvalidJSON)
		return
	}

	if err := validate.Struct(&req); err != nil {
		h.logger.WithError(err).Warn("Validation failed for add to cart request")
		httpx.WriteError(w, http.StatusBadRequest, apperror.ErrValidationFailed, apperror.ErrBadRequest)
		return
	}

	userID := r.Context().Value("user_id").(int64)

	if err := h.cartService.AddItem(r.Context(), userID, req.OfferID, req.Quantity); err != nil {
		h.logger.WithError(err).Error("Failed to add item to cart")
		httpx.WriteError(w, http.StatusInternalServerError, apperror.ErrAddToCartFailed, apperror.ErrInternalServer)
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
		h.logger.WithError(err).Error("Failed to retrieve cart")
		httpx.WriteError(w, http.StatusInternalServerError, apperror.ErrCartFetchFailed, apperror.ErrInternalServer)
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
		h.logger.WithError(err).Warn("Invalid offer ID format")
		httpx.WriteError(w, http.StatusBadRequest, apperror.ErrBadRequest, apperror.ErrInvalidID)
		return
	}

	userID := r.Context().Value("user_id").(int64)

	if err = h.cartService.RemoveItem(r.Context(), userID, offerID); err != nil {
		h.logger.WithError(err).Error("Failed to remove item from cart")
		httpx.WriteError(w, http.StatusInternalServerError, apperror.ErrRemoveFromCart, apperror.ErrInternalServer)
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

	if err := h.cartService.ClearCart(r.Context(), userID); err != nil {
		h.logger.WithError(err).Error("Failed to clear cart")
		httpx.WriteError(w, http.StatusInternalServerError, apperror.ErrClearCartFailed, apperror.ErrInternalServer)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, "Cart cleared successfully", nil)
}
