package services

import (
	"context"
	"errors"
	"fmt"
	"go-app-marketplace/internal/redisdb"
	"go-app-marketplace/internal/usecases"
	"go-app-marketplace/pkg/domain"
	"time"
)

type CartService struct {
	usecase *usecases.CartUseCase
}

func NewCartService(usecase *usecases.CartUseCase) *CartService {
	return &CartService{usecase: usecase}
}

func (s *CartService) AddItem(ctx context.Context, userID, offerID int64, quantity int) error {
	err := s.usecase.AddItem(ctx, userID, offerID, quantity)
	if err == nil {
		return nil
	}

	if errors.Is(err, usecases.ErrNotEnoughStock) {
		return usecases.ErrNotEnoughStock
	}
	if errors.Is(err, usecases.ErrMaxQuantityExceeded) {
		return usecases.ErrMaxQuantityExceeded
	}
	if errors.Is(err, usecases.ErrAddItemFailed) {
		return usecases.ErrAddItemFailed
	}

	_ = redisdb.Rdb.Del(ctx, fmt.Sprintf("cart:%d", userID))

	return err
}

func (s *CartService) GetCart(ctx context.Context, userID int64) ([]domain.CartItem, error) {
	key := fmt.Sprintf("cart:%d", userID)
	
	cart, err := redisdb.CacheGetOrSet(ctx,key, 2*time.Minute, func() ([]domain.CartItem, error) {
		return s.usecase.GetItems(ctx, userID)
	})
	if err != nil{
		return cart, err
	}
	return cart, nil
}

func (s *CartService) RemoveItem(ctx context.Context, userID, offerID int64) error {
	err := s.usecase.RemoveItem(ctx, userID, offerID)
	if err == nil{
		_ = redisdb.Rdb.Del(ctx, fmt.Sprintf("cart:%d", userID))
	}
	return err
}

func (s *CartService) ClearCart(ctx context.Context, userID int64) error {
	err := s.usecase.ClearCart(ctx, userID)
	if err == nil {
		_ = redisdb.Rdb.Del(ctx, fmt.Sprintf("cart:%d", userID))
	}
	return err
}
