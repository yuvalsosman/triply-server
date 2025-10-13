package handlers

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/gofiber/fiber/v2"
)

type MapsHandler struct {
	apiKey string
}

func NewMapsHandler(apiKey string) *MapsHandler {
	return &MapsHandler{apiKey: apiKey}
}

// GetMapConfig returns the Google Maps API key to authenticated users
// This endpoint should have rate limiting applied
func (h *MapsHandler) GetMapConfig(c *fiber.Ctx) error {
	// Return the API key - it will be restricted by HTTP referrer in Google Cloud Console
	return c.JSON(fiber.Map{
		"apiKey": h.apiKey,
	})
}

// ProxyPhoto proxies requests to Google Maps/Places APIs with the server's API key
// This prevents exposing the API key in client-side requests
func (h *MapsHandler) ProxyPhoto(c *fiber.Ctx) error {
	// Get the target URL from query parameter
	targetURL := c.Query("url")
	if targetURL == "" {
		return fiber.NewError(fiber.StatusBadRequest, "url parameter is required")
	}

	// Decode the URL (it might be double-encoded)
	decodedURL, err := url.QueryUnescape(targetURL)
	if err != nil {
		// If it fails, use the original
		decodedURL = targetURL
	}

	// Parse and validate the URL
	parsedURL, err := url.Parse(decodedURL)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid url parameter")
	}

	// Only allow requests to Places API (no Street View to avoid charges)
	allowedHosts := []string{
		"places.googleapis.com",
		"lh3.googleusercontent.com",
	}

	allowed := false
	for _, host := range allowedHosts {
		if parsedURL.Host == host {
			allowed = true
			break
		}
	}

	if !allowed {
		return fiber.NewError(fiber.StatusForbidden, "Only Places API photo requests are allowed")
	}

	// Add or replace the API key in the query parameters
	query := parsedURL.Query()
	query.Set("key", h.apiKey)
	parsedURL.RawQuery = query.Encode()

	// Make the request to Google API
	resp, err := http.Get(parsedURL.String())
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, fmt.Sprintf("failed to fetch from Google API: %v", err))
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, "failed to read response from Google API")
	}

	// Forward the status code and content type
	c.Status(resp.StatusCode)
	if contentType := resp.Header.Get("Content-Type"); contentType != "" {
		c.Set("Content-Type", contentType)
	}

	// Set cache headers for performance
	c.Set("Cache-Control", "public, max-age=86400") // Cache for 24 hours

	return c.Send(body)
}
