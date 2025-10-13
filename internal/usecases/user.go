package usecases

import (
	"context"
	"database/sql"
	"errors"
	"go-app-marketplace/pkg/domain"
	"go-app-marketplace/pkg/hash"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *domain.User) (int64, error)
	IsEmailTaken(ctx context.Context, email string) (bool, error)
	IsUsernameTaken(ctx context.Context, username string) (bool, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id int64) (*domain.User, error)
	GetUserByID(ctx context.Context, userID int64) (*domain.User, error)
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

	taken, err = u.repo.IsUsernameTaken(ctx, user.Username)
	if err != nil {
		return 0, err
	}
	if taken {
		return 0, errors.New("username already exists")
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

func (u *UserUseCase) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	return u.repo.GetByID(ctx, id)
}

func (uc *UserUseCase) GetUserByID(ctx context.Context, userID int64) (*domain.User, error) {
	user, err := uc.repo.GetUserByID(ctx, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user, nil
}
