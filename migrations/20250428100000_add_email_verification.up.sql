-- Add email verification and password reset fields to users table

ALTER TABLE users 
ADD COLUMN IF NOT EXISTS email_verified BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS email_verified_at TIMESTAMP,
ADD COLUMN IF NOT EXISTS verification_token_hash VARCHAR(64),
ADD COLUMN IF NOT EXISTS verification_token_expires_at TIMESTAMP,
ADD COLUMN IF NOT EXISTS verification_code VARCHAR(6),
ADD COLUMN IF NOT EXISTS reset_token_hash VARCHAR(64),
ADD COLUMN IF NOT EXISTS reset_token_expires_at TIMESTAMP;

-- Index for token lookups
CREATE INDEX IF NOT EXISTS idx_users_verification_token ON users(verification_token_hash) WHERE verification_token_hash IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_reset_token ON users(reset_token_hash) WHERE reset_token_hash IS NOT NULL;

-- Comment on new columns
COMMENT ON COLUMN users.email_verified IS 'Whether the user has verified their email address';
COMMENT ON COLUMN users.email_verified_at IS 'When the email was verified';
COMMENT ON COLUMN users.verification_token_hash IS 'SHA-256 hash of email verification token';
COMMENT ON COLUMN users.verification_token_expires_at IS 'When the verification token expires';
COMMENT ON COLUMN users.verification_code IS 'Short numeric code for email verification';
COMMENT ON COLUMN users.reset_token_hash IS 'SHA-256 hash of password reset token';
COMMENT ON COLUMN users.reset_token_expires_at IS 'When the reset token expires';

