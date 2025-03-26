package usecases

import (
	"context"
	"errors"
	"go-app-marketplace/pkg/domain"
	"go-app-marketplace/pkg/hash"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *domain.User) (int64, error)
	IsEmailTaken(ctx context.Context, email string) (bool, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
}

type UserUseCase struct {
	repo UserRepository
}

func NewUserUseCase(repo UserRepository) *UserUseCase {
	return &UserUseCase{repo: repo}
}

func (u *UserUseCase) Register(ctx context.Context, user *domain.User) (int64, error) {
	taken, err := u.repo.IsEmailTaken(ctx, user.Email)
	if err != nil {
		return 0, err
	}
	if taken {
		return 0, errors.New("email already exists")
	}

	return u.repo.CreateUser(ctx, user)
}

func (u *UserUseCase) Login(ctx context.Context, email, password string) (*domain.User, error) {
	user, err := u.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	match, err := hash.ComparePassword(user.Password, password)
	if err != nil || !match {
		return nil, errors.New("invalid credentials")
	}

	return user, nil
}
