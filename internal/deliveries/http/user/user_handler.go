package user

import (
	"encoding/json"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"go-app-marketplace/internal/services"
	"go-app-marketplace/pkg/httpx"
	"go-app-marketplace/pkg/logger"
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
// @Router /api/register [post]
func RegisterHandler(service *services.UserService, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.WithOperation("user_register").Info("Starting user registration")

		var req reqresp.RegisterUserRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.WithError(err).WithField("operation", "user_register").Error("Failed to decode registration request")
			httpx.WriteError(w, http.StatusBadRequest, "Invalid request", "Malformed JSON")
			return
		}

		if err := validate.Struct(&req); err != nil {
			log.WithError(err).WithFields(logrus.Fields{
				"email":      req.Email,
				"username":   req.Username,
				"operation":  "user_register",
			}).Warn("Validation failed for registration request")
			httpx.WriteError(w, http.StatusBadRequest, "Validation failed", err.Error())
			return
		}

		res, err := service.Register(r.Context(), &req)
		if err != nil {
			status := http.StatusInternalServerError
			if strings.Contains(err.Error(), "email already exists") || strings.Contains(err.Error(), "username already exists") {
				status = http.StatusConflict
				log.WithError(err).WithFields(logrus.Fields{
					"email":     req.Email,
					"username":  req.Username,
					"operation": "user_register",
				}).Warn("User registration failed - user already exists")
			} else {
				log.WithError(err).WithFields(logrus.Fields{
					"email":     req.Email,
					"username":  req.Username,
					"operation": "user_register",
				}).Error("User registration failed")
			}
			httpx.WriteError(w, status, "Registration failed", err.Error())
			return
		}

		log.WithFields(logrus.Fields{
			"userID":    res.ID,
			"email":     req.Email,
			"username":  req.Username,
			"operation": "user_register",
		}).Info("User registered successfully")

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
// @Router /api/login [post]
func LoginHandler(service *services.UserService, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.WithOperation("user_login").Info("Starting user login")

		var req reqresp.LoginUserRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.WithError(err).WithField("operation", "user_login").Error("Failed to decode login request")
			httpx.WriteError(w, http.StatusBadRequest, "Invalid request", "Malformed JSON")
			return
		}

		if err := validate.Struct(&req); err != nil {
			log.WithError(err).WithFields(logrus.Fields{
				"email":     req.Email,
				"operation": "user_login",
			}).Warn("Validation failed for login request")
			httpx.WriteError(w, http.StatusBadRequest, "Validation failed", err.Error())
			return
		}

		resp, err := service.Login(r.Context(), &req)
		if err != nil {
			log.WithError(err).WithFields(logrus.Fields{
				"email":     req.Email,
				"operation": "user_login",
			}).Warn("Login failed - invalid credentials")
			httpx.WriteError(w, http.StatusUnauthorized, "Login failed", err.Error())
			return
		}

		log.WithFields(logrus.Fields{
			"email":     req.Email,
			"operation": "user_login",
		}).Info("User logged in successfully")

		httpx.WriteSuccess(w, http.StatusOK, "Login successful", resp)
	}
}

// @Summary Refresh access token
// @Description Refreshes JWT access token using refresh_token cookie
// @Tags users
// @Produce json
// @Success 200 {object} reqresp.StandardResponse
// @Failure 401 {object} reqresp.StandardResponse
// @Router /api/refresh [post]
func RefreshHandler(service *services.UserService, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.WithOperation("token_refresh").Info("Starting token refresh")

		cookie, err := r.Cookie("refresh_token")
		if err != nil {
			log.WithError(err).WithField("operation", "token_refresh").Warn("Missing refresh token cookie")
			httpx.WriteError(w, http.StatusUnauthorized, "Unauthorized", "Missing refresh token cookie")
			return
		}

		token, err := service.Refresh(r.Context(), cookie.Value)
		if err != nil {
			log.WithError(err).WithField("operation", "token_refresh").Warn("Token refresh failed - invalid or expired refresh token")
			httpx.WriteError(w, http.StatusUnauthorized, "Unauthorized", "Invalid or expired refresh token")
			return
		}

		log.WithOperation("token_refresh").Info("Token refreshed successfully")
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
// @Router /api/verify [get]
func VerifyHandler(service *services.UserService, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.WithOperation("token_verify").Info("Starting token verification")

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			log.WithOperation("token_verify").Warn("Missing or invalid Authorization header")
			httpx.WriteError(w, http.StatusUnauthorized, "Unauthorized", "Missing or invalid Authorization header")
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		if err := service.Verify(tokenStr); err != nil {
			log.WithError(err).WithField("operation", "token_verify").Warn("Token verification failed - invalid or expired token")
			httpx.WriteError(w, http.StatusUnauthorized, "Unauthorized", "Invalid or expired token")
			return
		}

		log.WithOperation("token_verify").Info("Token verified successfully")
		httpx.WriteSuccess(w, http.StatusOK, "Token is valid", nil)
	}
}

// @Summary Get current user
// @Tags users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} reqresp.StandardResponse
// @Failure 401 {object} reqresp.StandardResponse
// @Failure 404 {object} reqresp.StandardResponse
// @Router /api/me [get]
func GetCurrentUserHandler(service *services.UserService, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value("user_id").(int64)
		if !ok {
			log.WithOperation("get_current_user").Warn("User ID not found in context")
			httpx.WriteError(w, http.StatusUnauthorized, "Unauthorized", "User ID not found in context")
			return
		}

		log.WithFields(logrus.Fields{
			"userID":    userID,
			"operation": "get_current_user",
		}).Info("Getting current user")

		user, err := service.GetCurrentUser(r.Context(), userID)
		if err != nil {
			if err.Error() == "user not found" {
				log.WithFields(logrus.Fields{
					"userID":    userID,
					"operation": "get_current_user",
				}).Warn("User not found")
				httpx.WriteError(w, http.StatusNotFound, "Not Found", "User not found")
				return
			}
			log.WithError(err).WithFields(logrus.Fields{
				"userID":    userID,
				"operation": "get_current_user",
			}).Error("Failed to get current user")
			httpx.WriteError(w, http.StatusInternalServerError, "Internal Server Error", err.Error())
			return
		}

		response := reqresp.UserResponse{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Role:     string(user.Role),
		}

		log.WithFields(logrus.Fields{
			"userID":    userID,
			"operation": "get_current_user",
		}).Info("Current user retrieved successfully")
		httpx.WriteSuccess(w, http.StatusOK, "User retrieved successfully", response)
	}
}
