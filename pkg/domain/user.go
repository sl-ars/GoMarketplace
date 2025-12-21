package domain

import (
	"database/sql"
	"time"
)

type UserRole string

const (
	UserRoleAdmin    UserRole = "admin"
	UserRoleSeller   UserRole = "seller"
	UserRoleCustomer UserRole = "customer"
)

type User struct {
	ID        int64     `db:"id"`
	Username  string    `db:"username"`
	Email     string    `db:"email"`
	Password  string    `db:"password_hash"`
	Role      UserRole  `db:"role"`
	CreatedAt time.Time `db:"created_at"`

	// Email verification fields
	EmailVerified              bool           `db:"email_verified"`
	EmailVerifiedAt            sql.NullTime   `db:"email_verified_at"`
	VerificationTokenHash      sql.NullString `db:"verification_token_hash"`
	VerificationTokenExpiresAt sql.NullTime   `db:"verification_token_expires_at"`
	VerificationCode           sql.NullString `db:"verification_code"`

	// Password reset fields
	ResetTokenHash      sql.NullString `db:"reset_token_hash"`
	ResetTokenExpiresAt sql.NullTime   `db:"reset_token_expires_at"`

	// Ban fields
	IsBanned   bool           `db:"is_banned"`
	BannedAt   sql.NullTime   `db:"banned_at"`
	BanReason  sql.NullString `db:"ban_reason"`
	BannedByID sql.NullInt64  `db:"banned_by_id"`
}

// IsBannedUser returns true if the user is banned
func (u *User) IsBannedUser() bool {
	return u.IsBanned
}

func IsValidRole(role UserRole) bool {
	switch role {
	case UserRoleAdmin, UserRoleSeller, UserRoleCustomer:
		return true
	default:
		return false
	}
}

// IsEmailVerified returns true if the user has verified their email
func (u *User) IsEmailVerified() bool {
	return u.EmailVerified
}

// HasValidVerificationToken checks if the user has a valid verification token
func (u *User) HasValidVerificationToken() bool {
	if !u.VerificationTokenHash.Valid || !u.VerificationTokenExpiresAt.Valid {
		return false
	}
	return time.Now().Before(u.VerificationTokenExpiresAt.Time)
}

// HasValidResetToken checks if the user has a valid password reset token
func (u *User) HasValidResetToken() bool {
	if !u.ResetTokenHash.Valid || !u.ResetTokenExpiresAt.Valid {
		return false
	}
	return time.Now().Before(u.ResetTokenExpiresAt.Time)
}
