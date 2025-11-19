package messagebus

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

type OrderCreatedEvent struct {
	OrderID     int64   `json:"order_id"`
	UserID      int64   `json:"user_id"`
	TotalAmount float64 `json:"total_amount"`
}

// Interface used by your usecase/service
type OrderEventPublisher interface {
	PublishOrderCreated(ctx context.Context, evt OrderCreatedEvent) error
}

// Concrete RabbitMQ implementation
type RabbitMQOrderPublisher struct {
	ch         *amqp.Channel
	exchange   string
	routingKey string
}

func NewRabbitMQOrderPublisher(ch *amqp.Channel, exchange, routingKey string) (*RabbitMQOrderPublisher, error) {
	// Make sure exchange exists (idempotent)
	if err := ch.ExchangeDeclare(
		exchange,
		"topic",
		true,  // durable
		false, // auto-delete
		false, // internal
		false, // no-wait
		nil,
	); err != nil {
		return nil, err
	}

	return &RabbitMQOrderPublisher{
		ch:         ch,
		exchange:   exchange,
		routingKey: routingKey,
	}, nil
}

func (p *RabbitMQOrderPublisher) PublishOrderCreated(ctx context.Context, evt OrderCreatedEvent) error {
	body, err := json.Marshal(evt)
	if err != nil {
		return err
	}

	return p.ch.PublishWithContext(ctx,
		p.exchange,
		p.routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
			Type:        "order.created",
		},
	)
}
