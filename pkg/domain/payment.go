package domain

type PaymentIntentStatus string

const (
	PaymentIntentStatusSucceeded  PaymentIntentStatus = "succeeded"
	PaymentIntentStatusCanceled   PaymentIntentStatus = "canceled"
	PaymentIntentStatusProcessing PaymentIntentStatus = "processing"
)

type PaymentIntent struct {
	ID           string              `json:"id"`
	OrderID      int64               `json:"order_id"`
	Amount       float64             `json:"amount"`
	Currency     string              `json:"currency"`
	Status       PaymentIntentStatus `json:"status"`
	ClientSecret string              `json:"client_secret"`
}
