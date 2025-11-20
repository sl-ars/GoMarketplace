package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go"
	"go-app-marketplace/internal/messagebus"
)

func main() {
	rmqURL := os.Getenv("RABBITMQ_URL")
	if rmqURL == "" {
		rmqURL = "amqp://guest:guest@localhost:5672/"
	}

	conn, err := amqp.Dial(rmqURL)
	if err != nil {
		log.Fatalf("failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("failed to open channel: %v", err)
	}
	defer ch.Close()

	exchange := "app.exchange"
	routingKey := "orders.created"
	queueName := "orders.created.worker"

	// Declare exchange & queue & binding (idempotent)
	if err := ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil); err != nil {
		log.Fatalf("exchange declare: %v", err)
	}

	q, err := ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		log.Fatalf("queue declare: %v", err)
	}

	if err := ch.QueueBind(q.Name, routingKey, exchange, false, nil); err != nil {
		log.Fatalf("queue bind: %v", err)
	}

	msgs, err := ch.Consume(
		q.Name,
		"order_events_worker",
		true,  // auto-ack for simplicity
		false, // not exclusive
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("failed to register consumer: %v", err)
	}

	log.Println("order_events_worker started, waiting for messages...")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("shutting down worker...")
			return
		case m := <-msgs:
			if m.Body == nil {
				continue
			}
			var evt messagebus.OrderCreatedEvent
			if err := json.Unmarshal(m.Body, &evt); err != nil {
				log.Printf("failed to decode message: %v", err)
				continue
			}
			log.Printf("OrderCreatedEvent received: order_id=%d user_id=%d total=%.2f",
				evt.OrderID, evt.UserID, evt.TotalAmount)

			// TODO: here you can:
			// - call email sending service
			// - push notification
			// - write to some analytics store
		}
	}
}
