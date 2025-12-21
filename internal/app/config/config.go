//config.go
package config

import (
	"log"
	"os"

	"go-app-marketplace/pkg/logger"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

type Config struct {
	HTTPServer          HTTPServerConfig    `envPrefix:"HTTP_"`
	DB                  *DBConfig           `envPrefix:"DB_"`
	Redis               RedisConfig         `envPrefix:"REDIS_"`
	Elasticsearch       ElasticsearchConfig `envPrefix:"ES_"`
	RateLimit           RateLimitConfig     `envPrefix:"RATELIMIT_"`
	JWTSecret           string              `env:"JWT_SECRET"`
	StripeSecretKey     string              `env:"STRIPE_SECRET_KEY"`
	StripeWebhookSecret string              `env:"STRIPE_WEBHOOK_SECRET"`
	RabbitMQURL         string              `env:"RABBIT_MQ_URL"`
	Logger              logger.Config       `envPrefix:"LOG_"`
}

// RateLimitConfig defines rate limiting settings
type RateLimitConfig struct {
	Enabled          bool `env:"ENABLED" envDefault:"true"`
	AuthRequests     int  `env:"AUTH_REQUESTS" envDefault:"5"`
	PublicRequests   int  `env:"PUBLIC_REQUESTS" envDefault:"60"`
	StandardRequests int  `env:"STANDARD_REQUESTS" envDefault:"120"`
}

type RedisConfig struct {
	Addr     string `env:"ADDR" envDefault:"localhost:6379"`
	Password string `env:"PASSWORD" envDefault:""`
	DB       int    `env:"DB" envDefault:"0"`
}

type ElasticsearchConfig struct {
	Enabled   bool     `env:"ENABLED" envDefault:"true"`
	Addresses []string `env:"ADDRESSES" envDefault:"http://localhost:9200" envSeparator:","`
	Username  string   `env:"USERNAME" envDefault:""`
	Password  string   `env:"PASSWORD" envDefault:""`
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
