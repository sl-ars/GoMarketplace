package usecases

import (
	"context"
	"errors"

	"go-app-marketplace/pkg/domain"
	"go-app-marketplace/pkg/hash"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailTaken         = errors.New("email already exists")
	ErrUsernameTaken      = errors.New("username already exists")
	ErrUserNotFound       = errors.New("user not found")
)

// AuthRepository defines the interface for auth-related database operations
type AuthRepository interface {
	CreateUser(ctx context.Context, user *domain.User) (int64, error)
	IsEmailTaken(ctx context.Context, email string) (bool, error)
	IsUsernameTaken(ctx context.Context, username string) (bool, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id int64) (*domain.User, error)
}

// AuthUseCase handles authentication business logic
type AuthUseCase struct {
	repo AuthRepository
}

// NewAuthUseCase creates a new AuthUseCase instance
func NewAuthUseCase(repo AuthRepository) *AuthUseCase {
	return &AuthUseCase{repo: repo}
}

// Register creates a new user account
func (uc *AuthUseCase) Register(ctx context.Context, username, email, password string) (*domain.User, error) {
	// Check if email is taken
	taken, err := uc.repo.IsEmailTaken(ctx, email)
	if err != nil {
		return nil, err
	}
	if taken {
		return nil, ErrEmailTaken
	}

	// Check if username is taken
	taken, err = uc.repo.IsUsernameTaken(ctx, username)
	if err != nil {
		return nil, err
	}
	if taken {
		return nil, ErrUsernameTaken
	}

	// Hash password
	hashedPassword, err := hash.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Username: username,
		Email:    email,
		Password: hashedPassword,
		Role:     domain.UserRoleCustomer,
	}

	id, err := uc.repo.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	user.ID = id
	return user, nil
}

// Login authenticates user by email and password
func (uc *AuthUseCase) Login(ctx context.Context, email, password string) (*domain.User, error) {
	user, err := uc.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	match, err := hash.ComparePassword(user.Password, password)
	if err != nil || !match {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

// GetUserByID retrieves user by ID (for token refresh)
func (uc *AuthUseCase) GetUserByID(ctx context.Context, userID int64) (*domain.User, error) {
	user, err := uc.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}
