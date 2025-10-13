package usecases

import (
	"context"
	"go-app-marketplace/pkg/domain"
)

type OfferRepository interface {
	CreateOffer(ctx context.Context, offer *domain.Offer) (int64, error)
	GetOfferByID(ctx context.Context, id int64) (*domain.Offer, error)
	ListOffersByProduct(ctx context.Context, productID int64) ([]*domain.Offer, error)
	UpdateOffer(ctx context.Context, offer *domain.Offer) error
	DeleteOffer(ctx context.Context, id int64, sellerID int64) error
	ListOffersBySeller(ctx context.Context, sellerID int64) ([]*domain.Offer, error)
}

type OfferUseCase struct {
	repo OfferRepository
}

func NewOfferUseCase(repo OfferRepository) *OfferUseCase {
	return &OfferUseCase{repo: repo}
}

func (uc *OfferUseCase) CreateOffer(ctx context.Context, offer *domain.Offer) (int64, error) {
	return uc.repo.CreateOffer(ctx, offer)
}

func (uc *OfferUseCase) GetOfferByID(ctx context.Context, id int64) (*domain.Offer, error) {
	return uc.repo.GetOfferByID(ctx, id)
}

func (uc *OfferUseCase) ListOffersByProduct(ctx context.Context, productID int64) ([]*domain.Offer, error) {
	return uc.repo.ListOffersByProduct(ctx, productID)
}

func (uc *OfferUseCase) UpdateOffer(ctx context.Context, offer *domain.Offer) error {
	return uc.repo.UpdateOffer(ctx, offer)
}

func (uc *OfferUseCase) DeleteOffer(ctx context.Context, id int64, sellerID int64) error {
	return uc.repo.DeleteOffer(ctx, id, sellerID)
}

func (uc *OfferUseCase) ListOffersBySeller(ctx context.Context, sellerID int64) ([]*domain.Offer, error) {
	return uc.repo.ListOffersBySeller(ctx, sellerID)
}
