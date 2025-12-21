package services

import (
	"context"
	"errors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"

	"go-app-marketplace/internal/usecases"
	"go-app-marketplace/pkg/auth"
	"go-app-marketplace/pkg/logger"
	"go-app-marketplace/pkg/reqresp"
)

var (
	ErrInvalidToken  = errors.New("invalid or expired token")
	ErrMissingToken  = errors.New("missing token")
	ErrInvalidClaims = errors.New("invalid token claims")
)

// AuthService handles authentication operations
type AuthService struct {
	usecase *usecases.AuthUseCase
	jwtKey  []byte
	logger  *logger.Logger
}

// NewAuthService creates a new AuthService instance
func NewAuthService(uc *usecases.AuthUseCase, jwtSecret string, log *logger.Logger) *AuthService {
	return &AuthService{
		usecase: uc,
		jwtKey:  []byte(jwtSecret),
		logger:  log,
	}
}

// Register creates a new user account and returns registration response
func (s *AuthService) Register(ctx context.Context, req *reqresp.RegisterRequest) (*reqresp.RegisterResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"email":     req.Email,
		"username":  req.Username,
		"operation": "auth_register",
	}).Info("Processing user registration")

	user, err := s.usecase.Register(ctx, req.Username, req.Email, req.Password)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"email":     req.Email,
			"username":  req.Username,
			"operation": "auth_register",
		}).Error("Failed to register user")
		return nil, err
	}

	s.logger.WithFields(logrus.Fields{
		"userID":    user.ID,
		"email":     req.Email,
		"username":  req.Username,
		"operation": "auth_register",
	}).Info("User registered successfully")

	return &reqresp.RegisterResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
	}, nil
}

// Login authenticates user and returns access and refresh tokens
func (s *AuthService) Login(ctx context.Context, req *reqresp.LoginRequest) (*reqresp.LoginResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"email":     req.Email,
		"operation": "auth_login",
	}).Info("Processing user login")

	user, err := s.usecase.Login(ctx, req.Email, req.Password)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"email":     req.Email,
			"operation": "auth_login",
		}).Warn("Login failed - invalid credentials")
		return nil, err
	}

	accessToken, err := auth.GenerateAccessToken(user.ID, string(user.Role), s.jwtKey)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"userID":    user.ID,
			"operation": "auth_login",
		}).Error("Failed to generate access token")
		return nil, err
	}

	refreshToken, err := auth.GenerateRefreshToken(user.ID, s.jwtKey)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"userID":    user.ID,
			"operation": "auth_login",
		}).Error("Failed to generate refresh token")
		return nil, err
	}

	s.logger.WithFields(logrus.Fields{
		"userID":    user.ID,
		"operation": "auth_login",
	}).Info("User login successful - tokens generated")

	return &reqresp.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// Refresh generates a new access token using a valid refresh token
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*reqresp.RefreshResponse, error) {
	s.logger.WithOperation("auth_refresh").Info("Processing token refresh")

	if refreshToken == "" {
		return nil, ErrMissingToken
	}

	token, err := auth.ParseToken(refreshToken, s.jwtKey)
	if err != nil || !token.Valid {
		s.logger.WithError(err).WithField("operation", "auth_refresh").Warn("Invalid refresh token")
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidClaims
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return nil, ErrInvalidClaims
	}
	userID := int64(userIDFloat)

	user, err := s.usecase.GetUserByID(ctx, userID)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"userID":    userID,
			"operation": "auth_refresh",
		}).Warn("User not found for refresh token")
		return nil, err
	}

	accessToken, err := auth.GenerateAccessToken(user.ID, string(user.Role), s.jwtKey)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"userID":    userID,
			"operation": "auth_refresh",
		}).Error("Failed to generate new access token")
		return nil, err
	}

	s.logger.WithFields(logrus.Fields{
		"userID":    userID,
		"operation": "auth_refresh",
	}).Info("Token refreshed successfully")

	return &reqresp.RefreshResponse{
		AccessToken: accessToken,
	}, nil
}

// Verify validates an access token
func (s *AuthService) Verify(tokenStr string) (*reqresp.VerifyResponse, error) {
	s.logger.WithOperation("auth_verify").Info("Processing token verification")

	if tokenStr == "" {
		return nil, ErrMissingToken
	}

	token, err := auth.ParseToken(tokenStr, s.jwtKey)
	if err != nil || !token.Valid {
		s.logger.WithError(err).WithField("operation", "auth_verify").Warn("Token verification failed")
		return &reqresp.VerifyResponse{Valid: false}, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return &reqresp.VerifyResponse{Valid: false}, ErrInvalidClaims
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return &reqresp.VerifyResponse{Valid: false}, ErrInvalidClaims
	}

	s.logger.WithOperation("auth_verify").Info("Token verified successfully")

	return &reqresp.VerifyResponse{
		Valid:  true,
		UserID: int64(userIDFloat),
	}, nil
}

// GetJWTKey returns the JWT secret key (for middleware use)
func (s *AuthService) GetJWTKey() []byte {
	return s.jwtKey
}
