package repositories

import (
	"context"
	"github.com/jmoiron/sqlx"
	"go-app-marketplace/pkg/domain"
)

type OfferRepository struct {
	db *sqlx.DB
}

func NewOfferRepository(db *sqlx.DB) *OfferRepository {
	return &OfferRepository{db: db}
}

func (r *OfferRepository) CreateOffer(ctx context.Context, offer *domain.Offer) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO offers (product_id, seller_id, price, stock, is_available)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, offer.ProductID, offer.SellerID, offer.Price, offer.Stock, offer.IsAvailable).Scan(&id)
	return id, err
}

func (r *OfferRepository) GetOfferByID(ctx context.Context, id int64) (*domain.Offer, error) {
	var offer domain.Offer
	err := r.db.GetContext(ctx, &offer, `
		SELECT id, product_id, seller_id, price, stock, is_available, created_at, updated_at
		FROM offers
		WHERE id = $1
	`, id)
	if err != nil {
		return nil, err
	}
	return &offer, nil
}

func (r *OfferRepository) ListOffersByProduct(ctx context.Context, productID int64) ([]*domain.Offer, error) {
	var offers []*domain.Offer
	err := r.db.SelectContext(ctx, &offers, `
		SELECT id, product_id, seller_id, price, stock, is_available, created_at, updated_at
		FROM offers
		WHERE product_id = $1
		ORDER BY price ASC
	`, productID)
	return offers, err
}

func (r *OfferRepository) UpdateOffer(ctx context.Context, offer *domain.Offer) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE offers
		SET price = $1, stock = $2, is_available = $3, updated_at = NOW()
		WHERE id = $4 AND seller_id = $5
	`, offer.Price, offer.Stock, offer.IsAvailable, offer.ID, offer.SellerID)
	return err
}

func (r *OfferRepository) DeleteOffer(ctx context.Context, id int64, sellerID int64) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM offers
		WHERE id = $1 AND seller_id = $2
	`, id, sellerID)
	return err
}

func (r *OfferRepository) ListOffersBySeller(ctx context.Context, sellerID int64) ([]*domain.Offer, error) {
	var offers []*domain.Offer
	err := r.db.SelectContext(ctx, &offers, `
		SELECT id, product_id, seller_id, price, stock, is_available, created_at, updated_at
		FROM offers
		WHERE seller_id = $1
		ORDER BY updated_at DESC
	`, sellerID)
	return offers, err
}
