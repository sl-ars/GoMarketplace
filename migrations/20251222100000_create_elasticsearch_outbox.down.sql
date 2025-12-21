-- Drop outbox table
DROP TABLE IF EXISTS elasticsearch_outbox;
DROP INDEX IF EXISTS idx_elasticsearch_outbox_unprocessed;
DROP INDEX IF EXISTS idx_elasticsearch_outbox_processed;
