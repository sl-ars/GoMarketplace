package repositories

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
	"go-app-marketplace/pkg/domain"
	"time"
)

var (
	ErrRefundTooLate         = errors.New("refund window (14 days) has expired")
	ErrRefundAlreadyExists   = errors.New("refund already requested for this item")
	ErrRefundNotFound        = sql.ErrNoRows
	ErrRefundStatusForbidden = errors.New("illegal refund status transition")
)

type RefundRepository struct{ db *sqlx.DB }

func NewRefundRepository(db *sqlx.DB) *RefundRepository { return &RefundRepository{db} }

// customer side — request refund
func (r *RefundRepository) Create(ctx context.Context, item domain.OrderItem, amount float64, reason string) (int64, error) {
	// 14-days rule
	if time.Since(item.UpdatedAt) > 14*24*time.Hour {
		return 0, ErrRefundTooLate
	}
	// check duplicate
	var exists bool
	_ = r.db.GetContext(ctx, &exists, `SELECT TRUE FROM refunds WHERE order_item_id=$1`, item.ID)
	if exists {
		return 0, ErrRefundAlreadyExists
	}
	var id int64
	err := r.db.GetContext(ctx, &id, `
		INSERT INTO refunds (order_item_id, requester_id, seller_id, amount, reason)
		VALUES ($1,$2,$3,$4,$5) RETURNING id`,
		item.ID, item.OrderID, item.SellerID, amount, reason)
	return id, err
}

// seller side — approve / reject
func (r *RefundRepository) UpdateStatus(ctx context.Context, refundID int64, next domain.RefundStatus) error {
	// allowed only from pending
	res, err := r.db.ExecContext(ctx,
		`UPDATE refunds SET status=$1, updated_at=now()
		  WHERE id=$2 AND status='pending'`, next, refundID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrRefundStatusForbidden
	}
	return nil
}

func (r *RefundRepository) GetByID(ctx context.Context, id int64) (*domain.Refund, error) {
	var rf domain.Refund
	err := r.db.GetContext(ctx, &rf, `SELECT * FROM refunds WHERE id=$1`, id)
	if err != nil {
		return nil, err
	}
	return &rf, nil
}
