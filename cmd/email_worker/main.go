package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go"

	"go-app-marketplace/internal/app/config"
	"go-app-marketplace/internal/messagebus"
	"go-app-marketplace/pkg/email"
	"go-app-marketplace/pkg/logger"
)

func main() {
	configFile := flag.String("config", "./configs/.env", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.NewConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	appLogger := logger.New(cfg.Logger)
	appLogger.Info("Starting email worker")

	// Check if email is enabled
	if !cfg.Email.Enabled {
		appLogger.Fatal("Email is disabled. Set EMAIL_ENABLED=true to run the email worker.")
	}

	// Initialize email sender
	emailSender := email.NewSMTPSender(email.Config{
		Host:     cfg.Email.Host,
		Port:     cfg.Email.Port,
		Username: cfg.Email.Username,
		Password: cfg.Email.Password,
		From:     cfg.Email.From,
		FromName: cfg.Email.FromName,
		UseTLS:   cfg.Email.UseTLS,
	})
	appLogger.Info("Email sender initialized")

	// Connect to RabbitMQ
	conn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		appLogger.WithError(err).Fatal("Failed to connect to RabbitMQ")
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		appLogger.WithError(err).Fatal("Failed to open channel")
	}
	defer ch.Close()

	// Set QoS (prefetch count)
	err = ch.Qos(
		5,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		appLogger.WithError(err).Fatal("Failed to set QoS")
	}

	// Consume messages from the emails queue
	msgs, err := ch.Consume(
		"emails", // queue
		"",       // consumer
		false,    // auto-ack (manual ack for reliability)
		false,    // exclusive
		false,    // no-local
		false,    // no-wait
		nil,      // args
	)
	if err != nil {
		appLogger.WithError(err).Fatal("Failed to register consumer")
	}

	// Handle graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		appLogger.Info("Shutting down email worker...")
		cancel()
	}()

	appLogger.Info("Email worker is ready. Waiting for messages...")

	// Process messages
	for {
		select {
		case <-ctx.Done():
			appLogger.Info("Email worker stopped")
			return
		case msg, ok := <-msgs:
			if !ok {
				appLogger.Warn("Channel closed, reconnecting...")
				return
			}

			processEmail(appLogger, emailSender, msg)
		}
	}
}

func processEmail(log *logger.Logger, sender *email.SMTPSender, msg amqp.Delivery) {
	var emailMsg messagebus.EmailMessage
	if err := json.Unmarshal(msg.Body, &emailMsg); err != nil {
		log.WithError(err).Error("Failed to unmarshal email message")
		// Reject and don't requeue malformed messages
		msg.Reject(false)
		return
	}

	log.WithField("type", emailMsg.Type).WithField("to", emailMsg.To).Info("Processing email")

	var err error

	switch emailMsg.Type {
	case messagebus.EmailTypeVerification:
		err = sendVerificationEmail(sender, emailMsg)
	case messagebus.EmailTypePasswordReset:
		err = sendPasswordResetEmail(sender, emailMsg)
	default:
		log.WithField("type", emailMsg.Type).Warn("Unknown email type")
		msg.Reject(false)
		return
	}

	if err != nil {
		log.WithError(err).WithField("to", emailMsg.To).Error("Failed to send email")
		// Requeue for retry
		msg.Nack(false, true)
		return
	}

	log.WithField("to", emailMsg.To).Info("Email sent successfully")
	msg.Ack(false)
}

func sendVerificationEmail(sender *email.SMTPSender, msg messagebus.EmailMessage) error {
	data := email.VerifyEmailData{
		Username:  getString(msg.Data, "username"),
		VerifyURL: getString(msg.Data, "verify_url"),
		Code:      getString(msg.Data, "code"),
		ExpiresIn: getString(msg.Data, "expires_in"),
	}
	return sender.SendTemplate(msg.To, msg.Subject, "verify_email", data)
}

func sendPasswordResetEmail(sender *email.SMTPSender, msg messagebus.EmailMessage) error {
	data := email.ResetPasswordData{
		Username:  getString(msg.Data, "username"),
		ResetURL:  getString(msg.Data, "reset_url"),
		ExpiresIn: getString(msg.Data, "expires_in"),
	}
	return sender.SendTemplate(msg.To, msg.Subject, "reset_password", data)
}

func getString(data map[string]interface{}, key string) string {
	if v, ok := data[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
