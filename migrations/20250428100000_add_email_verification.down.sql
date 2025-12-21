-- Remove email verification and password reset fields from users table

DROP INDEX IF EXISTS idx_users_verification_token;
DROP INDEX IF EXISTS idx_users_reset_token;

ALTER TABLE users 
DROP COLUMN IF EXISTS email_verified,
DROP COLUMN IF EXISTS email_verified_at,
DROP COLUMN IF EXISTS verification_token_hash,
DROP COLUMN IF EXISTS verification_token_expires_at,
DROP COLUMN IF EXISTS verification_code,
DROP COLUMN IF EXISTS reset_token_hash,
DROP COLUMN IF EXISTS reset_token_expires_at;

