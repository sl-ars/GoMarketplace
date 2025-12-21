package main

import (
	"fmt"
	"go-app-marketplace/internal/app/config"
	"go-app-marketplace/pkg/hash"
	"log"
	"math/rand"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	// Seed random for offer prices
	rand.Seed(time.Now().UnixNano())

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

	fmt.Println("🌱 Starting database seeding...")

	// Seed users
	seedUsers(db)

	// Seed products
	productIDs := seedProducts(db)

	// Seed offers
	seedOffers(db, productIDs)

	fmt.Println("\n✅ Database seeding completed successfully!")
}

// seedUsers creates initial users
func seedUsers(db *sqlx.DB) {
	fmt.Println("\n👥 Seeding users...")

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
		{"seller_four", "seller4@example.com", "seller", "password"},
		{"seller_five", "seller5@example.com", "seller", "password"},
		{"customer", "customer@example.com", "customer", "password"},
		{"customer2", "customer2@example.com", "customer", "password"},
		{"customer3", "customer3@example.com", "customer", "password"},
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

		fmt.Printf("  ✓ %s (%s)\n", u.username, u.role)
	}
}

// seedProducts creates initial products
func seedProducts(db *sqlx.DB) []int64 {
	fmt.Println("\n📱 Seeding products...")

	products := []struct {
		name        string
		description string
	}{
		// Smartphones
		{"iPhone 17 Pro Max", "Latest iPhone with advanced camera system and A18 chip"},
		{"Samsung Galaxy S25 Ultra", "Premium Android flagship with S Pen and 200MP camera"},
		{"Google Pixel 9 Pro", "AI-powered camera with 7 years of updates"},
		{"OnePlus 13", "Fast charging flagship with OxygenOS"},

		// Laptops
		{"MacBook Pro 16\" M4", "Professional laptop with M4 chip and Liquid Retina XDR display"},
		{"Dell XPS 15", "Ultra-thin laptop with Intel Core i9 and RTX 4070"},
		{"Lenovo ThinkPad X1 Carbon", "Business ultrabook with legendary keyboard"},
		{"ASUS ROG Zephyrus G16", "Gaming laptop with RTX 4090 and 240Hz display"},

		// Graphics Cards
		{"NVIDIA RTX 5090", "Flagship graphics card for 4K gaming and content creation"},
		{"NVIDIA RTX 5070 Ti", "High-end GPU with ray tracing and DLSS 4"},
		{"AMD Radeon RX 7900 XTX", "Competitive alternative with excellent rasterization"},
		{"NVIDIA RTX 4060 Ti", "Mid-range GPU perfect for 1440p gaming"},

		// Tablets
		{"iPad Pro 13\" M4", "Professional tablet with Apple Pencil Pro support"},
		{"Samsung Galaxy Tab S10+", "Android tablet with S Pen and DeX mode"},
		{"Microsoft Surface Pro 10", "2-in-1 laptop replacement with Intel Core Ultra"},

		// Accessories
		{"Apple AirPods Pro (2nd gen)", "Active noise cancellation with spatial audio"},
		{"Sony WH-1000XM5", "Industry-leading wireless noise canceling headphones"},
		{"Logitech MX Master 3S", "Advanced wireless mouse for productivity"},
		{"Razer BlackWidow V4 Pro", "Mechanical gaming keyboard with yellow switches"},
		{"Samsung 49\" Odyssey OLED G9", "Ultrawide gaming monitor with 240Hz refresh rate"},
		{"WD Black SN850X NVMe SSD 2TB", "Fastest PCIe 5.0 SSD for gaming and content creation"},
	}

	var productIDs []int64

	for _, p := range products {
		var productID int64
		err := db.QueryRow(`
			INSERT INTO products (name, description)
			VALUES ($1, $2)
			ON CONFLICT (name) DO UPDATE SET description = EXCLUDED.description
			RETURNING id
		`, p.name, p.description).Scan(&productID)

		if err != nil {
			log.Fatalf("failed to insert product %s: %v", p.name, err)
		}

		productIDs = append(productIDs, productID)
		fmt.Printf("  ✓ %s (ID: %d)\n", p.name, productID)
	}

	return productIDs
}

