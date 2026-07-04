package pagination

import (
	"strconv"

	"github.com/google/uuid"
)

const (
	DefaultCount = 20
	MaxCount     = 100
)

type options struct {
	defaultCount int
	maxCount     int
}

// Option configures pagination behavior.
type Option func(*options)

// WithDefaultCount overrides the default page size used by ParseCount.
func WithDefaultCount(count int) Option {
	return func(o *options) {
		if count > 0 {
			o.defaultCount = count
		}
	}
}

// WithMaxCount overrides the maximum page size used by ParseCount.
func WithMaxCount(count int) Option {
	return func(o *options) {
		if count > 0 {
			o.maxCount = count
		}
	}
}

func applyOptions(opts []Option) options {
	cfg := options{
		defaultCount: DefaultCount,
		maxCount:     MaxCount,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	return cfg
}

// Page represents a page of items with pagination cursors.
type Page[T any] struct {
	Items   []T
	HasMore bool
	Prev    []uuid.UUID
	Next    []uuid.UUID
}

// ParseCount returns a validated page size from raw input.
func ParseCount(raw string, opts ...Option) int {
	cfg := applyOptions(opts)
	if raw == "" {
		return cfg.defaultCount
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 {
		return cfg.defaultCount
	}
	if n > cfg.maxCount {
		return cfg.maxCount
	}
	return n
}

// BuildPage builds a page from the provided items.
func BuildPage[T any](items []T, hasMore bool, cursorProvided bool, idOf func(T) uuid.UUID) Page[T] {
	page := Page[T]{
		Items:   items,
		HasMore: hasMore,
		Prev:    []uuid.UUID{},
		Next:    []uuid.UUID{},
	}

	if len(items) == 0 {
		return page
	}

	if cursorProvided {
		page.Prev = []uuid.UUID{idOf(items[0])}
	}

	if hasMore {
		page.Next = []uuid.UUID{idOf(items[len(items)-1])}
	}

	return page
}
