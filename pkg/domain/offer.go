package domain

import "time"

type Offer struct {
	ID          int64     `db:"id"`
	ProductID   int64     `db:"product_id"`
	SellerID    int64     `db:"seller_id"`
	Price       float64   `db:"price"`
	Stock       int       `db:"stock"`
	IsAvailable bool      `db:"is_available"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}
