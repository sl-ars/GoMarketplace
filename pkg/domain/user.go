package domain

import "time"

type User struct {
	ID        int64     `db:"id"`
	Username  string    `db:"username"`
	Email     string    `db:"email"`
	Password  string    `db:"password_hash"`
	Role      string    `db:"role"`
	CreatedAt time.Time `db:"created_at"`
}
