package domain

import "time"

type RefundStatus string

const (
	RefundPending   RefundStatus = "pending"
	RefundApproved  RefundStatus = "approved"
	RefundRejected  RefundStatus = "rejected"
	RefundCompleted RefundStatus = "completed"
)

type Refund struct {
	ID          int64        `db:"id"`
	OrderItemID int64        `db:"order_item_id"`
	RequesterID int64        `db:"requester_id"`
	SellerID    int64        `db:"seller_id"`
	Amount      float64      `db:"amount"`
	Reason      string       `db:"reason"`
	Status      RefundStatus `db:"status"`
	CreatedAt   time.Time    `db:"created_at"`
	UpdatedAt   time.Time    `db:"updated_at"`
}
