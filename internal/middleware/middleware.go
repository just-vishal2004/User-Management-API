package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/vishalkumar/ainyx-backend/internal/logger"
)

// RequestID is a middleware that injects a unique request ID into
// every request and response. If the client sends an X-Request-ID
// header, we use that value — this allows distributed systems to
// trace a request across multiple services. Otherwise we generate
// a new UUID.
func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check if the client already sent a request ID.
		requestID := c.Get("X-Request-ID")

		// If not, generate a new UUID v4.
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Store the request ID in Fiber's context locals.
		// This makes it available to handlers and other middleware
		// via c.Locals("requestID").
		c.Locals("requestID", requestID)

		// Set the request ID in the response header so clients
		// can correlate their request with server logs.
		c.Set("X-Request-ID", requestID)

		// Call the next middleware or handler in the chain.
		return c.Next()
	}
}

// Logger is a middleware that logs every request using Uber Zap.
// It records the method, path, status code, duration, and request ID.
// This runs after the handler completes so it can capture the status code.
func Logger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Record the time before the handler runs.
		start := time.Now()

		// Call the next middleware or handler.
		// Any error returned is captured here.
		err := c.Next()

		// Calculate how long the handler took.
		duration := time.Since(start)

		// Retrieve the request ID we set in RequestID middleware.
		requestID, _ := c.Locals("requestID").(string)

		// Get the HTTP status code from the response.
		status := c.Response().StatusCode()

		// Choose log level based on status code:
		// 5xx = server errors → ERROR level
		// 4xx = client errors → WARN level
		// 2xx/3xx = success  → INFO level
		if status >= 500 {
			logger.Log.Error("request completed",
				zap.String("request_id", requestID),
				zap.String("method", c.Method()),
				zap.String("path", c.Path()),
				zap.Int("status", status),
				zap.Duration("duration", duration),
			)
		} else if status >= 400 {
			logger.Log.Warn("request completed",
				zap.String("request_id", requestID),
				zap.String("method", c.Method()),
				zap.String("path", c.Path()),
				zap.Int("status", status),
				zap.Duration("duration", duration),
			)
		} else {
			logger.Log.Info("request completed",
				zap.String("request_id", requestID),
				zap.String("method", c.Method()),
				zap.String("path", c.Path()),
				zap.Int("status", status),
				zap.Duration("duration", duration),
			)
		}

		return err
	}
}
