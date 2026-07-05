package pagination

import "strconv"

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
