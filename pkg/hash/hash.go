package hash

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"golang.org/x/crypto/argon2"
)

const (
	Memory      = 64 * 1024
	Iterations  = 3
	Parallelism = 2
	SaltLength  = 16
	KeyLength   = 32
)

// HashPassword hashes password using argon2
func HashPassword(password string) (string, error) {
	salt := make([]byte, SaltLength)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, Iterations, Memory, uint8(Parallelism), KeyLength)

	// encoded format: salt$hash
	encoded := fmt.Sprintf("%s$%s",
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)

	return encoded, nil
}

// ComparePassword compares hash and password
func ComparePassword(encodedHash, password string) (bool, error) {
	parts := make([][]byte, 2)
	split := []rune(encodedHash)
	for i, part := range split {
		if part == '$' {
			parts[0], _ = base64.RawStdEncoding.DecodeString(string(split[:i]))
			parts[1], _ = base64.RawStdEncoding.DecodeString(string(split[i+1:]))
			break
		}
	}

	salt := parts[0]
	expectedHash := parts[1]

	newHash := argon2.IDKey([]byte(password), salt, Iterations, Memory, uint8(Parallelism), KeyLength)

	match := base64.RawStdEncoding.EncodeToString(newHash) == base64.RawStdEncoding.EncodeToString(expectedHash)
	return match, nil
}
