package net

import (
	"strings"

	"github.com/gofiber/fiber/v3"
)

// RealIP returns the real client IP from the request.
// It checks Cloudflare and proxy headers before falling back to the remote address.
func RealIP(c fiber.Ctx) string {
	if ip := c.Get("X-Real-IP"); ip != "" {
		return ip
	}
	if ip := c.Get("X-Forwarded-For"); ip != "" {
		return strings.SplitN(ip, ",", 2)[0]
	}
	return c.IP()
}
