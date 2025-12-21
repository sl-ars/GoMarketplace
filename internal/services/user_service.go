package services

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"go-app-marketplace/internal/redisdb"
	"go-app-marketplace/internal/usecases"
	"go-app-marketplace/pkg/domain"
	"go-app-marketplace/pkg/logger"
)

type UserService struct {
	usecase *usecases.UserUseCase
	logger  *logger.Logger
}

func NewUserService(u *usecases.UserUseCase, log *logger.Logger) *UserService {
	return &UserService{
		usecase: u,
		logger:  log,
	}
}

func (s *UserService) GetCurrentUser(ctx context.Context, userID int64) (*domain.User, error) {
	s.logger.WithFields(logrus.Fields{
		"userID":    userID,
		"operation": "get_current_user",
	}).Info("Getting current user")

	key := fmt.Sprintf("user:%d", userID)

	user, err := redisdb.CacheGetOrSet(ctx, key, 10*time.Minute, func() (*domain.User, error) {
		return s.usecase.GetUserByID(ctx, userID)
	})
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"userID":    userID,
			"operation": "get_current_user",
		}).Error("Failed to get user")
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetUserByID(ctx context.Context, userID int64) (*domain.User, error) {
	return s.usecase.GetUserByID(ctx, userID)
}
