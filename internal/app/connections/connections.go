package connections

import (
	"fmt"

	"go-app-marketplace/internal/app/config"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

type Connections struct {
	DB    *sqlx.DB
	Redis *redis.Client
}

func NewConnections(cfg *config.Config) (*Connections, error) {
	db, err := sqlx.Connect("postgres", cfg.DB.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Run database migrations
	if err := RunMigrations(cfg.DB.DSN); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	redisClient, err := NewRedisClient(cfg.Redis)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Connections{DB: db, Redis: redisClient}, nil
}

func (c *Connections) Close() {
	if c.DB != nil {
		_ = c.DB.Close()
	}
	if c.Redis != nil {
		_ = c.Redis.Close()
	}
}
