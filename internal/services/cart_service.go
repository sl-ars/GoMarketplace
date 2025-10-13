package services

import (
	"context"
	"errors"
	"go-app-marketplace/internal/usecases"
	"go-app-marketplace/pkg/domain"
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

	return err
}

func (s *CartService) GetCart(ctx context.Context, userID int64) ([]domain.CartItem, error) {
	return s.usecase.GetItems(ctx, userID)
}

func (s *CartService) RemoveItem(ctx context.Context, userID, offerID int64) error {
	return s.usecase.RemoveItem(ctx, userID, offerID)
}

func (s *CartService) ClearCart(ctx context.Context, userID int64) error {
	return s.usecase.ClearCart(ctx, userID)
}
