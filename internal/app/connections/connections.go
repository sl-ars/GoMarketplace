package connections

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go-app-marketplace/internal/app/config"
)

type Connections struct {
	DB *sqlx.DB
}

func NewConnections(cfg *config.Config) (*Connections, error) {

	db, err := sqlx.Connect("postgres", cfg.DB.DSN)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to database: %v", err)
	}
	return &Connections{DB: db}, nil
}

func (c *Connections) Close() {
	if c.DB != nil {
		_ = c.DB.Close()
	}
}
