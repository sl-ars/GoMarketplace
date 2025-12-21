package usecases

import (
	"context"
	"errors"
	"time"

	"go-app-marketplace/pkg/domain"
	"go-app-marketplace/pkg/hash"
	"go-app-marketplace/pkg/token"
)

var (
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrEmailTaken           = errors.New("email already exists")
	ErrUsernameTaken        = errors.New("username already exists")
	ErrUserNotFound         = errors.New("user not found")
	ErrEmailNotVerified     = errors.New("email not verified")
	ErrInvalidToken         = errors.New("invalid or expired token")
	ErrEmailAlreadyVerified = errors.New("email already verified")
)

// AuthRepository defines the interface for auth-related database operations
type AuthRepository interface {
	CreateUser(ctx context.Context, user *domain.User) (int64, error)
	IsEmailTaken(ctx context.Context, email string) (bool, error)
	IsUsernameTaken(ctx context.Context, username string) (bool, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id int64) (*domain.User, error)

	// Email verification
	SetVerificationToken(ctx context.Context, userID int64, tokenHash, code string, expiresAt time.Time) error
	VerifyEmail(ctx context.Context, userID int64) error
	GetByVerificationToken(ctx context.Context, tokenHash string) (*domain.User, error)
	GetByVerificationCode(ctx context.Context, email, code string) (*domain.User, error)

	// Password reset
	SetResetToken(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error
	GetByResetToken(ctx context.Context, tokenHash string) (*domain.User, error)
	UpdatePassword(ctx context.Context, userID int64, passwordHash string) error
	GetByEmailForPasswordReset(ctx context.Context, email string) (*domain.User, error)
}

// AuthUseCase handles authentication business logic
type AuthUseCase struct {
	repo AuthRepository
}

// NewAuthUseCase creates a new AuthUseCase instance
func NewAuthUseCase(repo AuthRepository) *AuthUseCase {
	return &AuthUseCase{repo: repo}
}

// VerificationToken holds verification data
type VerificationToken struct {
	Token     string
	Code      string
	ExpiresAt time.Time
}

// ResetToken holds password reset data
type ResetToken struct {
	Token     string
	ExpiresAt time.Time
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
		Username:      username,
		Email:         email,
		Password:      hashedPassword,
		Role:          domain.UserRoleCustomer,
		EmailVerified: false,
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

// CreateVerificationToken creates a new email verification token for a user
func (uc *AuthUseCase) CreateVerificationToken(ctx context.Context, userID int64) (*VerificationToken, error) {
	// Generate secure token
	tok, err := token.Generate(token.EmailVerificationTTL)
	if err != nil {
		return nil, err
	}

	// Generate 6-digit code
	code, err := token.GenerateCode(6)
	if err != nil {
		return nil, err
	}

	// Store hash in database
	if err := uc.repo.SetVerificationToken(ctx, userID, tok.Hash, code, tok.ExpiresAt); err != nil {
		return nil, err
	}

	return &VerificationToken{
		Token:     tok.PlainText,
		Code:      code,
		ExpiresAt: tok.ExpiresAt,
	}, nil
}

// VerifyEmailWithToken verifies a user's email using a token
func (uc *AuthUseCase) VerifyEmailWithToken(ctx context.Context, tokenPlainText string) (*domain.User, error) {
	tokenHash := token.HashToken(tokenPlainText)

	user, err := uc.repo.GetByVerificationToken(ctx, tokenHash)
	if err != nil {
		return nil, ErrInvalidToken
	}

	if user.EmailVerified {
		return nil, ErrEmailAlreadyVerified
	}

	if err := uc.repo.VerifyEmail(ctx, user.ID); err != nil {
		return nil, err
	}

	user.EmailVerified = true
	return user, nil
}

// VerifyEmailWithCode verifies a user's email using email + code
func (uc *AuthUseCase) VerifyEmailWithCode(ctx context.Context, email, code string) (*domain.User, error) {
	user, err := uc.repo.GetByVerificationCode(ctx, email, code)
	if err != nil {
		return nil, ErrInvalidToken
	}

	if user.EmailVerified {
		return nil, ErrEmailAlreadyVerified
	}

	if err := uc.repo.VerifyEmail(ctx, user.ID); err != nil {
		return nil, err
	}

	user.EmailVerified = true
	return user, nil
}

// ResendVerificationToken creates a new verification token (for resend)
func (uc *AuthUseCase) ResendVerificationToken(ctx context.Context, email string) (*domain.User, *VerificationToken, error) {
	user, err := uc.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, nil, ErrUserNotFound
	}

	if user.EmailVerified {
		return nil, nil, ErrEmailAlreadyVerified
	}

	verificationToken, err := uc.CreateVerificationToken(ctx, user.ID)
	if err != nil {
		return nil, nil, err
	}

	return user, verificationToken, nil
}

// CreatePasswordResetToken creates a password reset token for a user
func (uc *AuthUseCase) CreatePasswordResetToken(ctx context.Context, email string) (*domain.User, *ResetToken, error) {
	user, err := uc.repo.GetByEmailForPasswordReset(ctx, email)
	if err != nil {
		return nil, nil, err
	}
	if user == nil {
		// Return nil without error to not reveal if email exists
		return nil, nil, nil
	}

	// Generate secure token
	tok, err := token.Generate(token.PasswordResetTTL)
	if err != nil {
		return nil, nil, err
	}

	// Store hash in database
	if err := uc.repo.SetResetToken(ctx, user.ID, tok.Hash, tok.ExpiresAt); err != nil {
		return nil, nil, err
	}

	return user, &ResetToken{
		Token:     tok.PlainText,
		ExpiresAt: tok.ExpiresAt,
	}, nil
}

// ResetPassword resets a user's password using a token
func (uc *AuthUseCase) ResetPassword(ctx context.Context, tokenPlainText, newPassword string) (*domain.User, error) {
	tokenHash := token.HashToken(tokenPlainText)

	user, err := uc.repo.GetByResetToken(ctx, tokenHash)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// Hash new password
	hashedPassword, err := hash.HashPassword(newPassword)
	if err != nil {
		return nil, err
	}

	// Update password and clear reset token
	if err := uc.repo.UpdatePassword(ctx, user.ID, hashedPassword); err != nil {
		return nil, err
	}

	return user, nil
}
