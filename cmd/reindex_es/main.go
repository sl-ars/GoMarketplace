package main

import (
	"context"
	"fmt"
	"go-app-marketplace/internal/app/config"
	"go-app-marketplace/internal/elasticsearch"
	"go-app-marketplace/internal/repositories"
	"go-app-marketplace/pkg/logger"
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

	// Connect to database
	db, err := sqlx.Connect("postgres", cfg.DB.DSN)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize logger
	appLogger := logger.New(logger.Config{
		Level:  "info",
		Format: "text",
	})

	// Initialize Elasticsearch
	esClient, err := elasticsearch.NewClient(
		cfg.Elasticsearch.Addresses,
		cfg.Elasticsearch.Username,
		cfg.Elasticsearch.Password,
		appLogger.Logger,
	)
	if err != nil {
		log.Fatalf("failed to connect to Elasticsearch: %v", err)
	}

	productSearchService := elasticsearch.NewProductSearchService(esClient)

	// Initialize index
	if err := productSearchService.InitializeIndex(context.Background()); err != nil {
		log.Fatalf("failed to initialize index: %v", err)
	}

	// Get repositories
	productRepo := repositories.NewProductRepository(db)
	offerRepo := repositories.NewOfferRepository(db)

	fmt.Println("🔄 Starting Elasticsearch re-indexing...")

	// Get all products
	products, err := productRepo.GetAllProducts(context.Background())
	if err != nil {
		log.Fatalf("failed to get products: %v", err)
	}

	fmt.Printf("Found %d products to index\n", len(products))

	indexed := 0
	for _, product := range products {
		// Get all offers for this product
		offers, err := offerRepo.ListOffersByProduct(context.Background(), product.ID)
		if err != nil {
			log.Printf("failed to get offers for product %d: %v", product.ID, err)
			continue
		}

		// Index the product with its offers
		if err := productSearchService.IndexProduct(context.Background(), product, offers); err != nil {
			log.Printf("failed to index product %d: %v", product.ID, err)
			continue
		}

		indexed++
		fmt.Printf("  ✓ Indexed product: %s (ID: %d, %d offers)\n", product.Name, product.ID, len(offers))
	}

	fmt.Printf("\n✅ Successfully indexed %d/%d products in Elasticsearch!\n", indexed, len(products))

	// You can verify with: curl localhost:9200/products/_count
	fmt.Println("📊 To verify: curl localhost:9200/products/_count")
}
