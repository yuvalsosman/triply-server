package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// NewCORS creates a configured CORS middleware
func NewCORS(origin string) fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins:     origin,
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Content-Type,Authorization,X-Shadow-User-ID",
		AllowCredentials: true,
	})
}
