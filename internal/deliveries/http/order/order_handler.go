package order

import (
	"encoding/json"
	"go-app-marketplace/internal/services"
	"go-app-marketplace/pkg/domain"
	"go-app-marketplace/pkg/httpx"
	"go-app-marketplace/pkg/reqresp"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type OrderHandler struct {
	orderService *services.OrderService
}

func NewOrderHandler(orderService *services.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

// @Summary Checkout cart
// @Description Create a new order from cart items
// @Tags orders
// @Security BearerAuth
// @Produce json
// @Success 200 {object} reqresp.CheckoutResponse
// @Failure 400 {object} reqresp.StandardResponse
// @Failure 401 {object} reqresp.StandardResponse
// @Failure 500 {object} reqresp.StandardResponse
// @Router /api/orders/checkout [post]
func (h *OrderHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int64)

	resp, err := h.orderService.Checkout(r.Context(), userID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to checkout", err.Error())
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, "Checkout successful", resp)
}

// @Summary Cancel order item
// @Description Cancel a specific order item
// @Tags orders
// @Security BearerAuth
// @Param id path int true "Order item ID"
// @Produce json
// @Success 200 {object} reqresp.StandardResponse
// @Failure 400 {object} reqresp.StandardResponse
// @Failure 401 {object} reqresp.StandardResponse
// @Failure 404 {object} reqresp.StandardResponse
// @Failure 500 {object} reqresp.StandardResponse
// @Router /api/orders/{id}/cancel [post]
func (h *OrderHandler) CancelOrderItem(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int64)

	idStr := mux.Vars(r)["id"]
	itemID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	err = h.orderService.CancelOrderItem(r.Context(), userID, itemID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to cancel order item", err.Error())
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, "Order item canceled successfully", nil)
}

// @Summary List user orders
// @Description Get a list of all user's orders
// @Tags orders
// @Security BearerAuth
// @Produce json
// @Success 200 {array} reqresp.OrderResponse
// @Failure 401 {object} reqresp.StandardResponse
// @Failure 500 {object} reqresp.StandardResponse
// @Router /api/orders [get]
func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int64)

	orders, err := h.orderService.ListOrders(r.Context(), userID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to list orders", err.Error())
		return
	}

	var resp []reqresp.OrderResponse
	for _, o := range orders {
		var items []reqresp.OrderItemResponse
		for _, item := range o.Items {
			items = append(items, reqresp.OrderItemResponse{
				ID:        item.ID,
				OfferID:   item.OfferID,
				ProductID: item.ProductID,
				SellerID:  item.SellerID,
				Quantity:  item.Quantity,
				UnitPrice: item.UnitPrice,
				Status:    item.Status,
			})
		}

		resp = append(resp, reqresp.OrderResponse{
			ID:            o.ID,
			UserID:        o.UserID,
			TotalAmount:   o.TotalAmount,
			Status:        string(o.Status),
			PaymentStatus: string(o.PaymentStatus),
			Items:         items,
		})
	}

	httpx.WriteSuccess(w, http.StatusOK, "Orders fetched successfully", resp)
}

// @Summary Get order details
// @Description Get detailed information about a specific order
// @Tags orders
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Produce json
// @Success 200 {object} reqresp.OrderResponse
// @Failure 400 {object} reqresp.StandardResponse
// @Failure 401 {object} reqresp.StandardResponse
// @Failure 404 {object} reqresp.StandardResponse
// @Failure 500 {object} reqresp.StandardResponse
// @Router /api/orders/{id} [get]
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int64)

	idStr := mux.Vars(r)["id"]
	orderID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "Invalid order ID", err.Error())
		return
	}

	order, err := h.orderService.GetOrderByID(r.Context(), userID, orderID)
	if err != nil {
		httpx.WriteError(w, http.StatusNotFound, "Order not found or access denied", err.Error())
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, "Order details retrieved successfully", order)
}

// @Summary Checkout existing order
// @Description Create new payment session for an existing order
// @Tags orders
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Produce json
// @Success 200 {object} reqresp.CheckoutResponse
// @Failure 400 {object} reqresp.StandardResponse
// @Failure 401 {object} reqresp.StandardResponse
// @Failure 404 {object} reqresp.StandardResponse
// @Failure 500 {object} reqresp.StandardResponse
// @Router /api/orders/checkout/{id} [post]
func (h *OrderHandler) CheckoutExistingOrder(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int64)

	idStr := mux.Vars(r)["id"]
	orderID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "Invalid order ID", err.Error())
		return
	}

	resp, err := h.orderService.CheckoutExistingOrder(r.Context(), userID, orderID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to create checkout session", err.Error())
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, "Checkout session created successfully", resp)
}

// PATCH /api/orders/items/{id}/status
// @Summary Seller updates order-item status
// @Description Seller sets status to 'processing' or 'delivered'
// @Tags orders
// @Security BearerAuth
// @Param id path int true "Order item ID"
// @Accept json
// @Produce json
// @Param input body reqresp.UpdateOrderItemStatusRequest true "New status"
// @Success 200 {object} reqresp.StandardResponse
// @Failure 400,401,403,500 {object} reqresp.StandardResponse
// @Router /api/seller/orders/items/{id}/status [patch]
func (h *OrderHandler) UpdateOrderItemStatus(w http.ResponseWriter, r *http.Request) {
	sellerID := r.Context().Value("user_id").(int64)
	idStr := mux.Vars(r)["id"]
	itemID, _ := strconv.ParseInt(idStr, 10, 64)

	var req reqresp.UpdateOrderItemStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "Invalid body", err.Error())
		return
	}
	if req.Status != "processing" && req.Status != "delivered" {
		httpx.WriteError(w, http.StatusBadRequest, "Invalid status", "allowed: processing, delivered")
		return
	}

	err := h.orderService.SellerUpdateOrderItemStatus(
		r.Context(),
		sellerID,
		itemID,
		domain.OrderItemStatus(req.Status),
	)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "Update failed", err.Error())
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, "Status updated", nil)
}

// @Summary   Seller: list own order-items
// @Tags      orders
// @Security  BearerAuth
// @Produce   json
// @Success   200 {object} reqresp.StandardResponse
// @Failure   401,500 {object} reqresp.StandardResponse
// @Router    /api/seller/orders [get]
func (h *OrderHandler) ListSellerOrderItems(w http.ResponseWriter, r *http.Request) {
	sellerID := r.Context().Value("user_id").(int64)

	items, err := h.orderService.ListSellerOrderItems(r.Context(), sellerID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to fetch seller orders", err.Error())
		return
	}
	httpx.WriteSuccess(w, http.StatusOK, "Orders fetched successfully", items)
}
