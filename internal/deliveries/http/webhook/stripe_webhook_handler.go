package webhook

import (
	"encoding/json"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/webhook"
	"go-app-marketplace/internal/services"
	"go-app-marketplace/pkg/domain"
	"go-app-marketplace/pkg/httpx"
	"go-app-marketplace/pkg/reqresp"
	"io"
	"log"
	"net/http"
)

var _ = reqresp.StandardResponse{}

type StripeWebhookHandler struct {
	orderService  *services.OrderService
	signingSecret string
}

func NewStripeWebhookHandler(orderService *services.OrderService, secret string) *StripeWebhookHandler {
	return &StripeWebhookHandler{
		orderService:  orderService,
		signingSecret: secret,
	}
}

// @Summary Handle Stripe webhook events
// @Description Process Stripe webhook events for payment status updates
// @Tags webhooks
// @Accept json
// @Produce json
// @Param Stripe-Signature header string true "Stripe webhook signature"
// @Success 200 {object} reqresp.StandardResponse "Webhook received"
// @Failure 400 {object} reqresp.StandardResponse "Invalid webhook signature"
// @Failure 503 {object} reqresp.StandardResponse "Service unavailable"
// @Router /api/webhook/stripe [post]
func (h *StripeWebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	log.Println("Webhook triggered")

	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		httpx.WriteError(w, http.StatusServiceUnavailable, "Failed to read request body", err.Error())
		return
	}

	event, err := webhook.ConstructEvent(payload, r.Header.Get("Stripe-Signature"), h.signingSecret)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "Invalid Stripe signature", err.Error())
		return
	}

	switch event.Type {
	case "payment_intent.succeeded":
		var pi stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err == nil {
			orderIDStr := pi.Metadata["order_id"]
			if orderIDStr == "" {
				log.Println("Missing order_id in metadata")
				break
			}

			log.Printf("Payment succeeded for Order ID: %s", orderIDStr)

			err := h.orderService.UpdatePaymentStatusByOrderID(r.Context(), orderIDStr, domain.PaymentStatusSuccessful)
			if err != nil {
				log.Printf("Failed to update order status: %v", err)
			}
		}

	case "payment_intent.payment_failed":
		var pi stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err == nil {
			orderIDStr := pi.Metadata["order_id"]
			if orderIDStr == "" {
				log.Println("Missing order_id in metadata")
				break
			}

			log.Printf("Payment failed for Order ID: %s", orderIDStr)

			err := h.orderService.UpdatePaymentStatusByOrderID(r.Context(), orderIDStr, domain.PaymentStatusFailed)
			if err != nil {
				log.Printf("Failed to update order status: %v", err)
			}
		}
	}

	log.Println("Webhook OK")
	httpx.WriteSuccess(w, http.StatusOK, "Webhook processed", nil)
}
