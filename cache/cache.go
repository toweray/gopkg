package cache

import (
	"context"
	"errors"
	"time"
)

// ErrNotFound is returned when a key is not found in the cache.
var ErrNotFound = errors.New("cache: key not found")

// Cache defines the interface for a key-value cache store.
type Cache interface {
	// Get retrieves the value associated with the given key.
	// It returns ErrNotFound if the key does not exist in the cache.
	Get(ctx context.Context, key string) (string, error)

	// Set stores the value associated with the given key in the cache.
	Set(ctx context.Context, key string, value string, ttl time.Duration) error

	// Delete removes the value associated with the given key from the cache.
	Delete(ctx context.Context, key string) error

	// Exists checks if the given key exists in the cache.
	Exists(ctx context.Context, key string) (bool, error)
}
