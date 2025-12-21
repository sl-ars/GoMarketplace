package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-app-marketplace/internal/app/config"
	"go-app-marketplace/internal/elasticsearch"
	"go-app-marketplace/internal/repositories"
	"go-app-marketplace/pkg/logger"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const (
	maxRetries      = 3
	batchSize       = 10
	pollInterval    = 5 * time.Second
	cleanupInterval = 24 * time.Hour
)

func main() {
	configFile := flag.String("config", "./configs/.env", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.NewConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	appLogger := logger.New(cfg.Logger)
	appLogger.Info("Starting Elasticsearch synchronization worker")

	// Validate configuration
	if !cfg.Elasticsearch.Enabled {
		appLogger.Fatal("Elasticsearch is disabled. Set ES_ENABLED=true to run the worker.")
	}

	// Connect to database
	db, err := sqlx.Connect("postgres", cfg.DB.DSN)
	if err != nil {
		appLogger.WithError(err).Fatal("Failed to connect to database")
	}
	defer db.Close()

	// Initialize repositories
	outboxRepo := repositories.NewElasticsearchOutboxRepository(db)
	productRepo := repositories.NewProductRepository(db)
	offerRepo := repositories.NewOfferRepository(db)

	// Initialize Elasticsearch client
	esClient, err := elasticsearch.NewClient(
		cfg.Elasticsearch.Addresses,
		cfg.Elasticsearch.Username,
		cfg.Elasticsearch.Password,
		appLogger.Logger,
	)
	if err != nil {
		appLogger.WithError(err).Fatal("Failed to connect to Elasticsearch")
	}

	// Initialize product search service
	productSearchService := elasticsearch.NewProductSearchService(esClient)

	// Ensure index exists
	if err := productSearchService.InitializeIndex(context.Background()); err != nil {
		appLogger.WithError(err).Fatal("Failed to initialize Elasticsearch index")
	}

	appLogger.Info("Elasticsearch worker initialized and ready")

	// Start cleanup goroutine
	go func() {
		ticker := time.NewTicker(cleanupInterval)
		defer ticker.Stop()

		for range ticker.C {
			if err := outboxRepo.DeleteProcessedEvents(context.Background(), 7*24*time.Hour); err != nil {
				appLogger.WithError(err).Warn("Failed to cleanup old outbox events")
			} else {
				appLogger.Info("Cleaned up old processed outbox events")
			}
		}
	}()

	// Handle graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start processing loop
	go func() {
		ticker := time.NewTicker(pollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				processOutboxEvents(appLogger, outboxRepo, productRepo, offerRepo, productSearchService)
			}
		}
	}()

	// Start stats reporting
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				reportStats(appLogger, outboxRepo)
			}
		}
	}()

	appLogger.Info("Elasticsearch worker started. Processing outbox events...")

	// Wait for shutdown signal
	<-sigChan
	appLogger.Info("Shutting down Elasticsearch worker...")

	// Final stats report
	reportStats(appLogger, outboxRepo)
	appLogger.Info("Elasticsearch worker stopped")
}

func processOutboxEvents(
	logger *logger.Logger,
	outboxRepo *repositories.ElasticsearchOutboxRepository,
	productRepo *repositories.ProductRepository,
	offerRepo *repositories.OfferRepository,
	searchService *elasticsearch.ProductSearchService,
) {
	ctx := context.Background()

	// Get unprocessed events
	events, err := outboxRepo.GetUnprocessedEvents(ctx, batchSize)
	if err != nil {
		logger.WithError(err).Error("Failed to get unprocessed outbox events")
		return
	}

	if len(events) == 0 {
		return // No events to process
	}

	logger.WithField("count", len(events)).Debug("Processing outbox events")

	for _, event := range events {
		if err := processEvent(ctx, logger, event, outboxRepo, productRepo, offerRepo, searchService); err != nil {
			logger.WithError(err).WithFields(map[string]interface{}{
				"event_id":       event.ID,
				"aggregate_type": event.AggregateType,
				"aggregate_id":   event.AggregateID,
				"event_type":     event.EventType,
				"retry_count":    event.RetryCount,
			}).Error("Failed to process outbox event")

			// Mark as failed
			if markErr := outboxRepo.MarkEventFailed(ctx, event.ID, err.Error(), maxRetries); markErr != nil {
				logger.WithError(markErr).WithField("event_id", event.ID).Error("Failed to mark event as failed")
			}
		} else {
			// Mark as processed
			if markErr := outboxRepo.MarkEventProcessed(ctx, event.ID); markErr != nil {
				logger.WithError(markErr).WithField("event_id", event.ID).Error("Failed to mark event as processed")
			} else {
				logger.WithFields(map[string]interface{}{
					"event_id":       event.ID,
					"aggregate_type": event.AggregateType,
					"aggregate_id":   event.AggregateID,
					"event_type":     event.EventType,
				}).Info("Successfully processed outbox event")
			}
		}
	}
}

func processEvent(
	ctx context.Context,
	logger *logger.Logger,
	event *repositories.ElasticsearchOutboxEvent,
	outboxRepo *repositories.ElasticsearchOutboxRepository,
	productRepo *repositories.ProductRepository,
	offerRepo *repositories.OfferRepository,
	searchService *elasticsearch.ProductSearchService,
) error {
	switch event.AggregateType {
	case "product":
		return processProductEvent(ctx, event, productRepo, offerRepo, searchService)
	default:
		return nil // Skip unknown aggregate types
	}
}

func processProductEvent(
	ctx context.Context,
	event *repositories.ElasticsearchOutboxEvent,
	productRepo *repositories.ProductRepository,
	offerRepo *repositories.OfferRepository,
	searchService *elasticsearch.ProductSearchService,
) error {
	productID := event.AggregateID

	switch event.EventType {
	case "updated", "created":
		// Get product
		product, err := productRepo.GetProductByID(ctx, productID)
		if err != nil {
			return err
		}

		// Get all offers for this product
		offers, err := offerRepo.ListOffersByProduct(ctx, productID)
		if err != nil {
			return err
		}

		// Index/update in Elasticsearch
		return searchService.IndexProduct(ctx, product, offers)

	case "deleted":
		// Delete from Elasticsearch
		return searchService.DeleteProduct(ctx, productID)

	default:
		return nil // Skip unknown event types
	}
}

func reportStats(logger *logger.Logger, outboxRepo *repositories.ElasticsearchOutboxRepository) {
	ctx := context.Background()

	stats, err := outboxRepo.GetEventStats(ctx)
	if err != nil {
		logger.WithError(err).Warn("Failed to get outbox stats")
		return
	}

	logger.WithFields(map[string]interface{}{
		"total_events":       stats["total"],
		"unprocessed_events": stats["unprocessed"],
		"failed_events":      stats["failed"],
	}).Info("Outbox statistics")
}
