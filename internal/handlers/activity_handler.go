package handlers

import (
	"triply-server/internal/dto"
	"triply-server/internal/service"

	"github.com/gofiber/fiber/v2"
)

// ActivityHandler handles activity-related HTTP requests
type ActivityHandler struct {
	activityService service.ActivityService
}

// NewActivityHandler creates a new activity handler instance
func NewActivityHandler(activityService service.ActivityService) *ActivityHandler {
	return &ActivityHandler{activityService: activityService}
}

// UpdateActivityOrder handles POST /api/activities/order
func (h *ActivityHandler) UpdateActivityOrder(c *fiber.Ctx) error {
	var req dto.ActivityOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if req.DayID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "dayId is required")
	}

	// Update activity orders in database
	if err := h.activityService.UpdateActivityOrders(c.Context(), req.DayID, req.Activities); err != nil {
		return err
	}

	// Return normalized orders
	response := make([]dto.ActivityOrderPayload, 0, len(req.Activities))
	for _, activity := range req.Activities {
		response = append(response, dto.ActivityOrderPayload{
			ID:        activity.ID,
			Order:     activity.Order,
			TimeOfDay: activity.TimeOfDay,
		})
	}

	return c.JSON(response)
}
