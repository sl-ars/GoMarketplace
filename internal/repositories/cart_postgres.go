package repositories

import (
	"context"
	"errors"
	"github.com/jmoiron/sqlx"
	"go-app-marketplace/pkg/domain"
)

type CartRepository struct {
	db *sqlx.DB
}

func NewCartRepository(db *sqlx.DB) *CartRepository {
	return &CartRepository{db: db}
}

func (r *CartRepository) checkStockAvailability(ctx context.Context, offerID int64, requestedQuantity int) error {
	var currentStock int
	err := r.db.GetContext(ctx, &currentStock, `
		SELECT stock 
		FROM offers 
		WHERE id = $1 AND is_available = true
	`, offerID)
	if err != nil {
		return err
	}

	if currentStock < requestedQuantity {
		return errors.New("not enough stock available")
	}
	return nil
}

func (r *CartRepository) checkCartQuantity(ctx context.Context, userID, offerID int64, requestedQuantity int) error {
	var currentQuantity int
	err := r.db.GetContext(ctx, &currentQuantity, `
		SELECT COALESCE(quantity, 0)
		FROM cart_items
		WHERE user_id = $1 AND offer_id = $2
	`, userID, offerID)
	if err != nil {
		return err
	}

	const maxQuantityPerUser = 10
	if currentQuantity+requestedQuantity > maxQuantityPerUser {
		return errors.New("maximum quantity per user exceeded")
	}
	return nil
}

func (r *CartRepository) AddItem(ctx context.Context, userID, offerID int64, quantity int) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO cart_items (user_id, offer_id, quantity) 
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, offer_id) 
		DO UPDATE SET quantity = cart_items.quantity + $3
	`, userID, offerID, quantity)
	return err
}

func (r *CartRepository) GetItems(ctx context.Context, userID int64) ([]domain.CartItem, error) {
	var items []domain.CartItem
	err := r.db.SelectContext(ctx, &items, `
		SELECT id, user_id, offer_id, quantity
		FROM cart_items
		WHERE user_id = $1
	`, userID)
	return items, err
}

func (r *CartRepository) RemoveItem(ctx context.Context, userID, offerID int64) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM cart_items 
		WHERE user_id = $1 AND offer_id = $2
	`, userID, offerID)
	return err
}

func (r *CartRepository) ClearCart(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM cart_items 
		WHERE user_id = $1
	`, userID)
	return err
}
