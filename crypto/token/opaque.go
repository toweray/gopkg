package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// Generate creates a cryptographically random token of the given byte length.
// Returns the plaintext token (base64-encoded) and its SHA-256 hex hash.
func Generate(length int) (plain, hash string, err error) {
	b := make([]byte, length)
	if _, err = rand.Read(b); err != nil {
		return "", "", fmt.Errorf("failed to generate token: %w", err)
	}

	plain = base64.RawURLEncoding.EncodeToString(b)
	hash = Hash(plain)

	return plain, hash, nil
}

// Hash returns the SHA-256 hex digest of token.
func Hash(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
