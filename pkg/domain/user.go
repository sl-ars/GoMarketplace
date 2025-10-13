package domain

import "time"

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
}

func IsValidRole(role UserRole) bool {
	switch role {
	case UserRoleAdmin, UserRoleSeller, UserRoleCustomer:
		return true
	default:
		return false
	}
}
