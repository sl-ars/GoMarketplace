package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"

	"go-app-marketplace/pkg/domain"
	"go-app-marketplace/pkg/logger"
)

type UserPostgresRepo struct {
	db     *sqlx.DB
	logger *logger.Logger
}

func NewUserPostgresRepo(db *sqlx.DB, log *logger.Logger) *UserPostgresRepo {
	return &UserPostgresRepo{db: db, logger: log}
}

func (r *UserPostgresRepo) IsEmailTaken(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM users WHERE email=$1)`, email)
	return exists, err
}

func (r *UserPostgresRepo) IsUsernameTaken(ctx context.Context, username string) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM users WHERE username=$1)`, username)
	return exists, err
}

func (r *UserPostgresRepo) CreateUser(ctx context.Context, user *domain.User) (int64, error) {
	r.logger.WithFields(logrus.Fields{
		"email":    user.Email,
		"username": user.Username,
		"role":     user.Role,
	}).WithField("operation", "create_user").Info("Creating user in database")

	var id int64
	err := r.db.GetContext(ctx, &id,
		`INSERT INTO users (username, email, password_hash, role, email_verified, created_at)
         VALUES ($1, $2, $3, $4, $5, NOW()) RETURNING id`,
		user.Username, user.Email, user.Password, user.Role, false)

	if err != nil {
		r.logger.WithError(err).WithFields(logrus.Fields{
			"email":    user.Email,
			"username": user.Username,
		}).WithField("operation", "create_user").Error("Failed to create user in database")
		return 0, err
	}

	r.logger.WithFields(logrus.Fields{
		"userID":   id,
		"email":    user.Email,
		"username": user.Username,
	}).WithField("operation", "create_user").Info("User created successfully in database")

	return id, err
}

func (r *UserPostgresRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.db.GetContext(ctx, &user, `SELECT * FROM users WHERE email = $1`, email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserPostgresRepo) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	var user domain.User
	err := r.db.GetContext(ctx, &user, `SELECT * FROM users WHERE id = $1`, id)
	return &user, err
}

func (r *UserPostgresRepo) GetUserByID(ctx context.Context, userID int64) (*domain.User, error) {
	var user domain.User
	err := r.db.GetContext(ctx, &user, `
		SELECT id, username, email, role, email_verified, created_at
		FROM users
		WHERE id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// SetVerificationToken sets the email verification token for a user
func (r *UserPostgresRepo) SetVerificationToken(ctx context.Context, userID int64, tokenHash, code string, expiresAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users 
		SET verification_token_hash = $1, 
		    verification_code = $2,
		    verification_token_expires_at = $3
		WHERE id = $4
	`, tokenHash, code, expiresAt, userID)
	return err
}

// VerifyEmail marks the user's email as verified and clears the verification token
func (r *UserPostgresRepo) VerifyEmail(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users 
		SET email_verified = true,
		    email_verified_at = NOW(),
		    verification_token_hash = NULL,
		    verification_token_expires_at = NULL,
		    verification_code = NULL
		WHERE id = $1
	`, userID)
	return err
}

// GetByVerificationToken finds a user by their verification token hash
func (r *UserPostgresRepo) GetByVerificationToken(ctx context.Context, tokenHash string) (*domain.User, error) {
	var user domain.User
	err := r.db.GetContext(ctx, &user, `
		SELECT * FROM users 
		WHERE verification_token_hash = $1 
		AND verification_token_expires_at > NOW()
	`, tokenHash)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByVerificationCode finds a user by their verification code (for email + code verification)
func (r *UserPostgresRepo) GetByVerificationCode(ctx context.Context, email, code string) (*domain.User, error) {
	var user domain.User
	err := r.db.GetContext(ctx, &user, `
		SELECT * FROM users 
		WHERE email = $1 
		AND verification_code = $2
		AND verification_token_expires_at > NOW()
	`, email, code)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// SetResetToken sets the password reset token for a user
func (r *UserPostgresRepo) SetResetToken(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users 
		SET reset_token_hash = $1, 
		    reset_token_expires_at = $2
		WHERE id = $3
	`, tokenHash, expiresAt, userID)
	return err
}

// GetByResetToken finds a user by their password reset token hash
func (r *UserPostgresRepo) GetByResetToken(ctx context.Context, tokenHash string) (*domain.User, error) {
	var user domain.User
	err := r.db.GetContext(ctx, &user, `
		SELECT * FROM users 
		WHERE reset_token_hash = $1 
		AND reset_token_expires_at > NOW()
	`, tokenHash)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdatePassword updates the user's password and clears the reset token
func (r *UserPostgresRepo) UpdatePassword(ctx context.Context, userID int64, passwordHash string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users 
		SET password_hash = $1,
		    reset_token_hash = NULL,
		    reset_token_expires_at = NULL
		WHERE id = $2
	`, passwordHash, userID)
	return err
}

// ClearResetToken clears the password reset token
func (r *UserPostgresRepo) ClearResetToken(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users 
		SET reset_token_hash = NULL,
		    reset_token_expires_at = NULL
		WHERE id = $1
	`, userID)
	return err
}

