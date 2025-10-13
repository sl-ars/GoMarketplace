package services

import (
	"fmt"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/checkout/session"

	"strconv"
)

type PaymentService struct {
	secretKey     string
	webhookSecret string
}

func NewPaymentService(secretKey, webhookSecret string) *PaymentService {
	stripe.Key = secretKey
	return &PaymentService{
		secretKey:     secretKey,
		webhookSecret: webhookSecret,
	}
}

func (p *PaymentService) GetWebhookSecret() string {
	return p.webhookSecret
}

func (p *PaymentService) CreateCheckoutSession(orderID int64, amount float64, successURL, cancelURL string) (*stripe.CheckoutSession, error) {
	params := &stripe.CheckoutSessionParams{
		PaymentIntentData: &stripe.CheckoutSessionPaymentIntentDataParams{
			Metadata: map[string]string{
				"order_id": strconv.FormatInt(orderID, 10),
			},
		},
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("usd"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String(fmt.Sprintf("Order #%d", orderID)),
					},
					UnitAmount: stripe.Int64(int64(amount * 100)),
				},
				Quantity: stripe.Int64(1),
			},
		},
		Mode:               stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL:         stripe.String(successURL),
		CancelURL:          stripe.String(cancelURL),
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
	}
	return session.New(params)
}
