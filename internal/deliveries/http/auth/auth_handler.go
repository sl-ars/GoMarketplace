package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"

	"go-app-marketplace/internal/services"
	"go-app-marketplace/internal/usecases"
	"go-app-marketplace/pkg/apperror"
	"go-app-marketplace/pkg/httpx"
	"go-app-marketplace/pkg/logger"
	"go-app-marketplace/pkg/reqresp"
)

var validate = validator.New()

// Handler handles authentication HTTP requests
type Handler struct {
	service *services.AuthService
	logger  *logger.Logger
}

// NewHandler creates a new auth Handler
func NewHandler(service *services.AuthService, log *logger.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  log,
	}
}

// Register handles user registration
// @Summary Register new user
// @Description Registers a new user with username, email, and password
// @Tags auth
// @Accept json
// @Produce json
// @Param input body reqresp.RegisterRequest true "User registration input"
// @Success 201 {object} reqresp.StandardResponse
// @Failure 400 {object} reqresp.StandardResponse
// @Failure 409 {object} reqresp.StandardResponse
// @Router /api/auth/register [post]
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req reqresp.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Failed to decode registration request")
		httpx.WriteError(w, http.StatusBadRequest, apperror.ErrBadRequest, apperror.ErrInvalidJSON)
		return
	}

	if err := validate.Struct(&req); err != nil {
		h.logger.WithError(err).Warn("Validation failed for registration request")
		httpx.WriteError(w, http.StatusBadRequest, apperror.ErrValidationFailed, apperror.ErrBadRequest)
		return
	}

	res, err := h.service.Register(r.Context(), &req)
	if err != nil {
		if errors.Is(err, usecases.ErrEmailTaken) {
			httpx.WriteError(w, http.StatusConflict, apperror.ErrConflict, apperror.ErrEmailTaken)
			return
		}
		if errors.Is(err, usecases.ErrUsernameTaken) {
			httpx.WriteError(w, http.StatusConflict, apperror.ErrConflict, apperror.ErrUsernameTaken)
			return
		}
		h.logger.WithError(err).Error("Registration failed")
		httpx.WriteError(w, http.StatusInternalServerError, apperror.ErrInternalServer, apperror.ErrInternalServer)
		return
	}

	h.logger.WithField("userID", res.ID).Info("User registered successfully")
	httpx.WriteSuccess(w, http.StatusCreated, "User registered successfully", res)
}

// Login handles user authentication
// @Summary Login user
// @Description Authenticates user and returns JWT tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param input body reqresp.LoginRequest true "Login credentials"
// @Success 200 {object} reqresp.StandardResponse
// @Failure 400 {object} reqresp.StandardResponse
// @Failure 401 {object} reqresp.StandardResponse
// @Router /api/auth/login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req reqresp.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Failed to decode login request")
		httpx.WriteError(w, http.StatusBadRequest, apperror.ErrBadRequest, apperror.ErrInvalidJSON)
		return
	}

	if err := validate.Struct(&req); err != nil {
		h.logger.WithError(err).Warn("Validation failed for login request")
		httpx.WriteError(w, http.StatusBadRequest, apperror.ErrValidationFailed, apperror.ErrBadRequest)
		return
	}

	resp, err := h.service.Login(r.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Warn("Login failed")
		httpx.WriteError(w, http.StatusUnauthorized, apperror.ErrUnauthorized, apperror.ErrInvalidCredentials)
		return
	}

	h.logger.Info("User logged in successfully")
	httpx.WriteSuccess(w, http.StatusOK, "Login successful", resp)
}

// Refresh handles token refresh
// @Summary Refresh access token
// @Description Refreshes JWT access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param input body reqresp.RefreshRequest true "Refresh token"
// @Success 200 {object} reqresp.StandardResponse
// @Failure 401 {object} reqresp.StandardResponse
// @Router /api/auth/refresh [post]
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var refreshToken string

	// Try to get from request body first
	var req reqresp.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err == nil && req.RefreshToken != "" {
		refreshToken = req.RefreshToken
	} else {
		// Fallback to cookie
		cookie, err := r.Cookie("refresh_token")
		if err != nil {
			h.logger.WithError(err).Warn("Missing refresh token")
			httpx.WriteError(w, http.StatusUnauthorized, apperror.ErrUnauthorized, apperror.ErrMissingToken)
			return
		}
		refreshToken = cookie.Value
	}

	resp, err := h.service.Refresh(r.Context(), refreshToken)
	if err != nil {
		h.logger.WithError(err).Warn("Token refresh failed")
		httpx.WriteError(w, http.StatusUnauthorized, apperror.ErrUnauthorized, apperror.ErrInvalidToken)
		return
	}

	h.logger.Info("Token refreshed successfully")
	httpx.WriteSuccess(w, http.StatusOK, "Token refreshed", resp)
}

// Verify handles token verification
// @Summary Verify JWT access token
// @Description Verifies the JWT token in Authorization header
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} reqresp.StandardResponse
// @Failure 401 {object} reqresp.StandardResponse
// @Router /api/auth/verify [get]
func (h *Handler) Verify(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		httpx.WriteError(w, http.StatusUnauthorized, apperror.ErrUnauthorized, apperror.ErrMissingToken)
		return
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

	resp, err := h.service.Verify(tokenStr)
	if err != nil {
		h.logger.WithError(err).Warn("Token verification failed")
		httpx.WriteError(w, http.StatusUnauthorized, apperror.ErrUnauthorized, apperror.ErrInvalidToken)
		return
	}

	h.logger.Info("Token verified successfully")
	httpx.WriteSuccess(w, http.StatusOK, "Token is valid", resp)
}
