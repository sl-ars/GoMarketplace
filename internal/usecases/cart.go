package usecases

import (
	"context"
	"errors"
	"go-app-marketplace/internal/repositories"
	"go-app-marketplace/pkg/domain"
)

var (
	ErrNotEnoughStock      = errors.New("not enough stock available")
	ErrMaxQuantityExceeded = errors.New("maximum quantity per user exceeded")
	ErrCartEmpty           = errors.New("cart is empty")
	ErrOfferNotFound       = errors.New("offer not found")
	ErrAddItemFailed       = errors.New("add item failed")
)

type CartUseCase struct {
	repo      *repositories.CartRepository
	offerRepo *repositories.OfferRepository
}

func NewCartUseCase(cartRepo *repositories.CartRepository, offerRepo *repositories.OfferRepository) *CartUseCase {
	return &CartUseCase{
		repo:      cartRepo,
		offerRepo: offerRepo,
	}
}

func (u *CartUseCase) AddItem(ctx context.Context, userID, offerID int64, quantity int) error {
	offer, err := u.offerRepo.GetOfferByID(ctx, offerID)
	if err != nil {
		return err
	}

	if !offer.IsAvailable {
		return ErrOfferNotFound
	}

	if offer.Stock < quantity {
		return ErrNotEnoughStock
	}

	items, err := u.repo.GetItems(ctx, userID)
	if err != nil {
		return err
	}

	var existingQty int
	for _, item := range items {
		if item.OfferID == offerID {
			existingQty = item.Quantity
			break
		}
	}
	if existingQty+quantity > 10 {
		return ErrMaxQuantityExceeded
	}

	return u.repo.AddItem(ctx, userID, offerID, quantity)
}

func (u *CartUseCase) GetItems(ctx context.Context, userID int64) ([]domain.CartItem, error) {
	return u.repo.GetItems(ctx, userID)
}

func (u *CartUseCase) RemoveItem(ctx context.Context, userID, offerID int64) error {
	return u.repo.RemoveItem(ctx, userID, offerID)
}

func (u *CartUseCase) ClearCart(ctx context.Context, userID int64) error {
	return u.repo.ClearCart(ctx, userID)
}
