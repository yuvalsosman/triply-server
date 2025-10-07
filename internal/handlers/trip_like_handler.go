package handlers

import (
	"triply-server/internal/service"

	"github.com/gofiber/fiber/v2"
)

type TripLikeHandler struct {
	service service.TripLikeService
}

func NewTripLikeHandler(service service.TripLikeService) *TripLikeHandler {
	return &TripLikeHandler{
		service: service,
	}
}

// ToggleLike handles POST /api/public-trips/:tripId/like
func (h *TripLikeHandler) ToggleLike(c *fiber.Ctx) error {
	// Get authenticated user ID from context (set by auth middleware)
	userID, ok := c.Locals("userId").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	tripID := c.Params("tripId")
	if tripID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Trip ID is required",
		})
	}

	response, err := h.service.ToggleLike(c.Context(), userID, tripID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to toggle like",
		})
	}

	return c.JSON(response)
}
