package connections

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQConfig struct {
	URL string
}

func NewRabbitMQConn(cfg RabbitMQConfig) (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		return nil, nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, nil, err
	}

	log.Println("RabbitMQ connected")
	return conn, ch, nil
}
