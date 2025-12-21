package services

import (
	"context"
	"fmt"
	"go-app-marketplace/internal/messagebus"
	"go-app-marketplace/internal/redisdb"
	"go-app-marketplace/internal/repositories"
	"go-app-marketplace/internal/usecases"
	"go-app-marketplace/pkg/domain"
	"time"
)

type OfferService struct {
	usecase     *usecases.OfferUseCase
	productRepo *repositories.ProductRepository
	outboxRepo  *repositories.ElasticsearchOutboxRepository
	publisher   messagebus.OfferEventPublisher
}

func NewOfferService(uc *usecases.OfferUseCase, productRepo *repositories.ProductRepository, outboxRepo *repositories.ElasticsearchOutboxRepository, publisher messagebus.OfferEventPublisher) *OfferService {
	return &OfferService{
		usecase:     uc,
		productRepo: productRepo,
		outboxRepo:  outboxRepo,
		publisher:   publisher,
	}
}

func (s *OfferService) CreateOffer(ctx context.Context, productID, sellerID int64, price float64, stock int, isAvailable bool) (int64, error) {
	// Get database transaction from context or create a new one
	// For now, we'll create a transaction at the service level
	// In a real implementation, you might want to pass transactions from higher layers

	// Since the repository uses *sqlx.DB, we need to access it directly
	// This is a simplified approach - in production you'd want better transaction management

	offer := &domain.Offer{
		ProductID:   productID,
		SellerID:    sellerID,
		Price:       price,
		Stock:       stock,
		IsAvailable: isAvailable,
	}

	// For now, we'll use the existing approach but add outbox event
	// In a full implementation, you'd wrap this in a transaction
	id, err := s.usecase.CreateOffer(ctx, offer)
	if err != nil {
		return 0, err
	}

	// Clear cache
	key := fmt.Sprintf("offers:product:%d", productID)
	_ = redisdb.Rdb.Del(ctx, key)

	// Insert event into outbox (this should be in the same transaction as the offer creation)
	// For now, we'll do it separately - in production, use transaction
	eventPayload := map[string]interface{}{
		"product_id":   productID,
		"seller_id":    sellerID,
		"price":        price,
		"stock":        stock,
		"is_available": isAvailable,
	}

	if err := s.outboxRepo.InsertEvent(ctx, nil, "product", productID, "updated", eventPayload); err != nil {
		// Log error but don't fail the offer creation
		// In production, this should be in the same transaction
		fmt.Printf("Failed to insert outbox event: %v\n", err)
	}

	return id, nil
}

func (s *OfferService) GetOfferByID(ctx context.Context, id int64) (*domain.Offer, error) {
	key := fmt.Sprintf("offer:%d", id)

	offer, err := redisdb.CacheGetOrSet(ctx, key, 3*time.Minute, func() (*domain.Offer, error) {
		return s.usecase.GetOfferByID(ctx, id)
	})

	if err != nil {
		return nil, err
	}
	return offer, nil
}

func (s *OfferService) ListOffersByProduct(ctx context.Context, productID int64) ([]*domain.Offer, error) {
	key := fmt.Sprintf("offers:product:%d", productID)

	offers, err := redisdb.CacheGetOrSet(ctx, key, 2*time.Minute, func() ([]*domain.Offer, error) {
		return s.usecase.ListOffersByProduct(ctx, productID)
	})

	if err != nil {
		return nil, err
	}

	return offers, nil
}

func (s *OfferService) UpdateOffer(ctx context.Context, id, sellerID int64, price float64, stock int, isAvailable bool) error {
	// If a publisher exists, schedule the update to be applied later (default 1 minute)
	if s.publisher != nil {
		evt := messagebus.OfferUpdateEvent{
			OfferID:     id,
			SellerID:    sellerID,
			Price:       price,
			Stock:       stock,
			IsAvailable: isAvailable,
		}
		// schedule with 60_000 ms delay (1 minute)
		_ = s.publisher.PublishOfferUpdateDelayed(ctx, evt, 60_000)
		return nil
	}

	// fallback: apply immediately
	// Get the offer first to get the product ID
	existingOffer, err := s.usecase.GetOfferByID(ctx, id)
	if err != nil {
		return err
	}

	offer := &domain.Offer{
		ID:          id,
		SellerID:    sellerID,
		ProductID:   existingOffer.ProductID, // Preserve the product ID
		Price:       price,
		Stock:       stock,
		IsAvailable: isAvailable,
	}
	if err := s.usecase.UpdateOffer(ctx, offer); err != nil {
		return err
	}

	// Clear cache
	offerKey := fmt.Sprintf("offer:%d", id)
	_ = redisdb.Rdb.Del(ctx, offerKey)

	// Insert event into outbox
	eventPayload := map[string]interface{}{
		"offer_id":     id,
		"product_id":   existingOffer.ProductID,
		"seller_id":    sellerID,
		"price":        price,
		"stock":        stock,
		"is_available": isAvailable,
	}

	if err := s.outboxRepo.InsertEvent(ctx, nil, "product", existingOffer.ProductID, "updated", eventPayload); err != nil {
		fmt.Printf("Failed to insert outbox event: %v\n", err)
	}

	return nil
}

func (s *OfferService) DeleteOffer(ctx context.Context, id, sellerID int64) error {
	// Get the offer first to get the product ID before deleting
	offer, err := s.usecase.GetOfferByID(ctx, id)
	if err != nil {
		return err
	}
	productID := offer.ProductID

	err = s.usecase.DeleteOffer(ctx, id, sellerID)
	if err != nil {
		return err
	}

	// Clear cache
	key := fmt.Sprintf("offer:%d", id)
	_ = redisdb.Rdb.Del(ctx, key)

	// Insert event into outbox
	eventPayload := map[string]interface{}{
		"offer_id":   id,
		"product_id": productID,
		"seller_id":  sellerID,
	}

	if err := s.outboxRepo.InsertEvent(ctx, nil, "product", productID, "updated", eventPayload); err != nil {
		fmt.Printf("Failed to insert outbox event: %v\n", err)
	}

	return nil
}

func (s *OfferService) ListOffersBySeller(ctx context.Context, sellerID int64) ([]*domain.Offer, error) {
	return s.usecase.ListOffersBySeller(ctx, sellerID)
}
