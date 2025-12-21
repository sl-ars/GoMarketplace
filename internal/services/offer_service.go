package services

import (
	"context"
	"fmt"
	"go-app-marketplace/internal/messagebus"
	"go-app-marketplace/internal/redisdb"
	"go-app-marketplace/internal/usecases"
	"go-app-marketplace/pkg/domain"
	"time"
)

type OfferService struct {
	usecase   *usecases.OfferUseCase
	publisher messagebus.OfferEventPublisher
}

func NewOfferService(uc *usecases.OfferUseCase, publisher messagebus.OfferEventPublisher) *OfferService {
	return &OfferService{usecase: uc, publisher: publisher}
}

func (s *OfferService) CreateOffer(ctx context.Context, productID, sellerID int64, price float64, stock int, isAvailable bool) (int64, error) {
	offer := &domain.Offer{
		ProductID:   productID,
		SellerID:    sellerID,
		Price:       price,
		Stock:       stock,
		IsAvailable: isAvailable,
	}
	id, err := s.usecase.CreateOffer(ctx, offer)
	if err != nil {
		return 0, err
	}

	// clear product offers cache
	key := fmt.Sprintf("offers:product:%d", productID)
	_ = redisdb.Rdb.Del(ctx, key)

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
	offer := &domain.Offer{
		ID:          id,
		SellerID:    sellerID,
		Price:       price,
		Stock:       stock,
		IsAvailable: isAvailable,
	}
	if err := s.usecase.UpdateOffer(ctx, offer); err != nil {
		return err
	}

	// clear offer cache
	offerKey := fmt.Sprintf("offer:%d", id)
	_ = redisdb.Rdb.Del(ctx, offerKey)
	return nil
}

func (s *OfferService) DeleteOffer(ctx context.Context, id, sellerID int64) error {
	err := s.usecase.DeleteOffer(ctx, id, sellerID)
	if err != nil {
		return err
	}

	// clear cache for the offer id
	key := fmt.Sprintf("offer:%d", id)
	_ = redisdb.Rdb.Del(ctx, key)

	return nil
}

func (s *OfferService) ListOffersBySeller(ctx context.Context, sellerID int64) ([]*domain.Offer, error) {
	return s.usecase.ListOffersBySeller(ctx, sellerID)
}
