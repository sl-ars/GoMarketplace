package user

import (
	"encoding/json"
	"github.com/go-playground/validator/v10"
	"go-app-marketplace/internal/services"
	"go-app-marketplace/pkg/httpx"
	"go-app-marketplace/pkg/reqresp"
	"net/http"
	"strings"
)

var validate = validator.New()

// @Summary Register new user
// @Description Registers a new user with username, email, and password
// @Tags users
// @Accept json
// @Produce json
// @Param input body reqresp.RegisterUserRequest true "User registration input"
// @Success 201 {object} reqresp.StandardResponse
// @Failure 400 {object} reqresp.StandardResponse
// @Failure 409 {object} reqresp.StandardResponse
// @Router /register [post]
func RegisterHandler(service *services.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req reqresp.RegisterUserRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "Invalid request", "Malformed JSON")
			return
		}

		if err := validate.Struct(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "Validation failed", err.Error())
			return
		}

		res, err := service.Register(r.Context(), &req)
		if err != nil {
			status := http.StatusInternalServerError
			if strings.Contains(err.Error(), "email already exists") {
				status = http.StatusConflict
			}
			httpx.WriteError(w, status, "Registration failed", err.Error())
			return
		}

		httpx.WriteSuccess(w, http.StatusCreated, "User registered successfully", res)
	}
}

// @Summary Login user
// @Description Authenticates user and returns JWT token
// @Tags users
// @Accept json
// @Produce json
// @Param input body reqresp.LoginUserRequest true "Login credentials"
// @Success 200 {object} reqresp.StandardResponse
// @Failure 400 {object} reqresp.StandardResponse
// @Failure 401 {object} reqresp.StandardResponse
// @Router /login [post]
func LoginHandler(service *services.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req reqresp.LoginUserRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "Invalid request", "Malformed JSON")
			return
		}

		if err := validate.Struct(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "Validation failed", err.Error())
			return
		}

		userResp, err := service.Login(r.Context(), &req)
		if err != nil {
			httpx.WriteError(w, http.StatusUnauthorized, "Login failed", err.Error())
			return
		}

		httpx.WriteSuccess(w, http.StatusOK, "Login successful", userResp)
	}
}

// @Summary Refresh access token
// @Description Refreshes JWT access token using refresh_token cookie
// @Tags users
// @Produce json
// @Success 200 {object} reqresp.StandardResponse
// @Failure 401 {object} reqresp.StandardResponse
// @Router /refresh [post]
func RefreshHandler(service *services.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("refresh_token")
		if err != nil {
			httpx.WriteError(w, http.StatusUnauthorized, "Unauthorized", "Missing refresh token cookie")
			return
		}

		token, err := service.Refresh(cookie.Value)
		if err != nil {
			httpx.WriteError(w, http.StatusUnauthorized, "Unauthorized", "Invalid or expired refresh token")
			return
		}

		httpx.WriteSuccess(w, http.StatusOK, "Token refreshed", map[string]string{
			"access_token": token,
		})
	}
}

// @Summary Verify JWT access token
// @Description Verifies the JWT token in Authorization header
// @Tags users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} reqresp.StandardResponse
// @Failure 401 {object} reqresp.StandardResponse
// @Router /verify [get]
func VerifyHandler(service *services.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			httpx.WriteError(w, http.StatusUnauthorized, "Unauthorized", "Missing or invalid Authorization header")
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		if err := service.Verify(tokenStr); err != nil {
			httpx.WriteError(w, http.StatusUnauthorized, "Unauthorized", "Invalid or expired token")
			return
		}

		httpx.WriteSuccess(w, http.StatusOK, "Token is valid", nil)
	}
}
