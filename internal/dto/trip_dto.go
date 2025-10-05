package dto

import "triply-server/internal/models"

// TripListResponse represents the response for listing trips
type TripListResponse struct {
	Trips []models.Trip `json:"trips"`
}

// TripDetailResponse represents the response for a single trip
type TripDetailResponse struct {
	Trip models.Trip `json:"trip"`
}

// CreateTripRequest represents the request to create a trip
type CreateTripRequest struct {
	Trip models.Trip `json:"trip"`
}

// UpdateTripRequest represents the request to update a trip
type UpdateTripRequest struct {
	Trip models.Trip `json:"trip"`
}

// DeleteResponse represents a successful deletion
type DeleteResponse struct {
	Success bool `json:"success"`
}
