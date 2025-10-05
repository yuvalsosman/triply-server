package middleware

import (
	"strings"
	"triply-server/internal/utils"

	"github.com/gofiber/fiber/v2"
)

// AuthMiddleware handles JWT authentication
type AuthMiddleware struct {
	jwtSecret string
}

// NewAuthMiddleware creates a new auth middleware instance
func NewAuthMiddleware(jwtSecret string) *AuthMiddleware {
	return &AuthMiddleware{jwtSecret: jwtSecret}
}

// RequireAuth validates JWT token and sets user ID in context
func (m *AuthMiddleware) RequireAuth(c *fiber.Ctx) error {
	// Try to get token from cookie first
	token := c.Cookies("triply_token")

	// If not in cookie, try Authorization header
	if token == "" {
		authHeader := c.Get("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	// For backward compatibility, check for simple user cookie
	if token == "" {
		userID := c.Cookies("triply_user")
		if userID != "" {
			// Simple auth mode (for development)
			c.Locals("userId", userID)
			return c.Next()
		}
	}

	if token == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "authentication required")
	}

	// Validate JWT token
	claims, err := utils.ValidateJWT(token, m.jwtSecret)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid token")
	}

	// Set user ID in context
	c.Locals("userId", claims.UserID)
	c.Locals("userEmail", claims.Email)

	return c.Next()
}

// OptionalAuth validates JWT token if present but doesn't require it
// Also checks for shadow user ID for unauthenticated users
func (m *AuthMiddleware) OptionalAuth(c *fiber.Ctx) error {
	token := c.Cookies("triply_token")

	if token == "" {
		authHeader := c.Get("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	// Backward compatibility
	if token == "" {
		userID := c.Cookies("triply_user")
		if userID != "" {
			c.Locals("userId", userID)
		}

		// Check for shadow user ID (for unauthenticated users)
		shadowUserID := c.Cookies("triply_shadow_user_id")
		if shadowUserID == "" {
			shadowUserID = c.Get("X-Shadow-User-ID")
		}
		if shadowUserID != "" {
			c.Locals("shadowUserId", shadowUserID)
		}

		return c.Next()
	}

	// Validate token if present
	claims, err := utils.ValidateJWT(token, m.jwtSecret)
	if err == nil {
		c.Locals("userId", claims.UserID)
		c.Locals("userEmail", claims.Email)
	} else {
		// If token is invalid, still check for shadow user ID
		shadowUserID := c.Cookies("triply_shadow_user_id")
		if shadowUserID == "" {
			shadowUserID = c.Get("X-Shadow-User-ID")
		}
		if shadowUserID != "" {
			c.Locals("shadowUserId", shadowUserID)
		}
	}

	return c.Next()
}

// GetUserID extracts user ID from context
func GetUserID(c *fiber.Ctx) string {
	if userID, ok := c.Locals("userId").(string); ok {
		return userID
	}
	return ""
}

// GetShadowUserID extracts shadow user ID from context
func GetShadowUserID(c *fiber.Ctx) string {
	if shadowUserID, ok := c.Locals("shadowUserId").(string); ok {
		return shadowUserID
	}
	return ""
}
