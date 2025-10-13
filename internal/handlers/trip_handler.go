package handlers

import (
	"time"
	"triply-server/internal/dto"
	"triply-server/internal/middleware"
	"triply-server/internal/models"
	"triply-server/internal/service"
	"triply-server/internal/utils"

	"github.com/gofiber/fiber/v2"
)

// TripHandler handles trip-related HTTP requests
type TripHandler struct {
	tripService service.TripService
}

// NewTripHandler creates a new trip handler instance
func NewTripHandler(tripService service.TripService) *TripHandler {
	return &TripHandler{tripService: tripService}
}

// ListTrips handles GET /api/users/:userId/trips
func (h *TripHandler) ListTrips(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	shadowUserID := middleware.GetShadowUserID(c)

	// Allow either authenticated user or shadow user
	if userID == "" && shadowUserID == "" {
		return utils.NewUnauthorizedError()
	}

	var trips []models.Trip
	var err error

	if userID != "" {
		trips, err = h.tripService.ListTrips(c.Context(), userID)
	} else {
		trips, err = h.tripService.GetShadowUserTrips(c.Context(), shadowUserID)
	}

	if err != nil {
		return err
	}

	return c.JSON(dto.TripListResponse{Trips: trips})
}

// CreateTrip handles POST /api/users/:userId/trips
func (h *TripHandler) CreateTrip(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	shadowUserID := middleware.GetShadowUserID(c)

	// Allow either authenticated user or shadow user
	if userID == "" && shadowUserID == "" {
		return utils.NewUnauthorizedError()
	}

	var req dto.CreateTripRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	var created *models.Trip
	var err error

	// Set the user ID on the trip
	if userID != "" {
		req.Trip.UserID = userID
		created, err = h.tripService.CreateTrip(c.Context(), &req.Trip)
	} else {
		created, err = h.tripService.CreateShadowTrip(c.Context(), &req.Trip, shadowUserID)
	}

	if err != nil {
		return err
	}

	return c.JSON(dto.TripDetailResponse{Trip: *created})
}

// UpdateTrip handles PUT /api/users/:userId/trips/:tripId
func (h *TripHandler) UpdateTrip(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	shadowUserID := middleware.GetShadowUserID(c)

	// Allow either authenticated user or shadow user
	if userID == "" && shadowUserID == "" {
		return utils.NewUnauthorizedError()
	}

	tripID := c.Params("tripId")
	if tripID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "tripId is required")
	}

	var req dto.UpdateTripRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	// Ensure IDs match
	if req.Trip.ID == "" {
		req.Trip.ID = tripID
	}
	req.Trip.UpdatedAt = time.Now()

	var updated *models.Trip
	var err error

	// Set the user ID on the trip
	if userID != "" {
		req.Trip.UserID = userID
		updated, err = h.tripService.UpdateTrip(c.Context(), &req.Trip)
	} else {
		updated, err = h.tripService.UpdateShadowTrip(c.Context(), &req.Trip, shadowUserID)
	}

	if err != nil {
		return err
	}

	return c.JSON(dto.TripDetailResponse{Trip: *updated})
}

// DeleteTrip handles DELETE /api/users/:userId/trips/:tripId
func (h *TripHandler) DeleteTrip(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return utils.NewUnauthorizedError()
	}

	tripID := c.Params("tripId")
	if tripID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "tripId is required")
	}

	if err := h.tripService.DeleteTrip(c.Context(), userID, tripID); err != nil {
		return err
	}

	return c.JSON(dto.DeleteResponse{Success: true})
}

// CloneTrip handles POST /api/trips/clone/:tripId
func (h *TripHandler) CloneTrip(c *fiber.Ctx) error {
	// Must be authenticated (shadow users cannot clone trips)
	userID := middleware.GetUserID(c)
	if userID == "" {
		return utils.NewUnauthorizedError()
	}

	tripID := c.Params("tripId")
	if tripID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "tripId is required")
	}

	// Parse request body
	var req struct {
		TripName string `json:"tripName" validate:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if req.TripName == "" {
		return fiber.NewError(fiber.StatusBadRequest, "tripName is required")
	}

	// Clone the trip
	clonedTrip, err := h.tripService.ClonePublicTrip(c.Context(), tripID, userID, req.TripName)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(clonedTrip)
}
