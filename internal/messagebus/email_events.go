package messagebus

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// EmailType represents the type of email to send
type EmailType string

const (
	EmailTypeVerification  EmailType = "email.verification"
	EmailTypePasswordReset EmailType = "email.password_reset"
	EmailTypeWelcome       EmailType = "email.welcome"
	EmailTypeOrderConfirm  EmailType = "email.order_confirm"
)

// EmailMessage represents an email to be sent via queue
type EmailMessage struct {
	Type      EmailType              `json:"type"`
	To        string                 `json:"to"`
	Subject   string                 `json:"subject"`
	Data      map[string]interface{} `json:"data"`
	Priority  int                    `json:"priority"` // 0-9, higher = more important
	Timestamp int64                  `json:"timestamp"`
}

// VerificationEmailData holds data for verification email
type VerificationEmailData struct {
	Username  string `json:"username"`
	VerifyURL string `json:"verify_url"`
	Code      string `json:"code"`
	ExpiresIn string `json:"expires_in"`
}

// PasswordResetEmailData holds data for password reset email
type PasswordResetEmailData struct {
	Username  string `json:"username"`
	ResetURL  string `json:"reset_url"`
	ExpiresIn string `json:"expires_in"`
}

// EmailPublisher publishes email messages to RabbitMQ
type EmailPublisher struct {
	channel      *amqp.Channel
	exchangeName string
}

// NewEmailPublisher creates a new email publisher
func NewEmailPublisher(ch *amqp.Channel, exchangeName string) (*EmailPublisher, error) {
	// Declare the exchange
	err := ch.ExchangeDeclare(
		exchangeName, // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declare the email queue
	queue, err := ch.QueueDeclare(
		"emails", // name
		true,     // durable
		false,    // delete when unused
		false,    // exclusive
		false,    // no-wait
		amqp.Table{
			"x-message-ttl":          86400000, // 24 hours TTL
			"x-dead-letter-exchange": exchangeName + ".dlx",
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue to exchange for all email routing keys
	err = ch.QueueBind(
		queue.Name,   // queue name
		"email.*",    // routing key pattern
		exchangeName, // exchange
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to bind queue: %w", err)
	}

	return &EmailPublisher{
		channel:      ch,
		exchangeName: exchangeName,
	}, nil
}

// Publish sends an email message to the queue
func (p *EmailPublisher) Publish(ctx context.Context, msg EmailMessage) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal email message: %w", err)
	}

	return p.channel.PublishWithContext(
		ctx,
		p.exchangeName,   // exchange
		string(msg.Type), // routing key
		false,            // mandatory
		false,            // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent, // Survive broker restart
			Priority:     uint8(msg.Priority),
		},
	)
}

// PublishVerificationEmail publishes a verification email to the queue
func (p *EmailPublisher) PublishVerificationEmail(ctx context.Context, to, username, verifyURL, code string) error {
	return p.Publish(ctx, EmailMessage{
		Type:     EmailTypeVerification,
		To:       to,
		Subject:  "Verify Your Email - GoMarketplace",
		Priority: 5,
		Data: map[string]interface{}{
			"username":   username,
			"verify_url": verifyURL,
			"code":       code,
			"expires_in": "24 hours",
		},
	})
}

// PublishPasswordResetEmail publishes a password reset email to the queue
func (p *EmailPublisher) PublishPasswordResetEmail(ctx context.Context, to, username, resetURL string) error {
	return p.Publish(ctx, EmailMessage{
		Type:     EmailTypePasswordReset,
		To:       to,
		Subject:  "Reset Your Password - GoMarketplace",
		Priority: 8, // Higher priority for security emails
		Data: map[string]interface{}{
			"username":   username,
			"reset_url":  resetURL,
			"expires_in": "1 hour",
		},
	})
}
