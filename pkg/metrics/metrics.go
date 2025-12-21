package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP Metrics
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status_code"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"method", "endpoint"},
	)

	HTTPRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Number of HTTP requests currently being processed",
		},
	)

	// Business Metrics - Users
	UsersRegisteredTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "users_registered_total",
			Help: "Total number of user registrations",
		},
	)

	UsersLoggedInTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "users_logged_in_total",
			Help: "Total number of successful logins",
		},
	)

	UsersActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "users_active",
			Help: "Number of active users (approximate based on recent activity)",
		},
	)

	UsersBannedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "users_banned_total",
			Help: "Total number of users banned",
		},
	)

	// Business Metrics - Orders
	OrdersCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "orders_created_total",
			Help: "Total number of orders created",
		},
	)

	OrdersCompletedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "orders_completed_total",
			Help: "Total number of orders completed",
		},
	)

	OrdersCancelledTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "orders_cancelled_total",
			Help: "Total number of orders cancelled",
		},
	)

	OrdersAmountTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "orders_amount_total",
			Help: "Total monetary amount of all orders",
		},
	)

	OrdersProcessingDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "orders_processing_duration_seconds",
			Help:    "Time to process an order",
			Buckets: []float64{.1, .5, 1, 2, 5, 10, 30, 60},
		},
	)

	// Business Metrics - Products & Offers
	ProductsTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "products_total",
			Help: "Total number of products in the system",
		},
	)

	OffersTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "offers_total",
			Help: "Total number of active offers",
		},
	)

	ProductSearchesTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "product_searches_total",
			Help: "Total number of product searches",
		},
	)

	// Business Metrics - Cart
	CartItemsAdded = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "cart_items_added_total",
			Help: "Total number of items added to carts",
		},
	)

	CartCheckoutsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "cart_checkouts_total",
			Help: "Total number of cart checkouts",
		},
	)

	CartAbandonedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "cart_abandoned_total",
			Help: "Total number of abandoned carts",
		},
	)

	// Business Metrics - Payments
	PaymentsSuccessfulTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "payments_successful_total",
			Help: "Total number of successful payments",
		},
	)

	PaymentsFailedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "payments_failed_total",
			Help: "Total number of failed payments",
		},
	)

	PaymentsAmountTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "payments_amount_total",
			Help: "Total monetary amount of successful payments",
		},
	)

	// Business Metrics - Refunds
	RefundsRequestedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "refunds_requested_total",
			Help: "Total number of refund requests",
		},
	)

	RefundsApprovedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "refunds_approved_total",
			Help: "Total number of approved refunds",
		},
	)

	RefundsRejectedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "refunds_rejected_total",
			Help: "Total number of rejected refunds",
		},
	)

	// Infrastructure Metrics - Database
	DBQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Duration of database queries",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
		},
		[]string{"operation"},
	)

	DBConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_active",
			Help: "Number of active database connections",
		},
	)

	// Infrastructure Metrics - Redis
	RedisOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "redis_operation_duration_seconds",
			Help:    "Duration of Redis operations",
			Buckets: []float64{.0001, .0005, .001, .005, .01, .025, .05},
		},
		[]string{"operation"},
	)

	RedisCacheHits = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "redis_cache_hits_total",
			Help: "Total number of Redis cache hits",
		},
	)

	RedisCacheMisses = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "redis_cache_misses_total",
			Help: "Total number of Redis cache misses",
		},
	)

	// Infrastructure Metrics - RabbitMQ
	RabbitMQMessagesPublished = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rabbitmq_messages_published_total",
			Help: "Total number of messages published to RabbitMQ",
		},
		[]string{"exchange", "routing_key"},
	)

	RabbitMQMessagesConsumed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rabbitmq_messages_consumed_total",
			Help: "Total number of messages consumed from RabbitMQ",
		},
		[]string{"queue"},
	)

	// Infrastructure Metrics - Elasticsearch
	ElasticsearchQueryDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "elasticsearch_query_duration_seconds",
			Help:    "Duration of Elasticsearch queries",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1},
		},
	)

	ElasticsearchOutboxPending = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "elasticsearch_outbox_pending",
			Help: "Number of pending events in the ES outbox",
		},
	)

	// Rate Limiting Metrics
	RateLimitExceededTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rate_limit_exceeded_total",
			Help: "Total number of rate limit exceeded events",
		},
		[]string{"endpoint"},
	)

	// Email Metrics
	EmailsSentTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "emails_sent_total",
			Help: "Total number of emails sent",
		},
		[]string{"type", "status"},
	)
)

// RecordHTTPRequest records an HTTP request metric
func RecordHTTPRequest(method, endpoint, statusCode string, duration float64) {
	HTTPRequestsTotal.WithLabelValues(method, endpoint, statusCode).Inc()
	HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(duration)
}

// RecordDBQuery records a database query metric
func RecordDBQuery(operation string, duration float64) {
	DBQueryDuration.WithLabelValues(operation).Observe(duration)
}

// RecordRedisOperation records a Redis operation metric
func RecordRedisOperation(operation string, duration float64, hit bool) {
	RedisOperationDuration.WithLabelValues(operation).Observe(duration)
	if hit {
		RedisCacheHits.Inc()
	} else {
		RedisCacheMisses.Inc()
	}
}

// RecordRabbitMQPublish records a RabbitMQ publish event
func RecordRabbitMQPublish(exchange, routingKey string) {
	RabbitMQMessagesPublished.WithLabelValues(exchange, routingKey).Inc()
}

// RecordRabbitMQConsume records a RabbitMQ consume event
func RecordRabbitMQConsume(queue string) {
	RabbitMQMessagesConsumed.WithLabelValues(queue).Inc()
}

// RecordRateLimitExceeded records a rate limit exceeded event
func RecordRateLimitExceeded(endpoint string) {
	RateLimitExceededTotal.WithLabelValues(endpoint).Inc()
}

// RecordEmailSent records an email sent event
func RecordEmailSent(emailType string, success bool) {
	status := "success"
	if !success {
		status = "failed"
	}
	EmailsSentTotal.WithLabelValues(emailType, status).Inc()
}
