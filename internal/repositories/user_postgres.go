package repositories

import (
	"context"
	"github.com/jmoiron/sqlx"
	"go-app-marketplace/pkg/domain"
)

type UserPostgresRepo struct {
	db *sqlx.DB
}

func NewUserPostgresRepo(db *sqlx.DB) *UserPostgresRepo {
	return &UserPostgresRepo{db: db}
}

func (r *UserPostgresRepo) IsEmailTaken(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM users WHERE email=$1)`, email)
	return exists, err
}

func (r *UserPostgresRepo) CreateUser(ctx context.Context, user *domain.User) (int64, error) {
	var id int64
	err := r.db.GetContext(ctx, &id,
		`INSERT INTO users (username, email, password_hash, role, created_at)
         VALUES ($1, $2, $3, $4, NOW()) RETURNING id`,
		user.Username, user.Email, user.Password, user.Role)
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
