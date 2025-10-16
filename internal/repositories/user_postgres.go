package repositories

import (
	"context"
	"github.com/jmoiron/sqlx"
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
	r.logger.WithFields(logger.Fields{
		"email":    user.Email,
		"username": user.Username,
		"role":     user.Role,
	}).WithOperation("create_user").Info("Creating user in database")

	var id int64
	err := r.db.GetContext(ctx, &id,
		`INSERT INTO users (username, email, password_hash, role, created_at)
         VALUES ($1, $2, $3, $4, NOW()) RETURNING id`,
		user.Username, user.Email, user.Password, user.Role)
	
	if err != nil {
		r.logger.WithError(err).WithFields(logger.Fields{
			"email":    user.Email,
			"username": user.Username,
		}).WithOperation("create_user").Error("Failed to create user in database")
		return 0, err
	}

	r.logger.WithFields(logger.Fields{
		"userID":   id,
		"email":    user.Email,
		"username": user.Username,
	}).WithOperation("create_user").Info("User created successfully in database")

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
	err := r.db.GetContext(ctx, &user, `SELECT id, username, email, password_hash, role, created_at FROM users WHERE id = $1`, id)
	return &user, err
}

func (r *UserPostgresRepo) GetUserByID(ctx context.Context, userID int64) (*domain.User, error) {
	var user domain.User
	err := r.db.GetContext(ctx, &user, `
		SELECT id, username, email, role, created_at
		FROM users
		WHERE id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
