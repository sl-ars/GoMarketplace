package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/jmoiron/sqlx"
)

type ElasticsearchOutboxEvent struct {
	ID            int64           `db:"id"`
	AggregateType string          `db:"aggregate_type"`
	AggregateID   int64           `db:"aggregate_id"`
	EventType     string          `db:"event_type"`
	Payload       json.RawMessage `db:"payload"`
	CreatedAt     time.Time       `db:"created_at"`
	ProcessedAt   sql.NullTime    `db:"processed_at"`
	RetryCount    int             `db:"retry_count"`
	ErrorMessage  sql.NullString  `db:"error_message"`
}

type ElasticsearchOutboxRepository struct {
	db *sqlx.DB
}

func NewElasticsearchOutboxRepository(db *sqlx.DB) *ElasticsearchOutboxRepository {
	return &ElasticsearchOutboxRepository{db: db}
}

// InsertEvent inserts a new event into the outbox (transactional or non-transactional)
func (r *ElasticsearchOutboxRepository) InsertEvent(ctx context.Context, tx *sqlx.Tx, aggregateType string, aggregateID int64, eventType string, payload interface{}) error {
	var payloadJSON []byte
	if payload != nil {
		var err error
		payloadJSON, err = json.Marshal(payload)
		if err != nil {
			return err
		}
	}

	query := `
		INSERT INTO elasticsearch_outbox (aggregate_type, aggregate_id, event_type, payload)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (aggregate_type, aggregate_id, event_type)
		DO UPDATE SET
			payload = EXCLUDED.payload,
			created_at = NOW(),
			processed_at = NULL,
			retry_count = 0,
			error_message = NULL
	`

	var err error
	if tx != nil {
		_, err = tx.ExecContext(ctx, query, aggregateType, aggregateID, eventType, payloadJSON)
	} else {
		_, err = r.db.ExecContext(ctx, query, aggregateType, aggregateID, eventType, payloadJSON)
	}

	return err
}

// GetUnprocessedEvents retrieves unprocessed events ordered by creation time
func (r *ElasticsearchOutboxRepository) GetUnprocessedEvents(ctx context.Context, limit int) ([]*ElasticsearchOutboxEvent, error) {
	var events []*ElasticsearchOutboxEvent
	err := r.db.SelectContext(ctx, &events, `
		SELECT id, aggregate_type, aggregate_id, event_type, payload, created_at, processed_at, retry_count, error_message
		FROM elasticsearch_outbox
		WHERE processed_at IS NULL
		ORDER BY created_at ASC
		LIMIT $1
	`, limit)
	return events, err
}

// MarkEventProcessed marks an event as successfully processed
func (r *ElasticsearchOutboxRepository) MarkEventProcessed(ctx context.Context, eventID int64) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE elasticsearch_outbox
		SET processed_at = NOW()
		WHERE id = $1
	`, eventID)
	return err
}

// MarkEventFailed marks an event as failed and increments retry count
func (r *ElasticsearchOutboxRepository) MarkEventFailed(ctx context.Context, eventID int64, errorMessage string, maxRetries int) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE elasticsearch_outbox
		SET retry_count = retry_count + 1,
		    error_message = $2
		WHERE id = $1 AND retry_count < $3
	`, eventID, errorMessage, maxRetries)
	return err
}

// DeleteProcessedEvents removes old processed events (cleanup)
func (r *ElasticsearchOutboxRepository) DeleteProcessedEvents(ctx context.Context, olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM elasticsearch_outbox
		WHERE processed_at IS NOT NULL AND processed_at < $1
	`, cutoff)
	return err
}

// GetEventStats returns statistics about the outbox
func (r *ElasticsearchOutboxRepository) GetEventStats(ctx context.Context) (map[string]int, error) {
	stats := make(map[string]int)

	// Total events
	var total int
	err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM elasticsearch_outbox")
	if err != nil {
		return nil, err
	}
	stats["total"] = total

	// Unprocessed events
	var unprocessed int
	err = r.db.GetContext(ctx, &unprocessed, "SELECT COUNT(*) FROM elasticsearch_outbox WHERE processed_at IS NULL")
	if err != nil {
		return nil, err
	}
	stats["unprocessed"] = unprocessed

	// Failed events (high retry count)
	var failed int
	err = r.db.GetContext(ctx, &failed, "SELECT COUNT(*) FROM elasticsearch_outbox WHERE retry_count >= 3 AND processed_at IS NULL")
	if err != nil {
		return nil, err
	}
	stats["failed"] = failed

	return stats, nil
}
