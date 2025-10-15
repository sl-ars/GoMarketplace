package services

import (
	"context"
	"go-app-marketplace/internal/usecases"
	"go-app-marketplace/pkg/domain"
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
	return s.usecase.CreateOffer(ctx, offer)
}

func (s *OfferService) GetOfferByID(ctx context.Context, id int64) (*domain.Offer, error) {
	return s.usecase.GetOfferByID(ctx, id)
}

func (s *OfferService) ListOffersByProduct(ctx context.Context, productID int64) ([]*domain.Offer, error) {
	return s.usecase.ListOffersByProduct(ctx, productID)
}

func (s *OfferService) UpdateOffer(ctx context.Context, id, sellerID int64, price float64, stock int, isAvailable bool) error {
	offer := &domain.Offer{
		ID:          id,
		SellerID:    sellerID,
		Price:       price,
		Stock:       stock,
		IsAvailable: isAvailable,
	}
	return s.usecase.UpdateOffer(ctx, offer)
}

func (s *OfferService) DeleteOffer(ctx context.Context, id, sellerID int64) error {
	return s.usecase.DeleteOffer(ctx, id, sellerID)
}

func (s *OfferService) ListOffersBySeller(ctx context.Context, sellerID int64) ([]*domain.Offer, error) {
	return s.usecase.ListOffersBySeller(ctx, sellerID)
}
