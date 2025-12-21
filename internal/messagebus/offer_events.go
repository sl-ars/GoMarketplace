package messagebus

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	amqp "github.com/rabbitmq/amqp091-go"
)

// OfferUpdateEvent is published when a seller requests an update that should be applied later
type OfferUpdateEvent struct {
	OfferID     int64   `json:"offer_id"`
	SellerID    int64   `json:"seller_id"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	IsAvailable bool    `json:"is_available"`
}

// Interface used by services to publish delayed offer updates
type OfferEventPublisher interface {
	PublishOfferUpdateDelayed(ctx context.Context, evt OfferUpdateEvent, delayMs int) error
}

// RabbitMQOfferPublisher supports two modes:
// - pluginMode: uses x-delayed-message exchange with x-delay header (requires rabbitmq_delayed_message_exchange plugin)
// - ttlMode: uses a TTL queue with x-dead-letter-exchange (no plugin required); messages are published to the delay queue with per-message Expiration
type RabbitMQOfferPublisher struct {
	ch             *amqp.Channel
	exchange       string
	routingKey     string
	pluginMode     bool
	delayQueueName string // used in ttl mode: messages are published to this queue (via default exchange)
}

// NewRabbitMQOfferPublisher will try to declare an x-delayed-message exchange first; if RabbitMQ does not recognize it
// it will fall back to declaring a normal topic exchange and a delay queue that dead-letters to the exchange.
func NewRabbitMQOfferPublisher(ch *amqp.Channel, exchange, routingKey string) (*RabbitMQOfferPublisher, error) {
	// Declare topic exchange (no plugin dependency)
	if err := ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil); err != nil {
		return nil, err
	}

	// Declare a delay queue that dead-letters to the real exchange + routing key
	qArgs := amqp.Table{
		"x-dead-letter-exchange":    exchange,
		"x-dead-letter-routing-key": routingKey,
	}
	delayQName := "offers.delay.queue"
	if _, err := ch.QueueDeclare(delayQName, true, false, false, false, qArgs); err != nil {
		return nil, err
	}

	return &RabbitMQOfferPublisher{ch: ch, exchange: exchange, routingKey: routingKey, delayQueueName: delayQName}, nil
}

func (p *RabbitMQOfferPublisher) PublishOfferUpdateDelayed(ctx context.Context, evt OfferUpdateEvent, delayMs int) error {
	body, err := json.Marshal(evt)
	if err != nil {
		return err
	}

	if p.delayQueueName == "" {
		return fmt.Errorf("delay queue not configured")
	}

	return p.ch.PublishWithContext(ctx, "", p.delayQueueName, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/json",
		Body:         body,
		Type:         "offer.update",
		Expiration:   strconv.Itoa(delayMs), // milliseconds as string
	})
}
