package repositories

import (
	"context"
	"errors"
	"github.com/jmoiron/sqlx"
	"go-app-marketplace/pkg/domain"
	"go-app-marketplace/pkg/reqresp"
)

type OrderRepository struct {
	db *sqlx.DB
}

func NewOrderRepository(db *sqlx.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) CreateOrder(ctx context.Context, userID int64, totalAmount float64, items []domain.OrderItem) (int64, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = tx.Rollback() // will rollback only if Commit hasn't been called
	}()

	var orderID int64
	err = tx.GetContext(ctx, &orderID, `
		INSERT INTO orders (user_id, total_amount, status, payment_status)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, userID, totalAmount, domain.OrderStatusPending, domain.PaymentStatusPending)
	if err != nil {
		return 0, err
	}

	for _, item := range items {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO order_items (order_id, offer_id, product_id, seller_id, quantity, unit_price, status)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, orderID, item.OfferID, item.ProductID, item.SellerID, item.Quantity, item.UnitPrice, domain.OrderItemStatusPending)
		if err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return orderID, nil
}

func (r *OrderRepository) CancelOrderItem(ctx context.Context, userID, itemID int64) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE order_items 
		SET status = $1 
		WHERE id = $2 
		  AND order_id IN (SELECT id FROM orders WHERE user_id = $3)
		  AND status != $1
	`, domain.OrderItemStatusCancelled, itemID, userID)
	return err
}

func (r *OrderRepository) ListOrders(ctx context.Context, userID int64) ([]*domain.Order, error) {
	var orders []*domain.Order
	err := r.db.SelectContext(ctx, &orders, `
		SELECT id, user_id, total_amount, status, payment_status, created_at, updated_at
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *OrderRepository) GetOrderByID(ctx context.Context, orderID int64) (*domain.Order, []domain.OrderItem, error) {
	var order domain.Order
	err := r.db.GetContext(ctx, &order, `
		SELECT id, user_id, total_amount, status, payment_status, created_at, updated_at
		FROM orders
		WHERE id = $1
	`, orderID)
	if err != nil {
		return nil, nil, err
	}

	var items []domain.OrderItem
	err = r.db.SelectContext(ctx, &items, `
		SELECT id, order_id, offer_id, product_id, seller_id, quantity, unit_price, status, created_at, updated_at
		FROM order_items
		WHERE order_id = $1
	`, orderID)
	if err != nil {
		return nil, nil, err
	}

	return &order, items, nil
}

func (r *OrderRepository) ListOrderItems(ctx context.Context, orderID int64) ([]domain.OrderItem, error) {
	var items []domain.OrderItem
	query := `
		SELECT id, order_id, offer_id, product_id, seller_id, quantity, unit_price, status, created_at, updated_at
		FROM order_items
		WHERE order_id = $1
	`
	err := r.db.SelectContext(ctx, &items, query, orderID)
	return items, err
}

func (r *OrderRepository) UpdatePaymentStatusByOrderID(ctx context.Context, orderIDStr string, status domain.PaymentStatus) error {
	query := `
		UPDATE orders
		SET payment_status = $1, updated_at = NOW()
		WHERE CAST(id AS TEXT) = $2
	`
	result, err := r.db.ExecContext(ctx, query, status, orderIDStr)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return errors.New("order not found for given order_id")
	}

	return nil
}

// Get order-item by id
func (r *OrderRepository) GetOrderItemByID(ctx context.Context, itemID int64) (*domain.OrderItem, error) {
	const q = `
		SELECT 
			oi.id, oi.order_id, oi.offer_id, oi.product_id,
			oi.seller_id, oi.quantity, oi.unit_price, oi.status,
			oi.created_at, oi.updated_at,
			o.user_id AS order_user_id     -- <-- ключевая строка
		FROM order_items oi
		JOIN orders o ON o.id = oi.order_id
		WHERE oi.id = $1
	`
	var item domain.OrderItem
	if err := r.db.GetContext(ctx, &item, q, itemID); err != nil {
		return nil, err
	}
	return &item, nil
}

// Seller changes status
func (r *OrderRepository) UpdateOrderItemStatus(ctx context.Context, itemID int64, status domain.OrderItemStatus) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE order_items
		SET status = $1, updated_at = now()
		WHERE id = $2
	`, status, itemID)
	return err
}

// List all order-items that belong to the given seller
func (r *OrderRepository) ListOrderItemsBySeller(
	ctx context.Context,
	sellerID int64,
) ([]reqresp.SellerOrderItem, error) {

	const q = `
	SELECT 
		oi.id            AS item_id,
		oi.order_id      AS order_id,
		oi.product_id    AS product_id,
		p.name           AS product_name,
		oi.quantity      AS quantity,
		oi.unit_price    AS unit_price,
		oi.status        AS status,
		(o.payment_status = 'successful') AS paid,
		o.created_at     AS placed_at,
		o.user_id        AS customer_id,
		u.username       AS customer_name,
		r.id             AS refund_id,
		r.status         AS refund_status,
		r.reason         AS refund_reason
	FROM order_items oi
	JOIN orders  o ON o.id  = oi.order_id
	JOIN products p ON p.id = oi.product_id
	JOIN users    u ON u.id = o.user_id
	LEFT JOIN refunds r ON r.order_item_id = oi.id
	WHERE oi.seller_id = $1
	ORDER BY o.created_at DESC`

	var rows []reqresp.SellerOrderItem
	if err := r.db.SelectContext(ctx, &rows, q, sellerID); err != nil {
		return nil, err
	}
	return rows, nil
}
