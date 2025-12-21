package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"
)

// Token represents a secure token with expiration
type Token struct {
	// PlainText is the token sent to the user (via email)
	PlainText string
	// Hash is stored in the database
	Hash string
	// ExpiresAt is when the token expires
	ExpiresAt time.Time
}

// Generate creates a new secure token
// Returns both the plain text (to send to user) and hash (to store in DB)
func Generate(ttl time.Duration) (*Token, error) {
	// Generate 32 random bytes
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Create the plain text token (URL-safe base64)
	plainText := base64.URLEncoding.EncodeToString(randomBytes)

	// Create a hash for storage
	hash := HashToken(plainText)

	return &Token{
		PlainText: plainText,
		Hash:      hash,
		ExpiresAt: time.Now().Add(ttl),
	}, nil
}

// GenerateCode creates a short numeric code (for email verification)
func GenerateCode(length int) (string, error) {
	if length <= 0 || length > 10 {
		length = 6
	}

	// Generate random bytes
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Convert to numeric string
	code := ""
	for i := 0; i < length; i++ {
		code += fmt.Sprintf("%d", bytes[i]%10)
	}

	return code, nil
}

// HashToken creates a SHA-256 hash of the token
// This is what gets stored in the database
func HashToken(plainText string) string {
	hash := sha256.Sum256([]byte(plainText))
	return hex.EncodeToString(hash[:])
}

// Validate checks if a plain text token matches a hash and hasn't expired
func Validate(plainText, hash string, expiresAt time.Time) bool {
	if time.Now().After(expiresAt) {
		return false
	}
	return HashToken(plainText) == hash
}

// Token TTL constants
const (
	EmailVerificationTTL = 24 * time.Hour
	PasswordResetTTL     = 1 * time.Hour
)
