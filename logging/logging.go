package logging

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Options holds logging configuration.
type Options struct {
	level       string
	development bool
	filePath    string
}

// Option configures a Logger.
type Option func(*Options)

// WithLevel sets the minimum log level ("debug", "info", "warn", "error").
func WithLevel(level string) Option {
	return func(o *Options) { o.level = level }
}

// WithDevelopment enables human-readable console output.
func WithDevelopment() Option {
	return func(o *Options) { o.development = true }
}

// WithFile appends log output to the given file path.
func WithFile(path string) Option {
	return func(o *Options) { o.filePath = path }
}

// New builds a *zap.Logger with the given options.
// Defaults: JSON encoding, info level, stdout output.
func New(opts ...Option) (*zap.Logger, error) {
	o := &Options{level: "info"}
	for _, opt := range opts {
		opt(o)
	}

	var level zapcore.Level
	if err := level.UnmarshalText([]byte(o.level)); err != nil {
		return nil, fmt.Errorf("invalid log level %q: %w", o.level, err)
	}

	cfg := zap.Config{
		Level:            zap.NewAtomicLevelAt(level),
		Development:      o.development,
		Encoding:         "json",
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig:    zap.NewProductionEncoderConfig(),
	}

	if o.development {
		cfg.Encoding = "console"
		cfg.EncoderConfig = zap.NewDevelopmentEncoderConfig()
	}

	if o.filePath != "" {
		cfg.OutputPaths = append(cfg.OutputPaths, o.filePath)
	}

	logger, err := cfg.Build(
		// Show stacktrace only for Error and above, not for Warn
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	return logger, nil
}

// NewNop returns a no-op logger suitable for use in tests.
func NewNop() *zap.Logger {
	return zap.NewNop()
}
