package domain

import "time"

type OrderStatus string
type PaymentStatus string
type OrderItemStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusCancelled OrderStatus = "cancelled"

	PaymentStatusPending    PaymentStatus = "pending"
	PaymentStatusSuccessful PaymentStatus = "successful"
	PaymentStatusFailed     PaymentStatus = "failed"

	OrderItemStatusPending    OrderItemStatus = "pending"
	OrderItemStatusProcessing OrderItemStatus = "processing"
	OrderItemStatusDelivered  OrderItemStatus = "delivered"
	OrderItemStatusCancelled  OrderItemStatus = "cancelled"
)

type Order struct {
	ID            int64         `db:"id"`
	UserID        int64         `db:"user_id"`
	TotalAmount   float64       `db:"total_amount"`
	Status        OrderStatus   `db:"status"`
	PaymentStatus PaymentStatus `db:"payment_status"`
	CreatedAt     time.Time     `db:"created_at"`
	UpdatedAt     time.Time     `db:"updated_at"`
}

type OrderItem struct {
	ID        int64           `db:"id"`
	OrderID   int64           `db:"order_id"`
	OfferID   int64           `db:"offer_id"`
	ProductID int64           `db:"product_id"`
	SellerID  int64           `db:"seller_id"`
	Quantity  int             `db:"quantity"`
	UnitPrice float64         `db:"unit_price"`
	Status    OrderItemStatus `db:"status"`
	CreatedAt time.Time       `db:"created_at"`
	UpdatedAt time.Time       `db:"updated_at"`

	OrderUserID int64 `db:"order_user_id" json:"-"`
}
