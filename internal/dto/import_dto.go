package dto

import "triply-server/internal/models"

// ImportSelection represents the parts of a trip to import
type ImportSelection struct {
	DayIDs      []string `json:"dayIds,omitempty"`
	LegIDs      []string `json:"legIds,omitempty"`
	ActivityIDs []string `json:"activityIds,omitempty"`
}

// ImportTarget represents where to insert imported content
type ImportTarget struct {
	DestinationID string  `json:"destinationId"`
	DayID         *string `json:"dayId,omitempty"`
	LegID         *string `json:"legId,omitempty"`
	Mode          string  `json:"mode"` // new-day|append-day|append-leg
}

// ImportTripRequest represents a request to import trip parts
type ImportTripRequest struct {
	SourceTripID string          `json:"sourceTripId"`
	Selection    ImportSelection `json:"selection"`
	Target       ImportTarget    `json:"target"`
}

// ImportTripResponse represents the response after importing
type ImportTripResponse struct {
	ImportID    string      `json:"importId"`
	UpdatedTrip models.Trip `json:"updatedTrip"`
}
