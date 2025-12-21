package refund

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"go-app-marketplace/internal/services"
	"go-app-marketplace/pkg/apperror"
	"go-app-marketplace/pkg/httpx"
	"go-app-marketplace/pkg/logger"
	"go-app-marketplace/pkg/reqresp"
)

type Handler struct {
	service *services.RefundService
	logger  *logger.Logger
}

func NewHandler(s *services.RefundService, log *logger.Logger) *Handler {
	return &Handler{
		service: s,
		logger:  log,
	}
}

// @Summary   Request a refund for an order-item
// @Tags      refunds
// @Security  BearerAuth
// @Accept    json
// @Produce   json
// @Param     item_id  path   int                         true  "Order-item ID"
// @Param     input    body   reqresp.RefundRequestBody   true  "Refund reason"
// @Success   201      {object} reqresp.StandardResponse
// @Failure   400      {object} reqresp.StandardResponse
// @Failure   401      {object} reqresp.StandardResponse
// @Router    /api/refunds/{item_id} [post]
func (h *Handler) Request(w http.ResponseWriter, r *http.Request) {
	itemIDStr := mux.Vars(r)["item_id"]
	itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		h.logger.WithError(err).Warn("Invalid item ID format")
		httpx.WriteError(w, http.StatusBadRequest, apperror.ErrBadRequest, apperror.ErrInvalidID)
		return
	}

	var body reqresp.RefundRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.logger.WithError(err).Error("Failed to decode refund request")
		httpx.WriteError(w, http.StatusBadRequest, apperror.ErrBadRequest, apperror.ErrInvalidJSON)
		return
	}

	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, apperror.ErrUnauthorized, apperror.ErrMissingToken)
		return
	}

	id, err := h.service.Request(r.Context(), userID, itemID, body.Reason)
	if err != nil {
		h.logger.WithError(err).Error("Refund request failed")
		httpx.WriteError(w, http.StatusBadRequest, apperror.ErrRefundCreateFailed, apperror.ErrRefundNotAllowed)
		return
	}

	httpx.WriteSuccess(w, http.StatusCreated, "Refund requested successfully",
		map[string]int64{"id": id})
}

// ============================================================================
// Seller side – approve / reject
// ============================================================================

// @Summary   Approve or reject a refund
// @Tags      refunds
// @Security  BearerAuth
// @Accept    json
// @Produce   json
// @Param     refund_id  path   int    true  "Refund ID"
// @Param     action     query  string true  "Action: approve | reject"
// @Success   200        {object} reqresp.StandardResponse
// @Failure   400        {object} reqresp.StandardResponse
// @Failure   401        {object} reqresp.StandardResponse
// @Router    /api/refunds/{refund_id}/decide [patch]
func (h *Handler) Decide(w http.ResponseWriter, r *http.Request) {
	refundIDStr := mux.Vars(r)["refund_id"]
	refundID, err := strconv.ParseInt(refundIDStr, 10, 64)
	if err != nil {
		h.logger.WithError(err).Warn("Invalid refund ID format")
		httpx.WriteError(w, http.StatusBadRequest, apperror.ErrBadRequest, apperror.ErrInvalidID)
		return
	}

	action := r.URL.Query().Get("action")
	if action != "approve" && action != "reject" {
		httpx.WriteError(w, http.StatusBadRequest, apperror.ErrBadRequest, "Action must be 'approve' or 'reject'")
		return
	}
	approve := action == "approve"

	sellerID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, apperror.ErrUnauthorized, apperror.ErrMissingToken)
		return
	}

	if err := h.service.Approve(r.Context(), sellerID, refundID, approve); err != nil {
		h.logger.WithError(err).Error("Failed to update refund status")
		httpx.WriteError(w, http.StatusBadRequest, apperror.ErrRefundUpdateFailed, apperror.ErrInternalServer)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, "Refund status updated successfully", nil)
}
