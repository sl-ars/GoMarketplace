# Elasticsearch Synchronization Worker

This worker implements the **Transactional Outbox Pattern** to reliably synchronize database changes with Elasticsearch. It processes events from an outbox table to keep search indices up-to-date.

## How It Works

### Transactional Outbox Pattern

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Database  │───▶│   Outbox    │───▶│   Worker    │
│  (Changes)  │    │   Table     │    │ (Process)  │
└─────────────┘    └─────────────┘    └─────┬──────┘
                                            │
                                            ▼
                                   ┌─────────────┐
                                   │ Elasticsearch│
                                   │   (Sync)     │
                                   └─────────────┘
```

### Process Flow

1. **Application** makes database changes + inserts event into `elasticsearch_outbox` table
2. **Worker** polls the outbox table for unprocessed events
3. **Worker** updates Elasticsearch based on event type
4. **Worker** marks events as processed or failed

### Event Types

| Event Type | Trigger | Action |
|------------|---------|--------|
| `created` | Product created | Index product in ES |
| `updated` | Product/offer changed | Re-index product with current offers |
| `deleted` | Product deleted | Remove from ES index |

## Setup Instructions

### 1. Run Database Migration
```bash
# Apply the outbox table migration
migrate -path ./migrations -database "postgres://user:pass@localhost/db?sslmode=disable" up
```

### 2. Start Services
```bash
# Start all services
docker-compose up -d

# Or individually:
docker-compose up -d postgres elasticsearch
```

### 3. Run the Worker
```bash
# Start the Elasticsearch synchronization worker
go run ./cmd/elasticsearch_worker --config ./configs/.env

# Or build and run:
go build ./cmd/elasticsearch_worker
./elasticsearch_worker --config ./configs/.env
```

### 4. Start Your Main Application
```bash
# Start the main API server
go run ./cmd/app --config ./configs/.env
```

## Configuration

### Environment Variables

```env
# Elasticsearch
ES_ENABLED=true
ES_ADDRESSES=http://localhost:9200

# Worker Settings (optional, has defaults)
WORKER_BATCH_SIZE=10
WORKER_POLL_INTERVAL=5s
WORKER_MAX_RETRIES=3
```

### Worker Behavior

- **Batch Size**: Processes 10 events at a time
- **Poll Interval**: Checks for new events every 5 seconds
- **Max Retries**: Failed events are retried up to 3 times
- **Cleanup**: Removes processed events older than 7 days

## Monitoring

### Check Outbox Status
```bash
# Connect to PostgreSQL
psql "postgres://user:pass@localhost/db"

# Check outbox statistics
SELECT
    COUNT(*) as total_events,
    COUNT(*) FILTER (WHERE processed_at IS NULL) as unprocessed,
    COUNT(*) FILTER (WHERE retry_count >= 3 AND processed_at IS NULL) as failed
FROM elasticsearch_outbox;
```

### Worker Logs
```
INFO Processing outbox events count=5
INFO Successfully processed outbox event event_id=123 aggregate_type=product event_type=updated
INFO Outbox statistics total_events=150 unprocessed_events=3 failed_events=0
```

### Elasticsearch Index
```bash
# Check document count
curl "localhost:9200/products/_count"

# Search for products
curl "localhost:9200/products/_search?q=iPhone&size=5"
```

## Reliability Features

### ✅ At-Least-Once Delivery
- Events stay in outbox until successfully processed
- Failed events are retried with exponential backoff

### ✅ Idempotent Operations
- Duplicate events are handled gracefully
- Same event processed multiple times = same result

### ✅ Monitoring & Alerting
- Failed event count tracked
- Stats reported every 5 minutes
- Errors logged with context

### ✅ Automatic Cleanup
- Old processed events removed automatically
- Prevents outbox table from growing indefinitely

## Troubleshooting

### Worker Not Processing Events
```bash
# Check worker logs
tail -f /var/log/elasticsearch_worker.log

# Check database connection
psql "postgres://..." -c "SELECT COUNT(*) FROM elasticsearch_outbox WHERE processed_at IS NULL;"

# Restart worker
kill $(pgrep elasticsearch_worker)
go run ./cmd/elasticsearch_worker --config ./configs/.env
```

### Events Stuck in Failed State
```bash
# Check failed events
psql "postgres://..." -c "SELECT * FROM elasticsearch_outbox WHERE retry_count >= 3 AND processed_at IS NULL;"

# Reset retry count (for debugging)
UPDATE elasticsearch_outbox SET retry_count = 0, error_message = NULL WHERE id = ?;
```

### Elasticsearch Connection Issues
```bash
# Check ES health
curl -X GET "localhost:9200/_cluster/health?pretty"

# Check worker can connect
curl -X GET "localhost:9200/products"
```

### High Outbox Backlog
```bash
# Check unprocessed count
psql "postgres://..." -c "SELECT COUNT(*) FROM elasticsearch_outbox WHERE processed_at IS NULL;"

# Scale up: Run multiple worker instances
go run ./cmd/elasticsearch_worker --config ./configs/.env &
go run ./cmd/elasticsearch_worker --config ./configs/.env &
```

## Production Deployment

### Docker Compose
```yaml
services:
  elasticsearch-worker:
    build: .
    command: ["./elasticsearch_worker", "--config", "/app/configs/.env"]
    depends_on:
      - postgres
      - elasticsearch
    restart: unless-stopped
    environment:
      - WORKER_BATCH_SIZE=50
      - WORKER_POLL_INTERVAL=1s
```

### Kubernetes
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: elasticsearch-worker
spec:
  replicas: 2  # Multiple workers for high throughput
  template:
    spec:
      containers:
      - name: worker
        image: your-app:latest
        command: ["./elasticsearch_worker"]
        env:
        - name: WORKER_BATCH_SIZE
          value: "50"
```

### Health Checks
```bash
# Worker health endpoint (if implemented)
curl http://localhost:8081/health

# Outbox health query
SELECT CASE WHEN COUNT(*) < 1000 THEN 'healthy' ELSE 'backlog' END
FROM elasticsearch_outbox WHERE processed_at IS NULL;
```

## Architecture Benefits

| Aspect | Old Approach | Transactional Outbox |
|--------|-------------|---------------------|
| **Reliability** | Lost updates on crashes | Guaranteed delivery |
| **Consistency** | Race conditions | ACID transactions |
| **Monitoring** | No visibility | Full observability |
| **Scalability** | Single goroutine | Multiple workers |
| **Error Handling** | Silent failures | Retry with backoff |
| **Maintenance** | Hard to debug | Structured logging |

This implementation ensures **100% reliability** for Elasticsearch synchronization, making it suitable for production use! 🚀