// GetByEmailForPasswordReset finds a user by email (for password reset flow)
func (r *UserPostgresRepo) GetByEmailForPasswordReset(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.db.GetContext(ctx, &user, `
		SELECT id, username, email, email_verified FROM users WHERE email = $1
	`, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Return nil, nil to not reveal if email exists
		}
		return nil, err
	}
	return &user, nil
}

// ============================================================================
// Admin User Management Methods
// ============================================================================

// isValidRoleFilter validates role filter against allowed values
// This prevents injection of arbitrary values into the query
func isValidRoleFilter(role string) bool {
	switch domain.UserRole(role) {
	case domain.UserRoleAdmin, domain.UserRoleSeller, domain.UserRoleCustomer:
		return true
	case "": // Empty string means no filter
		return true
	default:
		return false
	}
}

// ListUsers returns a paginated list of users with optional filters
// Security: Role filter is validated against whitelist to prevent injection
func (r *UserPostgresRepo) ListUsers(ctx context.Context, page, pageSize int, role string, isBanned *bool) ([]*domain.User, int64, error) {
	offset := (page - 1) * pageSize

	// Validate role against whitelist (CWE-20: Input Validation)
	if !isValidRoleFilter(role) {
		return nil, 0, fmt.Errorf("invalid role filter: %s", role)
	}

	// Build query dynamically based on filters
	query := `
		SELECT id, username, email, role, email_verified, created_at, is_banned, banned_at, ban_reason, banned_by_id
		FROM users
		WHERE 1=1
	`
	countQuery := `SELECT COUNT(*) FROM users WHERE 1=1`
	args := []interface{}{}
	argIndex := 1

	if role != "" {
		query += fmt.Sprintf(" AND role = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND role = $%d", argIndex)
		args = append(args, role)
		argIndex++
	}

	if isBanned != nil {
		query += fmt.Sprintf(" AND is_banned = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND is_banned = $%d", argIndex)
		args = append(args, *isBanned)
		argIndex++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, pageSize, offset)

	var users []*domain.User
	err := r.db.SelectContext(ctx, &users, query, args...)
	if err != nil {
		return nil, 0, err
	}

	// Get total count (remove limit/offset args)
	countArgs := args[:len(args)-2]
	var total int64
	err = r.db.GetContext(ctx, &total, countQuery, countArgs...)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// BanUser bans a user with a reason
func (r *UserPostgresRepo) BanUser(ctx context.Context, userID, bannedByID int64, reason string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users 
		SET is_banned = TRUE,
		    banned_at = NOW(),
		    ban_reason = $1,
		    banned_by_id = $2
		WHERE id = $3
	`, reason, bannedByID, userID)
	return err
}

// UnbanUser removes the ban from a user
func (r *UserPostgresRepo) UnbanUser(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users 
		SET is_banned = FALSE,
		    banned_at = NULL,
		    ban_reason = NULL,
		    banned_by_id = NULL
		WHERE id = $1
	`, userID)
	return err
}

// UpdateUserRole updates the user's role
func (r *UserPostgresRepo) UpdateUserRole(ctx context.Context, userID int64, role domain.UserRole) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users SET role = $1 WHERE id = $2
	`, role, userID)
	return err
}

// DeleteUser permanently deletes a user (use with caution)
func (r *UserPostgresRepo) DeleteUser(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, userID)
	return err
}

// GetUserStats returns statistics about users
func (r *UserPostgresRepo) GetUserStats(ctx context.Context) (map[string]int64, error) {
	stats := make(map[string]int64)

	// Total users
	var total int64
	err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM users")
	if err != nil {
		return nil, err
	}
	stats["total"] = total

	// By role
	rows, err := r.db.QueryContext(ctx, "SELECT role, COUNT(*) FROM users GROUP BY role")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var role string
		var count int64
		if err := rows.Scan(&role, &count); err != nil {
			return nil, err
		}
		stats["role_"+role] = count
	}

	// Banned users
	var banned int64
	err = r.db.GetContext(ctx, &banned, "SELECT COUNT(*) FROM users WHERE is_banned = TRUE")
	if err != nil {
		return nil, err
	}
	stats["banned"] = banned

	// Verified users
	var verified int64
	err = r.db.GetContext(ctx, &verified, "SELECT COUNT(*) FROM users WHERE email_verified = TRUE")
	if err != nil {
		return nil, err
	}
	stats["verified"] = verified

	return stats, nil
}