// seedOffers creates offers for products from different sellers
func seedOffers(db *sqlx.DB, productIDs []int64) {
	fmt.Println("\n💰 Seeding offers...")

	// Get seller user IDs
	var sellerIDs []int64
	err := db.Select(&sellerIDs, `
		SELECT id FROM users WHERE role = 'seller' ORDER BY id
	`)
	if err != nil {
		log.Fatalf("failed to get seller IDs: %v", err)
	}

	if len(sellerIDs) == 0 {
		log.Fatal("No sellers found. Please seed users first.")
	}

	// Define base prices for each product
	productBasePrices := map[int64]float64{}
	for i, productID := range productIDs {
		// Set different base prices based on product type
		var basePrice float64
		switch {
		case i < 4: // Smartphones
			basePrice = 800 + float64(i*200) // $800-$1400
		case i < 8: // Laptops
			basePrice = 1200 + float64(i-4)*400 // $1200-$2400
		case i < 12: // GPUs
			basePrice = 400 + float64(i-8)*300 // $400-$1300
		case i < 15: // Tablets
			basePrice = 600 + float64(i-12)*300 // $600-$1200
		default: // Accessories
			basePrice = 50 + float64(i-15)*100 // $50-$450
		}
		productBasePrices[productID] = basePrice
	}

	// Create offers from different sellers with varying prices and stock
	for _, productID := range productIDs {
		basePrice := productBasePrices[productID]

		// Each product gets offers from 2-4 random sellers
		numOffers := rand.Intn(3) + 2 // 2-4 offers per product
		selectedSellers := make([]int64, 0, numOffers)

		// Randomly select sellers
		for i := 0; i < numOffers && len(selectedSellers) < len(sellerIDs); i++ {
			sellerID := sellerIDs[rand.Intn(len(sellerIDs))]
			// Avoid duplicate sellers for same product
			found := false
			for _, existing := range selectedSellers {
				if existing == sellerID {
					found = true
					break
				}
			}
			if !found {
				selectedSellers = append(selectedSellers, sellerID)
			}
		}

		for _, sellerID := range selectedSellers {
			// Vary price by ±20% from base price
			priceVariation := (rand.Float64() - 0.5) * 0.4 // -20% to +20%
			price := basePrice * (1 + priceVariation)

			// Round to 2 decimal places
			price = float64(int(price*100)) / 100

			// Random stock between 1-50
			stock := rand.Intn(50) + 1

			// 90% chance of being available
			isAvailable := rand.Float32() < 0.9

			_, err := db.Exec(`
				INSERT INTO offers (product_id, seller_id, price, stock, is_available)
				VALUES ($1, $2, $3, $4, $5)
				ON CONFLICT (product_id, seller_id) DO UPDATE SET
					price = EXCLUDED.price,
					stock = EXCLUDED.stock,
					is_available = EXCLUDED.is_available
			`, productID, sellerID, price, stock, isAvailable)

			if err != nil {
				log.Fatalf("failed to insert offer for product %d, seller %d: %v", productID, sellerID, err)
			}

			status := "available"
			if !isAvailable {
				status = "unavailable"
			}

			fmt.Printf("  ✓ Product %d by Seller %d: $%.2f (%d in stock, %s)\n",
				productID, sellerID, price, stock, status)
		}
	}

	fmt.Printf("\n📊 Summary:\n")
	fmt.Printf("  • %d products created\n", len(productIDs))
	fmt.Printf("  • %d sellers available\n", len(sellerIDs))
	fmt.Printf("  • ~%d offers created\n", len(productIDs)*3) // average 3 offers per product
}
