package domain

type CartItem struct {
	ID       int64 `db:"id"`
	UserID   int64 `db:"user_id"`
	OfferID  int64 `db:"offer_id"`
	Quantity int   `db:"quantity"`
}
