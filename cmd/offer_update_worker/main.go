package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go"

	"go-app-marketplace/internal/app/config"
	"go-app-marketplace/internal/app/connections"
	"go-app-marketplace/internal/messagebus"
	"go-app-marketplace/internal/redisdb"
	"go-app-marketplace/internal/repositories"
	"go-app-marketplace/internal/usecases"
	"go-app-marketplace/pkg/domain"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	conns, err := connections.NewConnections(cfg)
	if err != nil {
		log.Fatalf("failed to connect to DB/Redis: %v", err)
	}
	defer conns.Close()

	redisdb.SetClient(conns.Redis)

	rmqURL := os.Getenv("RABBITMQ_URL")
	if rmqURL == "" {
		rmqURL = cfg.RabbitMQURL
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

	exchange := "app.offers.exchange"
	routingKey := "offers.update"
	queueName := "offers.update.worker"

	// Declare topic exchange and the delay queue (TTL + DLX); plugin not used
	if err := ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil); err != nil {
		log.Fatalf("exchange declare failed: %v", err)
	}
	// ensure delay queue exists which dead-letters to the exchange + routing key
	delayQName := "offers.delay.queue"
	qArgs := amqp.Table{
		"x-dead-letter-exchange":    exchange,
		"x-dead-letter-routing-key": routingKey,
	}
	if _, err := ch.QueueDeclare(delayQName, true, false, false, false, qArgs); err != nil {
		log.Fatalf("failed to declare delay queue: %v", err)
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
		"offer_update_worker",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("failed to register consumer: %v", err)
	}

	log.Println("offer_update_worker started, waiting for messages...")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// create usecase repo to apply updates
	offerRepo := repositories.NewOfferRepository(conns.DB)
	offerUC := usecases.NewOfferUseCase(offerRepo)

	for {
		select {
		case <-ctx.Done():
			log.Println("shutting down worker...")
			return
		case m := <-msgs:
			if m.Body == nil {
				continue
			}
			var evt messagebus.OfferUpdateEvent
			if err := json.Unmarshal(m.Body, &evt); err != nil {
				log.Printf("failed to decode message: %v", err)
				continue
			}
			log.Printf("OfferUpdateEvent received: offer_id=%d seller_id=%d price=%.2f stock=%d available=%v",
				evt.OfferID, evt.SellerID, evt.Price, evt.Stock, evt.IsAvailable)

			// apply update via usecase
			if err := offerUC.UpdateOffer(context.Background(), &domain.Offer{
				ID:          evt.OfferID,
				SellerID:    evt.SellerID,
				Price:       evt.Price,
				Stock:       evt.Stock,
				IsAvailable: evt.IsAvailable,
			}); err != nil {
				log.Printf("failed to apply update: %v", err)
				continue
			}

			// After applying update, evict cache keys
			updatedOffer, err := offerUC.GetOfferByID(context.Background(), evt.OfferID)
			if err == nil {
				_ = redisdb.Rdb.Del(context.Background(), fmt.Sprintf("offer:%d", evt.OfferID))
				_ = redisdb.Rdb.Del(context.Background(), fmt.Sprintf("offers:product:%d", updatedOffer.ProductID))
			}
		}
	}
}
