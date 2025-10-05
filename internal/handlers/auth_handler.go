package handlers

import (
	"context"
	"encoding/json"
	"triply-server/internal/middleware"
	"triply-server/internal/service"
	"triply-server/internal/utils"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/oauth2"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authService    service.AuthService
	tripService    service.TripService
	oauthConfig    *oauth2.Config
	jwtSecret      string
	frontendOrigin string
}

// NewAuthHandler creates a new auth handler instance
func NewAuthHandler(authService service.AuthService, tripService service.TripService, oauthConfig *oauth2.Config, jwtSecret, frontendOrigin string) *AuthHandler {
	return &AuthHandler{
		authService:    authService,
		tripService:    tripService,
		oauthConfig:    oauthConfig,
		jwtSecret:      jwtSecret,
		frontendOrigin: frontendOrigin,
	}
}

// GoogleLogin handles GET /auth/google
func (h *AuthHandler) GoogleLogin(c *fiber.Ctx) error {
	if h.oauthConfig == nil || h.oauthConfig.ClientID == "" {
		return fiber.NewError(fiber.StatusNotImplemented, "Google OAuth not configured")
	}

	state := "triply-state" // In production, use random state and store in session
	authURL := h.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	return c.Redirect(authURL, fiber.StatusTemporaryRedirect)
}

// GoogleCallback handles GET /auth/google/callback
func (h *AuthHandler) GoogleCallback(c *fiber.Ctx) error {
	if h.oauthConfig == nil || h.oauthConfig.ClientID == "" {
		return fiber.NewError(fiber.StatusNotImplemented, "Google OAuth not configured")
	}

	code := c.Query("code")
	if code == "" {
		return fiber.NewError(fiber.StatusBadRequest, "missing code")
	}

	// Exchange code for token
	token, err := h.oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "oauth exchange failed")
	}

	// Get user info from Google
	client := h.oauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "failed to fetch user info")
	}
	defer resp.Body.Close()

	var googleUser service.GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to decode user info")
	}

	// Get or create user
	user, err := h.authService.GetOrCreateUserFromGoogle(c.Context(), &googleUser)
	if err != nil {
		return err
	}

	// Generate JWT token
	jwtToken, err := h.authService.GenerateToken(user, h.jwtSecret)
	if err != nil {
		return err
	}

	// Set cookie with JWT token
	c.Cookie(&fiber.Cookie{
		Name:     "triply_token",
		Value:    jwtToken,
		HTTPOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: "Lax",
		Path:     "/",
		MaxAge:   60 * 60 * 24 * 7, // 7 days
	})

	// Also set user ID cookie for backward compatibility
	c.Cookie(&fiber.Cookie{
		Name:     "triply_user",
		Value:    user.ID,
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
		Path:     "/",
		MaxAge:   60 * 60 * 24 * 7,
	})

	// Redirect to frontend
	return c.Redirect(h.frontendOrigin)
}

// GetMe handles GET /auth/me
func (h *AuthHandler) GetMe(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return utils.NewUnauthorizedError()
	}

	user, err := h.authService.GetUserByID(c.Context(), userID)
	if err != nil {
		return utils.NewUnauthorizedError()
	}

	return c.JSON(user)
}

// Logout handles POST /auth/logout
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	// Clear cookies
	c.Cookie(&fiber.Cookie{
		Name:     "triply_token",
		Value:    "",
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
		Path:     "/",
		MaxAge:   -1,
	})

	c.Cookie(&fiber.Cookie{
		Name:     "triply_user",
		Value:    "",
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
		Path:     "/",
		MaxAge:   -1,
	})

	return c.SendStatus(fiber.StatusNoContent)
}

// DevLogin handles POST /auth/dev-login (for development only)
func (h *AuthHandler) DevLogin(c *fiber.Ctx) error {
	// Set demo user cookie for development
	c.Cookie(&fiber.Cookie{
		Name:     "triply_user",
		Value:    "user-sarah",
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
		Path:     "/",
		MaxAge:   60 * 60 * 24 * 30,
	})

	return c.JSON(fiber.Map{"ok": true})
}

// UpdateProfile handles PUT /api/user/profile
func (h *AuthHandler) UpdateProfile(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return utils.NewUnauthorizedError()
	}

	var req struct {
		DisplayName string `json:"displayName"`
	}
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	// Validate display name
	if req.DisplayName == "" {
		return utils.NewValidationError("display name is required")
	}
	if len(req.DisplayName) > 50 {
		return utils.NewValidationError("display name must be 50 characters or less")
	}

	user, err := h.authService.UpdateDisplayName(c.Context(), userID, req.DisplayName)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"user": user})
}

// MigrateShadowTrips handles POST /auth/migrate-shadow-trips
// Migrates trips from shadow user to authenticated user
func (h *AuthHandler) MigrateShadowTrips(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return utils.NewUnauthorizedError()
	}

	var req struct {
		ShadowUserID string `json:"shadowUserId"`
	}
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if req.ShadowUserID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "shadowUserId is required")
	}

	// Migrate shadow trips to authenticated user
	if err := h.tripService.MigrateShadowTrips(c.Context(), req.ShadowUserID, userID); err != nil {
		return err
	}

	return c.JSON(fiber.Map{"success": true, "message": "Shadow trips migrated successfully"})
}
