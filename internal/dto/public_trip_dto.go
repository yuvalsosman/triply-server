package dto

import "triply-server/internal/models"

// PublicTripSummary represents a summarized public trip for list views
type PublicTripSummary struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Slug          string   `json:"slug"`
	HeroImageURL  string   `json:"heroImageUrl"`
	Summary       *string  `json:"summary,omitempty"`
	OriginCities  []string `json:"originCities"`
	DurationDays  int      `json:"durationDays"`
	StartMonth    int      `json:"startMonth"`
	Seasons       []string `json:"seasons"`
	BudgetLevel   string   `json:"budgetLevel"`
	Pace          string   `json:"pace"`
	Tags          []string `json:"tags"`
	TravelerTypes []string `json:"travelerTypes"`
	UpdatedAt     string   `json:"updatedAt"`
	Likes         int      `json:"likes"`
}

// Author information for public trips
type Author struct {
	Name      string  `json:"name"`
	AvatarURL *string `json:"avatarUrl,omitempty"`
	City      *string `json:"city,omitempty"`
}

// EstimatedCost for public trips
type EstimatedCost struct {
	Amount   int    `json:"amount"`
	Currency string `json:"currency"`
}

// Metadata for public trips
type Metadata struct {
	CreatedAt   string `json:"createdAt"`
	PublishedAt string `json:"publishedAt"`
	ViewCount   int    `json:"viewCount"`
	Likes       int    `json:"likes"`
}

// PublicTripDetail represents full details of a public trip
type PublicTripDetail struct {
	PublicTripSummary
	Highlights    []string         `json:"highlights,omitempty"`
	Itinerary     []models.DayPlan `json:"itinerary"`
	Author        Author           `json:"author"`
	EstimatedCost *EstimatedCost   `json:"estimatedCost,omitempty"`
	Metadata      Metadata         `json:"metadata"`
}

// ListPublicTripsRequest represents filters for listing public trips
type ListPublicTripsRequest struct {
	Query         *string  `json:"query"`
	Cities        []string `json:"cities"`
	MinDays       *int     `json:"minDays"`
	MaxDays       *int     `json:"maxDays"`
	Months        []int    `json:"months"`
	Seasons       []string `json:"seasons"`
	BudgetLevels  []string `json:"budgetLevels"`
	Paces         []string `json:"paces"`
	Tags          []string `json:"tags"`
	TravelerTypes []string `json:"travelerTypes"`
	Sort          string   `json:"sort"`
	Page          int      `json:"page"`
	PageSize      int      `json:"pageSize"`
}

// ListPublicTripsResponse represents the response for listing public trips
type ListPublicTripsResponse struct {
	Trips    []PublicTripSummary `json:"trips"`
	Total    int                 `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"pageSize"`
}

// PublicTripDetailResponse represents the response for getting a public trip
type PublicTripDetailResponse struct {
	Trip PublicTripDetail `json:"trip"`
}

// ToggleVisibilityRequest represents a request to change trip visibility
type ToggleVisibilityRequest struct {
	TripID     string `json:"tripId"`
	Visibility string `json:"visibility"`
}

// ToggleVisibilityResponse represents the response for toggling visibility
type ToggleVisibilityResponse struct {
	Trip PublicTripDetail `json:"trip"`
}
