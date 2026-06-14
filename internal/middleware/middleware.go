package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// RequestIDHeader is the header key used for correlation IDs.
const RequestIDHeader = "X-Request-Id"

// RequestID ensures every request/response carries a correlation ID. If the
// client supplies one it is preserved, otherwise a new UUID is generated.
func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		rid := c.Get(RequestIDHeader)
		if rid == "" {
			rid = uuid.NewString()
		}
		c.Locals("requestid", rid)
		c.Set(RequestIDHeader, rid)
		return c.Next()
	}
}

// RequestLogger logs method, path, status and duration for each request using
// the structured Zap logger, tagging entries with the request ID.
func RequestLogger(log *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		duration := time.Since(start)

		rid, _ := c.Locals("requestid").(string)
		log.Info("request handled",
			zap.String("request_id", rid),
			zap.String("method", c.Method()),
			zap.String("path", c.OriginalURL()),
			zap.Int("status", c.Response().StatusCode()),
			zap.Duration("duration", duration),
			zap.String("ip", c.IP()),
		)
		return err
	}
}
