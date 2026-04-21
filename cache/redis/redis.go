package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/toweray/gopkg/cache"
)

// Option configures a redis.Options.
type Option func(*redis.Options)

// WithAddr sets the address of the Redis server.
func WithAddr(addr string) Option {
	return func(opts *redis.Options) {
		opts.Addr = addr
	}
}

// WithPassword sets the password for the Redis server.
func WithPassword(password string) Option {
	return func(opts *redis.Options) {
		opts.Password = password
	}
}

// WithDB sets the database number for the Redis server.
func WithDB(db int) Option {
	return func(opts *redis.Options) {
		opts.DB = db
	}
}

// Redis wraps a redis.Client and implements cache.Cache.
type Redis struct {
	client *redis.Client
}

// New creates a new Redis client, applies options, and pings the server.
func New(ctx context.Context, opts ...Option) (*Redis, error) {
	options := &redis.Options{}
	for _, opt := range opts {
		opt(options)
	}
	client := redis.NewClient(options)
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis server: %w", err)
	}
	return &Redis{client: client}, nil
}

// Get retrieves the value associated with the given key.
// It returns ErrNotFound if the key does not exist in the cache.
func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", cache.ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("redis get %q: %w", key, err)
	}
	return val, nil
}

// Set stores the value associated with the given key in the cache.
func (r *Redis) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	if err := r.client.Set(ctx, key, value, ttl).Err(); err != nil {
		return fmt.Errorf("redis set %q: %w", key, err)
	}
	return nil
}

// Delete removes the value associated with the given key from the cache.
func (r *Redis) Delete(ctx context.Context, key string) error {
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("redis delete %q: %w", key, err)
	}
	return nil
}

// Exists checks if the given key exists in the cache.
func (r *Redis) Exists(ctx context.Context, key string) (bool, error) {
	val, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("redis exists %q: %w", key, err)
	}
	return val > 0, nil
}

// Close closes the Redis client connection.
func (r *Redis) Close() error {
	return r.client.Close()
}
