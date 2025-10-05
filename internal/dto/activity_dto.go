package dto

import "triply-server/internal/models"

// ActivityOrderRequest represents a request to reorder activities
type ActivityOrderRequest struct {
	TripID     string            `json:"tripId"`
	DayID      string            `json:"dayId"`
	Activities []models.Activity `json:"activities"`
}

// ActivityOrderPayload represents the response for activity ordering
type ActivityOrderPayload struct {
	ID        string `json:"id"`
	Order     int    `json:"order"`
	TimeOfDay string `json:"timeOfDay"`
}
