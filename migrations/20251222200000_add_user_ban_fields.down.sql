-- Remove ban-related fields from users table
DROP INDEX IF EXISTS idx_users_is_banned;

ALTER TABLE users
DROP COLUMN IF EXISTS banned_by_id,
DROP COLUMN IF EXISTS ban_reason,
DROP COLUMN IF EXISTS banned_at,
DROP COLUMN IF EXISTS is_banned;

