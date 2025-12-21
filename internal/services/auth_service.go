package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"

	"go-app-marketplace/internal/messagebus"
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

// AuthServiceConfig holds auth service configuration
type AuthServiceConfig struct {
	JWTSecret       string
	BaseURL         string // For building verification/reset URLs
	RequireVerified bool   // Whether email verification is required for login
}

// AuthService handles authentication operations
type AuthService struct {
	usecase        *usecases.AuthUseCase
	emailPublisher *messagebus.EmailPublisher
	config         AuthServiceConfig
	jwtKey         []byte
	logger         *logger.Logger
}

// NewAuthService creates a new AuthService instance
func NewAuthService(uc *usecases.AuthUseCase, emailPublisher *messagebus.EmailPublisher, cfg AuthServiceConfig, log *logger.Logger) *AuthService {
	return &AuthService{
		usecase:        uc,
		emailPublisher: emailPublisher,
		config:         cfg,
		jwtKey:         []byte(cfg.JWTSecret),
		logger:         log,
	}
}

// Register creates a new user account, sends verification email, and returns registration response
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

	// Create verification token and queue email
	if s.emailPublisher != nil {
		verificationToken, err := s.usecase.CreateVerificationToken(ctx, user.ID)
		if err != nil {
			s.logger.WithError(err).WithField("userID", user.ID).Error("Failed to create verification token")
			// Don't fail registration, just log the error
		} else {
			// Publish to RabbitMQ queue
			verifyURL := fmt.Sprintf("%s/verify-email?token=%s", s.config.BaseURL, verificationToken.Token)
			if err := s.emailPublisher.PublishVerificationEmail(ctx, user.Email, user.Username, verifyURL, verificationToken.Code); err != nil {
				s.logger.WithError(err).WithField("email", user.Email).Error("Failed to queue verification email")
			} else {
				s.logger.WithField("email", user.Email).Info("Verification email queued")
			}
		}
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

	// Check if email verification is required
	if s.config.RequireVerified && !user.EmailVerified {
		s.logger.WithFields(logrus.Fields{
			"userID":    user.ID,
			"operation": "auth_login",
		}).Warn("Login failed - email not verified")
		return nil, usecases.ErrEmailNotVerified
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
		"userID":        user.ID,
		"emailVerified": user.EmailVerified,
		"operation":     "auth_login",
	}).Info("User login successful - tokens generated")

	return &reqresp.LoginResponse{
		AccessToken:   accessToken,
		RefreshToken:  refreshToken,
		EmailVerified: user.EmailVerified,
	}, nil
}

// VerifyEmail verifies a user's email using a token or code
func (s *AuthService) VerifyEmail(ctx context.Context, req *reqresp.VerifyEmailRequest) error {
	s.logger.WithOperation("auth_verify_email").Info("Processing email verification")

	var err error

	if req.Token != "" {
		// Verify using token (from URL link)
		_, err = s.usecase.VerifyEmailWithToken(ctx, req.Token)
	} else if req.Email != "" && req.Code != "" {
		// Verify using email + code
		_, err = s.usecase.VerifyEmailWithCode(ctx, req.Email, req.Code)
	} else {
		return ErrInvalidToken
	}

	if err != nil {
		s.logger.WithError(err).WithField("operation", "auth_verify_email").Warn("Email verification failed")
		return err
	}

	s.logger.WithOperation("auth_verify_email").Info("Email verified successfully")
	return nil
}

// ResendVerification resends the verification email
func (s *AuthService) ResendVerification(ctx context.Context, email string) error {
	s.logger.WithFields(logrus.Fields{
		"email":     email,
		"operation": "auth_resend_verification",
	}).Info("Processing resend verification request")

	user, verificationToken, err := s.usecase.ResendVerificationToken(ctx, email)
	if err != nil {
		s.logger.WithError(err).WithField("email", email).Warn("Resend verification failed")
		return err
	}

	if s.emailPublisher != nil {
		verifyURL := fmt.Sprintf("%s/verify-email?token=%s", s.config.BaseURL, verificationToken.Token)
		if err := s.emailPublisher.PublishVerificationEmail(ctx, user.Email, user.Username, verifyURL, verificationToken.Code); err != nil {
			s.logger.WithError(err).WithField("email", email).Error("Failed to queue verification email")
			return err
		}
	}

	s.logger.WithField("email", email).Info("Verification email resent successfully")
	return nil
}

// ForgotPassword initiates the password reset flow
func (s *AuthService) ForgotPassword(ctx context.Context, emailAddr string) error {
	s.logger.WithFields(logrus.Fields{
		"email":     emailAddr,
		"operation": "auth_forgot_password",
	}).Info("Processing forgot password request")

	user, resetToken, err := s.usecase.CreatePasswordResetToken(ctx, emailAddr)
	if err != nil {
		s.logger.WithError(err).WithField("email", emailAddr).Error("Failed to create reset token")
		return err
	}

	// If user doesn't exist, don't reveal it - just return success
	if user == nil || resetToken == nil {
		s.logger.WithField("email", emailAddr).Info("Password reset requested for non-existent email (silent)")
		return nil
	}

	// Queue password reset email
	if s.emailPublisher != nil {
		resetURL := fmt.Sprintf("%s/reset-password?token=%s", s.config.BaseURL, resetToken.Token)
		if err := s.emailPublisher.PublishPasswordResetEmail(ctx, user.Email, user.Username, resetURL); err != nil {
			s.logger.WithError(err).WithField("email", emailAddr).Error("Failed to queue password reset email")
			return err
		}
	}

	s.logger.WithField("email", emailAddr).Info("Password reset email queued")
	return nil
}

// ResetPassword resets the user's password using a token
func (s *AuthService) ResetPassword(ctx context.Context, req *reqresp.ResetPasswordRequest) error {
	s.logger.WithOperation("auth_reset_password").Info("Processing password reset")

	_, err := s.usecase.ResetPassword(ctx, req.Token, req.NewPassword)
	if err != nil {
		s.logger.WithError(err).WithField("operation", "auth_reset_password").Warn("Password reset failed")
		return err
	}

	s.logger.WithOperation("auth_reset_password").Info("Password reset successfully")
	return nil
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
