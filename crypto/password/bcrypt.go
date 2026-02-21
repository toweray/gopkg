package password

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// BcryptHasher implements Hasher using the bcrypt algorithm.
type BcryptHasher struct {
	cost int
}

// NewBcrypt returns a new BcryptHasher.
// If cost is 0, bcrypt.DefaultCost is used.
func NewBcrypt(cost int) *BcryptHasher {
	if cost == 0 {
		cost = bcrypt.DefaultCost
	}
	return &BcryptHasher{cost: cost}
}

// Hash hashes password using bcrypt.
func (h *BcryptHasher) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

// Verify checks whether password matches the bcrypt hash.
func (h *BcryptHasher) Verify(password, hash string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
