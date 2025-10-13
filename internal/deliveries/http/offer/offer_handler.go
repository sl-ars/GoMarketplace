package offer

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"go-app-marketplace/internal/services"
	"go-app-marketplace/pkg/httpx"
	"go-app-marketplace/pkg/reqresp"
	"net/http"
	"strconv"
	"time"
)

type OfferHandler struct {
	offerService *services.OfferService
}

func NewOfferHandler(offerService *services.OfferService) *OfferHandler {
	return &OfferHandler{offerService: offerService}
}

// @Summary Create offer
// @Description Create a new offer for a product
// @Tags offers
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param input body reqresp.OfferCreateRequest true "Offer data"
// @Success 201 {object} reqresp.StandardResponse{data=reqresp.OfferCreateResponse}
// @Failure 400 {object} reqresp.StandardResponse "Invalid request"
// @Failure 401 {object} reqresp.StandardResponse "Unauthorized"
// @Failure 500 {object} reqresp.StandardResponse "Server error"
// @Router /api/offers [post]
func (h *OfferHandler) CreateOffer(w http.ResponseWriter, r *http.Request) {
	var req reqresp.OfferCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	sellerID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "Unauthorized", "Invalid user context")
		return
	}

	id, err := h.offerService.CreateOffer(r.Context(), req.ProductID, sellerID, req.Price, req.Stock, req.IsAvailable)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to create offer", err.Error())
		return
	}

	httpx.WriteSuccess(w, http.StatusCreated, "Offer created successfully", reqresp.OfferCreateResponse{ID: id})
}

// @Summary Get offer by ID
// @Description Retrieve a specific offer by its ID
// @Tags offers
// @Security BearerAuth
// @Produce json
// @Param id path int true "Offer ID"
// @Success 200 {object} reqresp.StandardResponse{data=reqresp.OfferResponse}
// @Failure 400 {object} reqresp.StandardResponse "Invalid offer ID"
// @Failure 401 {object} reqresp.StandardResponse "Unauthorized"
// @Failure 404 {object} reqresp.StandardResponse "Offer not found"
// @Failure 500 {object} reqresp.StandardResponse "Server error"
// @Router /api/offers/{id} [get]
func (h *OfferHandler) GetOffer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "Invalid offer ID", err.Error())
		return
	}

	offer, err := h.offerService.GetOfferByID(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, http.StatusNotFound, "Offer not found", err.Error())
		return
	}

	response := reqresp.OfferResponse{
		ID:          offer.ID,
		ProductID:   offer.ProductID,
		SellerID:    offer.SellerID,
		Price:       offer.Price,
		Stock:       offer.Stock,
		IsAvailable: offer.IsAvailable,
		CreatedAt:   offer.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   offer.UpdatedAt.Format(time.RFC3339),
	}

	httpx.WriteSuccess(w, http.StatusOK, "Offer retrieved successfully", response)
}

// @Summary Update offer
// @Description Update an existing offer
// @Tags offers
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Offer ID"
// @Param input body reqresp.OfferUpdateRequest true "Updated offer data"
// @Success 200 {object} reqresp.StandardResponse "Offer updated successfully"
// @Failure 400 {object} reqresp.StandardResponse "Invalid request"
// @Failure 401 {object} reqresp.StandardResponse "Unauthorized"
// @Failure 403 {object} reqresp.StandardResponse "Forbidden - not the offer owner"
// @Failure 404 {object} reqresp.StandardResponse "Offer not found"
// @Failure 500 {object} reqresp.StandardResponse "Server error"
// @Router /api/offers/{id} [put]
func (h *OfferHandler) UpdateOffer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "Invalid offer ID", err.Error())
		return
	}

	var req reqresp.OfferUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	sellerID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "Unauthorized", "Invalid user context")
		return
	}

	// First verify this offer belongs to the seller
	offer, err := h.offerService.GetOfferByID(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, http.StatusNotFound, "Offer not found", err.Error())
		return
	}

	if offer.SellerID != sellerID {
		httpx.WriteError(w, http.StatusForbidden, "Forbidden", "You do not have permission to update this offer")
		return
	}

	err = h.offerService.UpdateOffer(r.Context(), id, sellerID, req.Price, req.Stock, req.IsAvailable)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to update offer", err.Error())
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, "Offer updated successfully", nil)
}

// @Summary Delete offer
// @Description Delete an existing offer
// @Tags offers
// @Security BearerAuth
// @Produce json
// @Param id path int true "Offer ID"
// @Success 200 {object} reqresp.StandardResponse "Offer deleted successfully"
// @Failure 400 {object} reqresp.StandardResponse "Invalid offer ID"
// @Failure 401 {object} reqresp.StandardResponse "Unauthorized"
// @Failure 403 {object} reqresp.StandardResponse "Forbidden - not the offer owner"
// @Failure 404 {object} reqresp.StandardResponse "Offer not found"
// @Failure 500 {object} reqresp.StandardResponse "Server error"
// @Router /api/offers/{id} [delete]
func (h *OfferHandler) DeleteOffer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "Invalid offer ID", err.Error())
		return
	}

	sellerID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "Unauthorized", "Invalid user context")
		return
	}

	// First verify this offer belongs to the seller
	offer, err := h.offerService.GetOfferByID(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, http.StatusNotFound, "Offer not found", err.Error())
		return
	}

	if offer.SellerID != sellerID {
		httpx.WriteError(w, http.StatusForbidden, "Forbidden", "You do not have permission to delete this offer")
		return
	}

	err = h.offerService.DeleteOffer(r.Context(), id, sellerID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to delete offer", err.Error())
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, "Offer deleted successfully", nil)
}

// @Summary List my offers
// @Description Get a list of all offers created by the current seller
// @Tags offers
// @Security BearerAuth
// @Produce json
// @Success 200 {object} reqresp.StandardResponse{data=[]reqresp.OfferResponse}
// @Failure 401 {object} reqresp.StandardResponse "Unauthorized"
// @Failure 500 {object} reqresp.StandardResponse "Server error"
// @Router /api/offers/me [get]
func (h *OfferHandler) ListMyOffers(w http.ResponseWriter, r *http.Request) {
	sellerID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "Unauthorized", "Invalid user context")
		return
	}

	offers, err := h.offerService.ListOffersBySeller(r.Context(), sellerID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to fetch offers", err.Error())
		return
	}

	response := make([]reqresp.OfferResponse, 0, len(offers))
	for _, offer := range offers {
		response = append(response, reqresp.OfferResponse{
			ID:          offer.ID,
			ProductID:   offer.ProductID,
			SellerID:    offer.SellerID,
			Price:       offer.Price,
			Stock:       offer.Stock,
			IsAvailable: offer.IsAvailable,
			CreatedAt:   offer.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   offer.UpdatedAt.Format(time.RFC3339),
		})
	}

	httpx.WriteSuccess(w, http.StatusOK, "Offers retrieved successfully", response)
}
