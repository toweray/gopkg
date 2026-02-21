package middleware

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/toweray/gopkg/net"
	"go.uber.org/zap"
)

const requestIDKey = "request_id"

// StatusResolver resolves an HTTP status code from an error.
// Returns 0 if the error is not handled by this resolver.
type StatusResolver func(err error) int

// Logger returns a Fiber middleware that logs each request
// and injects a unique request ID into the context.
func Logger(logger *zap.Logger, resolvers ...StatusResolver) fiber.Handler {
	return func(c fiber.Ctx) error {
		requestID := uuid.New().String()
		c.Locals(requestIDKey, requestID)
		c.Set("X-Request-ID", requestID)

		start := time.Now()
		err := c.Next()
		duration := time.Since(start)

		status := c.Response().StatusCode()
		if err != nil {
			status = resolveStatus(err, resolvers)
		}

		fields := []zap.Field{
			zap.String("request_id", requestID),
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Int("status", status),
			zap.Duration("duration", duration),
			zap.String("ip", net.RealIP(c)),
		}

		switch {
		case status >= 500:
			logger.Error("request failed", fields...)
		case status >= 400:
			logger.Warn("request warning", fields...)
		default:
			logger.Info("request completed", fields...)
		}

		return err
	}
}

// RequestID extracts the request ID injected by Logger middleware.
func RequestID(c fiber.Ctx) string {
	id, _ := c.Locals(requestIDKey).(string)
	return id
}

// resolveStatus tries each resolver in order, falling back to fiber.Error
// and then 500 if none match.
func resolveStatus(err error, resolvers []StatusResolver) int {
	for _, r := range resolvers {
		if code := r(err); code != 0 {
			return code
		}
	}

	var fe *fiber.Error
	if errors.As(err, &fe) {
		return fe.Code
	}

	return fiber.StatusInternalServerError
}
