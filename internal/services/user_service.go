package services

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
	"go-app-marketplace/internal/usecases"
	"go-app-marketplace/pkg/auth"
	"go-app-marketplace/pkg/domain"
	"go-app-marketplace/pkg/hash"
	"go-app-marketplace/pkg/reqresp"
)

type UserService struct {
	usecase *usecases.UserUseCase
	jwtKey  []byte
}

func NewUserService(u *usecases.UserUseCase, jwtSecret string) *UserService {
	return &UserService{
		usecase: u,
		jwtKey:  []byte(jwtSecret),
	}
}

func (s *UserService) Register(ctx context.Context, req *reqresp.RegisterUserRequest) (*reqresp.RegisterUserResponse, error) {
	hashed, err := hash.HashPassword(req.Password)
	if err != nil {
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
		return nil, err
	}

	return &reqresp.RegisterUserResponse{
		ID:       id,
		Username: req.Username,
		Email:    req.Email,
	}, nil
}

func (s *UserService) Login(ctx context.Context, req *reqresp.LoginUserRequest) (*reqresp.LoginUserResponse, error) {
	user, err := s.usecase.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	accessToken, err := auth.GenerateAccessToken(user.ID, s.jwtKey)
	if err != nil {
		return nil, err
	}

	refreshToken, err := auth.GenerateRefreshToken(user.ID, s.jwtKey)
	if err != nil {
		return nil, err
	}

	return &reqresp.LoginUserResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *UserService) Refresh(refreshToken string) (string, error) {
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

	newAccessToken, err := auth.GenerateAccessToken(userID, s.jwtKey)
	if err != nil {
		return "", err
	}

	return newAccessToken, nil
}

func (s *UserService) Verify(tokenStr string) error {
	token, err := auth.ParseToken(tokenStr, s.jwtKey)
	if err != nil || !token.Valid {
		return err
	}
	return nil
}
