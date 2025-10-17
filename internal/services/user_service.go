package services

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"go-app-marketplace/internal/usecases"
	"go-app-marketplace/pkg/auth"
	"go-app-marketplace/pkg/domain"
	"go-app-marketplace/pkg/hash"
	"go-app-marketplace/pkg/logger"
	"go-app-marketplace/pkg/reqresp"
)

type UserService struct {
	usecase *usecases.UserUseCase
	jwtKey  []byte
	logger  *logger.Logger
}

func NewUserService(u *usecases.UserUseCase, jwtSecret string, log *logger.Logger) *UserService {
	return &UserService{
		usecase: u,
		jwtKey:  []byte(jwtSecret),
		logger:  log,
	}
}

func (s *UserService) Register(ctx context.Context, req *reqresp.RegisterUserRequest) (*reqresp.RegisterUserResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"email":     req.Email,
		"username":  req.Username,
		"operation": "user_register",
	}).Info("Processing user registration")

	hashed, err := hash.HashPassword(req.Password)
	if err != nil {
		s.logger.WithError(err).WithField("operation", "user_register").Error("Failed to hash password")
		return nil, err
	}

	user := &domain.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashed,
		Role:     "customer",
	}

	id, err := s.usecase.Register(ctx, user)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"email":     req.Email,
			"username":  req.Username,
			"operation": "user_register",
		}).Error("Failed to register user in database")
		return nil, err
	}

	s.logger.WithFields(logrus.Fields{
		"userID":    id,
		"email":     req.Email,
		"username":  req.Username,
		"operation": "user_register",
	}).Info("User registered successfully in database")

	return &reqresp.RegisterUserResponse{
		ID:       id,
		Username: req.Username,
		Email:    req.Email,
	}, nil
}

func (s *UserService) Login(ctx context.Context, req *reqresp.LoginUserRequest) (*reqresp.LoginUserResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"email":     req.Email,
		"operation": "user_login",
	}).Info("Processing user login")

	user, err := s.usecase.Login(ctx, req.Email, req.Password)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"email":     req.Email,
			"operation": "user_login",
		}).Warn("User login failed - invalid credentials")
		return nil, err
	}

	accessToken, err := auth.GenerateAccessToken(user.ID, string(user.Role), s.jwtKey)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"userID":    user.ID,
			"operation": "user_login",
		}).Error("Failed to generate access token")
		return nil, err
	}

	refreshToken, err := auth.GenerateRefreshToken(user.ID, s.jwtKey)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"userID":    user.ID,
			"operation": "user_login",
		}).Error("Failed to generate refresh token")
		return nil, err
	}

	s.logger.WithFields(logrus.Fields{
		"userID":    user.ID,
		"operation": "user_login",
	}).Info("User login successful - tokens generated")

	return &reqresp.LoginUserResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *UserService) Refresh(ctx context.Context, refreshToken string) (string, error) {
	token, err := auth.ParseToken(refreshToken, s.jwtKey)
	if err != nil || !token.Valid {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", err
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return "", err
	}
	userID := int64(userIDFloat)

	user, err := s.usecase.GetByID(ctx, userID)
	if err != nil {
		return "", err
	}

	return auth.GenerateAccessToken(user.ID, string(user.Role), s.jwtKey)
}

func (s *UserService) Verify(tokenStr string) error {
	token, err := auth.ParseToken(tokenStr, s.jwtKey)
	if err != nil || !token.Valid {
		return err
	}
	return nil
}

func (s *UserService) GetCurrentUser(ctx context.Context, userID int64) (*domain.User, error) {
	return s.usecase.GetUserByID(ctx, userID)
}
