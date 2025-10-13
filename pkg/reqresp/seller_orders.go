package reqresp

import "time"

type SellerOrderItem struct {
	ItemID       int64     `db:"item_id"       json:"item_id"`
	OrderID      int64     `db:"order_id"      json:"order_id"`
	ProductID    int64     `db:"product_id"    json:"product_id"`
	ProductName  string    `db:"product_name"  json:"product_name"`
	Quantity     int       `db:"quantity"      json:"quantity"`
	UnitPrice    float64   `db:"unit_price"    json:"unit_price"`
	Status       string    `db:"status"        json:"status"`
	Paid         bool      `db:"paid"          json:"paid"`
	PlacedAt     time.Time `db:"placed_at"     json:"placed_at"`
	CustomerID   int64     `db:"customer_id"   json:"customer_id"`
	CustomerName string    `db:"customer_name" json:"customer_name"`
	RefundID     *int64    `db:"refund_id"     json:"refund_id,omitempty"`
	RefundStatus *string   `db:"refund_status" json:"refund_status,omitempty"`
	RefundReason *string   `db:"refund_reason" json:"refund_reason,omitempty"`
}
