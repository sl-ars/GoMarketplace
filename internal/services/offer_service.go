package services

import (
	"context"
	"fmt"
	"go-app-marketplace/internal/redisdb"
	"go-app-marketplace/internal/usecases"
	"go-app-marketplace/pkg/domain"
	"time"
)

type OfferService struct {
	usecase *usecases.OfferUseCase
}

func NewOfferService(uc *usecases.OfferUseCase) *OfferService {
	return &OfferService{usecase: uc}
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

	// Очистка кэша списка офферов по продукту
	key := fmt.Sprintf("offers:product:%d", productID)
	_ = redisdb.Rdb.Del(ctx, key)

	return id, nil
}

func (s *OfferService) GetOfferByID(ctx context.Context, id int64) (*domain.Offer, error) {
	key := fmt.Sprintf("offer:%d", id)

	offer, err := redisdb.CacheGetOrSet(ctx, key, 3*time.Minute, func() (*domain.Offer, error) {
		return s.usecase.GetOfferByID(ctx, id)
	})

	if err != nil{
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

	// Очистка кэша конкретного оффера
	offerKey := fmt.Sprintf("offer:%d", id)
	_ = redisdb.Rdb.Del(ctx, offerKey)
	return nil
}

func (s *OfferService) DeleteOffer(ctx context.Context, id, sellerID int64) error {
		err := s.usecase.DeleteOffer(ctx, id, sellerID)
	if err != nil {
		return err
	}

	// Очистка кэша по ID офферa
	key := fmt.Sprintf("offer:%d", id)
	_ = redisdb.Rdb.Del(ctx, key)

	return nil
}

func (s *OfferService) ListOffersBySeller(ctx context.Context, sellerID int64) ([]*domain.Offer, error) {
	return s.usecase.ListOffersBySeller(ctx, sellerID)
}
