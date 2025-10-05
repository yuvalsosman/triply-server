package handlers

import (
	"triply-server/internal/dto"
	"triply-server/internal/middleware"
	"triply-server/internal/service"

	"github.com/gofiber/fiber/v2"
)

// ImportHandler handles import-related HTTP requests
type ImportHandler struct {
	importService service.ImportService
}

// NewImportHandler creates a new import handler instance
func NewImportHandler(importService service.ImportService) *ImportHandler {
	return &ImportHandler{importService: importService}
}

// ImportTripParts handles POST /api/import-trip
func (h *ImportHandler) ImportTripParts(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "authentication required")
	}

	var req dto.ImportTripRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	result, err := h.importService.ImportTripParts(c.Context(), userID, &req)
	if err != nil {
		return err
	}

	return c.JSON(result)
}
