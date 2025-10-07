package handlers

import (
	"triply-server/internal/dto"
	"triply-server/internal/middleware"
	"triply-server/internal/service"

	"github.com/gofiber/fiber/v2"
)

// PublicTripHandler handles public trip-related HTTP requests
type PublicTripHandler struct {
	publicTripService service.PublicTripService
}

// NewPublicTripHandler creates a new public trip handler instance
func NewPublicTripHandler(publicTripService service.PublicTripService) *PublicTripHandler {
	return &PublicTripHandler{publicTripService: publicTripService}
}

// ListPublicTrips handles GET /api/public-trips
func (h *PublicTripHandler) ListPublicTrips(c *fiber.Ctx) error {
	// Parse query parameters
	req := &dto.ListPublicTripsRequest{
		Page:     c.QueryInt("page", 1),
		PageSize: c.QueryInt("pageSize", 12),
		Sort:     c.Query("sort", "featured"),
	}

	// Parse filters from query string if provided
	// For simplicity, we'll handle basic filters
	if query := c.Query("query"); query != "" {
		req.Query = &query
	}

	// Get user ID if authenticated (optional)
	var userID *string
	if uid := middleware.GetUserID(c); uid != "" {
		userID = &uid
	}

	resp, err := h.publicTripService.ListPublicTrips(c.Context(), req, userID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// GetPublicTripDetail handles GET /api/public-trips/:tripId
func (h *PublicTripHandler) GetPublicTripDetail(c *fiber.Ctx) error {
	tripID := c.Params("tripId")
	if tripID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "tripId is required")
	}

	// Get user ID if authenticated (optional)
	var userID *string
	if uid := middleware.GetUserID(c); uid != "" {
		userID = &uid
	}

	trip, err := h.publicTripService.GetPublicTrip(c.Context(), tripID, userID)
	if err != nil {
		return err
	}

	return c.JSON(dto.PublicTripDetailResponse{Trip: *trip})
}

// ToggleVisibility handles POST /api/public-trips/:tripId/visibility
func (h *PublicTripHandler) ToggleVisibility(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "authentication required")
	}

	tripID := c.Params("tripId")
	if tripID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "tripId is required")
	}

	var req dto.ToggleVisibilityRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	// Use tripId from URL if not in body
	if req.TripID == "" {
		req.TripID = tripID
	}

	result, err := h.publicTripService.ToggleVisibility(c.Context(), userID, req.TripID, req.Visibility)
	if err != nil {
		return err
	}

	return c.JSON(dto.ToggleVisibilityResponse{Trip: *result})
}
