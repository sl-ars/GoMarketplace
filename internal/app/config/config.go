package config

import (
	"log"
	"os"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	"go-app-marketplace/pkg/logger"
)

type Config struct {
	HTTPServer          HTTPServerConfig `envPrefix:"HTTP_"`
	DB                  *DBConfig        `envPrefix:"DB_"`
	JWTSecret           string           `env:"JWT_SECRET"`
	StripeSecretKey     string           `env:"STRIPE_SECRET_KEY"`
	StripeWebhookSecret string           `env:"STRIPE_WEBHOOK_SECRET"`
	Logger              logger.Config    `envPrefix:"LOG_"`
}

type HTTPServerConfig struct {
	Port string `env:"PORT" envDefault:"8080"`
}

type DBConfig struct {
	DSN string `env:"DB_DSN"`
}

func NewConfig(filenames ...string) (*Config, error) {
	_ = godotenv.Load(filenames...)

	cfg := &Config{DB: &DBConfig{}}

	if err := env.Parse(cfg); err != nil {
		log.Printf("env.Parse failed: %v", err)
	}

	if cfg.DB.DSN == "" {
		cfg.DB.DSN = os.Getenv("DB_DSN")
	}

	return cfg, nil
}
