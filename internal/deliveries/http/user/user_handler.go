package user

import (
	"net/http"

	"go-app-marketplace/internal/services"
	"go-app-marketplace/pkg/apperror"
	"go-app-marketplace/pkg/httpx"
	"go-app-marketplace/pkg/logger"
	"go-app-marketplace/pkg/reqresp"
)

// Handler handles user-related HTTP requests
type Handler struct {
	service *services.UserService
	logger  *logger.Logger
}

// NewHandler creates a new user Handler
func NewHandler(service *services.UserService, log *logger.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  log,
	}
}

// GetCurrentUser returns the currently authenticated user
// @Summary Get current user
// @Tags users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} reqresp.StandardResponse
// @Failure 401 {object} reqresp.StandardResponse
// @Failure 404 {object} reqresp.StandardResponse
// @Router /api/users/me [get]
func (h *Handler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, apperror.ErrUnauthorized, apperror.ErrMissingToken)
		return
	}

	user, err := h.service.GetCurrentUser(r.Context(), userID)
	if err != nil {
		h.logger.WithError(err).WithField("userID", userID).Warn("User not found")
		httpx.WriteError(w, http.StatusNotFound, apperror.ErrUserNotFound, apperror.ErrNotFound)
		return
	}

	response := reqresp.UserResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     string(user.Role),
	}

	httpx.WriteSuccess(w, http.StatusOK, "User retrieved successfully", response)
}
