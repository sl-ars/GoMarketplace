-- Create outbox table for Elasticsearch synchronization
CREATE TABLE IF NOT EXISTS elasticsearch_outbox (
    id BIGSERIAL PRIMARY KEY,
    aggregate_type VARCHAR(50) NOT NULL, -- "product"
    aggregate_id BIGINT NOT NULL,       -- product ID
    event_type VARCHAR(50) NOT NULL,     -- "created", "updated", "deleted"
    payload JSONB,                       -- additional data if needed
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMP,
    retry_count INTEGER NOT NULL DEFAULT 0,
    error_message TEXT,
    UNIQUE(aggregate_type, aggregate_id, event_type)
);

-- Index for efficient querying
CREATE INDEX IF NOT EXISTS idx_elasticsearch_outbox_unprocessed 
ON elasticsearch_outbox (created_at) 
WHERE processed_at IS NULL;

-- Index for cleanup
CREATE INDEX IF NOT EXISTS idx_elasticsearch_outbox_processed 
ON elasticsearch_outbox (processed_at) 
WHERE processed_at IS NOT NULL;
