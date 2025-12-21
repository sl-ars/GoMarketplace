# Monitoring with Prometheus & Grafana

This guide covers the monitoring setup for GoMarketplace using Prometheus for metrics collection and Grafana for visualization.

## Architecture

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  GoMarketplace  │────▶│   Prometheus    │────▶│    Grafana      │
│    /metrics     │     │   (Scraping)    │     │  (Dashboards)   │
└─────────────────┘     └─────────────────┘     └─────────────────┘
```

## Quick Start

### 1. Start All Services
```bash
docker-compose up -d
```

### 2. Start Your Application
```bash
go run ./cmd/app --config ./configs/.env
```

### 3. Access Dashboards

| Service | URL | Credentials |
|---------|-----|-------------|
| Grafana | http://localhost:3000 | admin / admin |
| Prometheus | http://localhost:9090 | - |
| Metrics Endpoint | http://localhost:8080/metrics | - |

## Available Metrics

### HTTP Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `http_requests_total` | Counter | Total HTTP requests by method, endpoint, status |
| `http_request_duration_seconds` | Histogram | Request duration distribution |
| `http_requests_in_flight` | Gauge | Current in-flight requests |

### Business Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `users_registered_total` | Counter | Total user registrations |
| `users_logged_in_total` | Counter | Total successful logins |
| `users_banned_total` | Counter | Total users banned |
| `orders_created_total` | Counter | Total orders created |
| `orders_completed_total` | Counter | Total orders completed |
| `orders_cancelled_total` | Counter | Total orders cancelled |
| `orders_amount_total` | Counter | Total order value (USD) |
| `payments_successful_total` | Counter | Successful payments |
| `payments_failed_total` | Counter | Failed payments |
| `payments_amount_total` | Counter | Total payment value |
| `products_total` | Gauge | Current product count |
| `offers_total` | Gauge | Current offer count |
| `product_searches_total` | Counter | Product search requests |
| `cart_items_added_total` | Counter | Items added to carts |
| `cart_checkouts_total` | Counter | Cart checkouts |
| `refunds_requested_total` | Counter | Refund requests |
| `refunds_approved_total` | Counter | Approved refunds |
| `refunds_rejected_total` | Counter | Rejected refunds |

### Infrastructure Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `db_query_duration_seconds` | Histogram | Database query latency |
| `db_connections_active` | Gauge | Active DB connections |
| `redis_operation_duration_seconds` | Histogram | Redis operation latency |
| `redis_cache_hits_total` | Counter | Redis cache hits |
| `redis_cache_misses_total` | Counter | Redis cache misses |
| `rabbitmq_messages_published_total` | Counter | Messages published |
| `rabbitmq_messages_consumed_total` | Counter | Messages consumed |
| `elasticsearch_query_duration_seconds` | Histogram | ES query latency |
| `elasticsearch_outbox_pending` | Gauge | Pending ES sync events |
| `rate_limit_exceeded_total` | Counter | Rate limit violations |
| `emails_sent_total` | Counter | Emails sent by type/status |

## Grafana Dashboards

### Pre-configured Dashboard: "GoMarketplace Overview"

The dashboard includes:

1. **HTTP Overview Row**
   - Requests/sec
   - P95 Latency
   - Error Rate (5xx)
   - In-Flight Requests
   - Requests by Endpoint (time series)
   - P95 Latency by Endpoint (time series)

2. **Business Metrics Row**
   - Total Registrations
   - Total Orders
   - Total Revenue
   - Products Count
   - Active Offers
   - Refund Requests
   - Orders per Hour
   - Payments per Hour

3. **Infrastructure Row**
   - DB Query P95 Latency
   - Redis Cache Hit/Miss Rate
   - ES Outbox Pending

## Configuration

### Environment Variables

Add these to your `configs/.env`:

```env
# Grafana (docker-compose)
GRAFANA_ADMIN_USER=admin
GRAFANA_ADMIN_PASSWORD=your-secure-password
```

### Prometheus Configuration

Edit `deployments/prometheus/prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'marketplace-api'
    static_configs:
      - targets: ['host.docker.internal:8080']  # Your app
    metrics_path: '/metrics'
    scrape_interval: 10s
```

## Using Metrics in Code

### Recording Business Events

```go
import "go-app-marketplace/pkg/metrics"

// When a user registers
metrics.UsersRegisteredTotal.Inc()

// When an order is created
metrics.OrdersCreatedTotal.Inc()
metrics.OrdersAmountTotal.Add(orderAmount)

// When a payment succeeds
metrics.PaymentsSuccessfulTotal.Inc()
metrics.PaymentsAmountTotal.Add(amount)

// When a payment fails
metrics.PaymentsFailedTotal.Inc()
```

### Recording Database Operations

```go
start := time.Now()
// ... execute query ...
metrics.RecordDBQuery("get_user", time.Since(start).Seconds())
```

### Recording Cache Operations

```go
start := time.Now()
val, err := redis.Get(ctx, key)
hit := err == nil
metrics.RecordRedisOperation("get", time.Since(start).Seconds(), hit)
```

## Alerting (Optional)

Add alert rules to Prometheus by creating `deployments/prometheus/alerts.yml`:

```yaml
groups:
  - name: marketplace
    rules:
      - alert: HighErrorRate
        expr: sum(rate(http_requests_total{status_code=~"5.."}[5m])) / sum(rate(http_requests_total[5m])) > 0.05
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High error rate detected"
          description: "Error rate is above 5%"

      - alert: HighLatency
        expr: histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le)) > 1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High latency detected"
          description: "P95 latency is above 1 second"

      - alert: ESOutboxBacklog
        expr: elasticsearch_outbox_pending > 100
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "ES outbox backlog"
          description: "More than 100 pending events in ES outbox"
```

## Useful Prometheus Queries

### Request Rate by Endpoint
```promql
sum(rate(http_requests_total[5m])) by (endpoint)
```

### P95 Latency
```promql
histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le))
```

### Error Rate (%)
```promql
sum(rate(http_requests_total{status_code=~"5.."}[5m])) / sum(rate(http_requests_total[5m])) * 100
```

### Orders per Hour
```promql
increase(orders_created_total[1h])
```

### Cache Hit Rate (%)
```promql
rate(redis_cache_hits_total[5m]) / (rate(redis_cache_hits_total[5m]) + rate(redis_cache_misses_total[5m])) * 100
```

### Revenue per Hour
```promql
increase(payments_amount_total[1h])
```

## Troubleshooting

### Metrics Not Showing in Prometheus

1. Check if the app is running and `/metrics` returns data:
```bash
curl http://localhost:8080/metrics
```

2. Check Prometheus targets:
   - Go to http://localhost:9090/targets
   - Ensure `marketplace-api` shows as "UP"

3. If running app outside Docker:
   - Update `prometheus.yml` target to use `host.docker.internal:8080`

### Grafana Can't Connect to Prometheus

1. Verify Prometheus is running:
```bash
docker-compose ps prometheus
```

2. Check Grafana data source configuration:
   - URL should be `http://prometheus:9090` (not localhost)

### High Cardinality Issues

If you see warnings about high cardinality:
- Check that endpoint normalization is working (replacing IDs with `{id}`)
- Review custom labels added to metrics

## Production Recommendations

1. **Enable authentication** for `/metrics` endpoint in production
2. **Use persistent storage** for Prometheus data (already configured)
3. **Set up alerting** via Alertmanager
4. **Add more dashboards** for specific services (Redis, RabbitMQ, PostgreSQL)
5. **Configure retention** in Prometheus for disk space management
6. **Use Grafana Cloud** for managed Grafana in production

