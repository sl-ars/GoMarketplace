package reqresp

type AddItemToCartRequest struct {
	OfferID  int64 `json:"offer_id" validate:"required"`
	Quantity int   `json:"quantity" validate:"required,min=1"`
}

type CartItemResponse struct {
	OfferID     int64   `json:"offer_id"`
	ProductID   int64   `json:"product_id"`
	SellerID    int64   `json:"seller_id"`
	Price       float64 `json:"price"`
	Quantity    int     `json:"quantity"`
	IsAvailable bool    `json:"is_available"`
}
