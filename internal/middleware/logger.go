package middleware

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Logger logs HTTP requests
func Logger(c *fiber.Ctx) error {
	start := time.Now()

	// Process request
	err := c.Next()

	// Log after response
	duration := time.Since(start)
	status := c.Response().StatusCode()

	log.Printf("[%s] %s %s - %d - %v",
		c.Method(),
		c.Path(),
		c.IP(),
		status,
		duration,
	)

	return err
}
