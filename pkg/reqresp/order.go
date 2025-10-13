package reqresp

const (
	OrderStatusPending   = "pending"
	OrderStatusPaid      = "paid"
	OrderStatusCancelled = "cancelled"

	PaymentStatusPending    = "pending"
	PaymentStatusSuccessful = "successful"
	PaymentStatusFailed     = "failed"
)

type CheckoutResponse struct {
	OrderID     int64   `json:"order_id"`
	TotalAmount float64 `json:"total_amount"`
	PaymentURL  string  `json:"payment_url"`
}

type OrderItemCreateInput struct {
	OfferID   int64   `json:"offer_id" validate:"required"`
	Quantity  int     `json:"quantity" validate:"required,min=1"`
	UnitPrice float64 `json:"unit_price" validate:"required,min=0"`
}

type OrderResponse struct {
	ID            int64               `json:"id"`
	UserID        int64               `json:"user_id"`
	TotalAmount   float64             `json:"total_amount"`
	Status        string              `json:"status"`
	PaymentStatus string              `json:"payment_status"`
	Items         []OrderItemResponse `json:"items"`
	CreatedAt     string              `json:"created_at"`
}

type OrderItemResponse struct {
	ID        int64   `json:"id"`
	OfferID   int64   `json:"offer_id"`
	ProductID int64   `json:"product_id"`
	SellerID  int64   `json:"seller_id"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
	Status    string  `json:"status"`
}

type UpdateOrderItemStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=processing delivered"`
}
