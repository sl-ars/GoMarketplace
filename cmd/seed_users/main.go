package main

import (
	"fmt"
	"go-app-marketplace/internal/app/config"
	"go-app-marketplace/pkg/hash"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	// Load config
	cfg, err := config.NewConfig("./configs/.env")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Connect to database via sqlx
	db, err := sqlx.Connect("postgres", cfg.DB.DSN)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	users := []struct {
		username string
		email    string
		role     string
		password string
	}{
		{"admin", "admin@example.com", "admin", "password"},
		{"seller_one", "seller1@example.com", "seller", "password"},
		{"seller_two", "seller2@example.com", "seller", "password"},
		{"seller_three", "seller3@example.com", "seller", "password"},
		{"customer", "customer@example.com", "customer", "password"},
	}

	for _, u := range users {
		hashedPassword, err := hash.HashPassword(u.password)
		if err != nil {
			log.Fatalf("failed to hash password for user %s: %v", u.username, err)
		}

		_, err = db.Exec(`
			INSERT INTO users (username, email, password_hash, role)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (email) DO NOTHING
		`, u.username, u.email, hashedPassword, u.role)

		if err != nil {
			log.Fatalf("failed to insert user %s: %v", u.username, err)
		}

		fmt.Printf("Inserted (or skipped) user: %s\n", u.username)
	}

	fmt.Println("\nDone seeding users!")
}
