package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	// bcryptCost is deliberately slightly above the minimum to balance
	// security and latency. Increase to 13+ for higher security environments.
	bcryptCost = 12
)

// HashPassword returns a bcrypt hash of the plaintext password.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("hashing password: %w", err)
	}
	return string(hash), nil
}

// CheckPassword compares a plaintext password against a bcrypt hash.
// Returns nil on success, or an error on mismatch/failure.
func CheckPassword(password, hash string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return fmt.Errorf("password mismatch: %w", err)
	}
	return nil
}
