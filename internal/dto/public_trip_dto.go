package dto

import "triply-server/internal/models"

// PublicTripSummary represents a summarized public trip for list views
type PublicTripSummary struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Slug          string   `json:"slug"`
	CoverImageURL string   `json:"coverImageUrl"`
	Summary       *string  `json:"summary,omitempty"`
	OriginCities  []string `json:"originCities"`
	DurationDays  int      `json:"durationDays"`
	StartDate     string   `json:"startDate"` // YYYY-MM-DD format
	EndDate       string   `json:"endDate"`   // YYYY-MM-DD format
	StartMonth    int      `json:"startMonth"`
	EndMonth      int      `json:"endMonth"`
	TravelerType  string   `json:"travelerType"` // Single value
	UpdatedAt     string   `json:"updatedAt"`
	Likes         int      `json:"likes"`
	HasLiked      *bool    `json:"hasLiked,omitempty"` // null if not authenticated
}

// Author information for public trips
type Author struct {
	Name      string  `json:"name"`
	AvatarURL *string `json:"avatarUrl,omitempty"`
	City      *string `json:"city,omitempty"`
}

// Metadata for public trips
type Metadata struct {
	CreatedAt string `json:"createdAt"`
	Likes     int    `json:"likes"`
}

// PublicTripDetail represents full details of a public trip
type PublicTripDetail struct {
	PublicTripSummary
	Highlights []string         `json:"highlights,omitempty"`
	Itinerary  []models.DayPlan `json:"itinerary"`
	Author     Author           `json:"author"`
	Metadata   Metadata         `json:"metadata"`
}

// DurationRange represents a duration filter range
type DurationRange struct {
	MinDays *int `json:"minDays"`
	MaxDays *int `json:"maxDays"`
}

// ListPublicTripsRequest represents filters for listing public trips
type ListPublicTripsRequest struct {
	Query         *string         `json:"query"`
	Cities        []string        `json:"cities"`
	Durations     []DurationRange `json:"durations"` // Array of duration ranges for multiple selections
	Months        []int           `json:"months"`
	TravelerTypes []string        `json:"travelerTypes"` // Array for multiple selections
	Sort          string          `json:"sort"`
	Page          int             `json:"page"`
	PageSize      int             `json:"pageSize"`
}

// ListPublicTripsResponse represents the response for listing public trips
type ListPublicTripsResponse struct {
	Trips        []PublicTripSummary `json:"trips"`
	Total        int                 `json:"total"`
	Page         int                 `json:"page"`
	PageSize     int                 `json:"pageSize"`
	HasMorePages bool                `json:"hasMorePages"`
}

// PublicTripDetailResponse represents the response for getting a public trip
type PublicTripDetailResponse struct {
	Trip PublicTripDetail `json:"trip"`
}

// ToggleVisibilityRequest represents a request to change trip visibility
type ToggleVisibilityRequest struct {
	TripID     string `json:"tripId"`
	Visibility string `json:"visibility"` // Must be "public" or "private"
}

// ToggleVisibilityResponse represents the response for toggling visibility
type ToggleVisibilityResponse struct {
	Trip PublicTripDetail `json:"trip"`
}

// LikeToggleResponse represents the response after toggling a like
type LikeToggleResponse struct {
	Liked      bool `json:"liked"`
	TotalLikes int  `json:"totalLikes"`
}
