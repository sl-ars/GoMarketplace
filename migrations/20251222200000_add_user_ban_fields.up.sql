-- Add ban-related fields to users table
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS is_banned BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS banned_at TIMESTAMP WITH TIME ZONE,
    ADD COLUMN IF NOT EXISTS ban_reason TEXT,
    ADD COLUMN IF NOT EXISTS banned_by_id BIGINT REFERENCES users(id);

-- Index for quickly finding banned users
CREATE INDEX IF NOT EXISTS idx_users_is_banned ON users(is_banned) WHERE is_banned = TRUE;

