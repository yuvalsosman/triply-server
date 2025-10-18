package service

import (
	"context"
	"time"
	"triply-server/internal/dto"
	"triply-server/internal/models"
	"triply-server/internal/repository"
	"triply-server/internal/utils"

	"gorm.io/gorm"
)

// PublicTripService defines the interface for public trip operations
type PublicTripService interface {
	ListPublicTrips(ctx context.Context, req *dto.ListPublicTripsRequest, userID *string) (*dto.ListPublicTripsResponse, error)
	GetPublicTrip(ctx context.Context, tripID string, userID *string) (*dto.PublicTripDetail, error)
	ToggleVisibility(ctx context.Context, userID, tripID, visibility string) (*dto.PublicTripDetail, error)
}

type publicTripService struct {
	publicTripRepo repository.PublicTripRepository
	tripRepo       repository.TripRepository
	tripLikeRepo   repository.TripLikeRepository
}

// NewPublicTripService creates a new public trip service instance
func NewPublicTripService(publicTripRepo repository.PublicTripRepository, tripRepo repository.TripRepository, tripLikeRepo repository.TripLikeRepository) PublicTripService {
	return &publicTripService{
		publicTripRepo: publicTripRepo,
		tripRepo:       tripRepo,
		tripLikeRepo:   tripLikeRepo,
	}
}

func (s *publicTripService) ListPublicTrips(ctx context.Context, req *dto.ListPublicTripsRequest, userID *string) (*dto.ListPublicTripsResponse, error) {
	// Set defaults
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 12
	}
	if req.Sort == "" {
		req.Sort = "featured"
	}

	// Convert DTO durations to repository durations
	var durations []repository.DurationRange
	for _, d := range req.Durations {
		durations = append(durations, repository.DurationRange{
			MinDays: d.MinDays,
			MaxDays: d.MaxDays,
		})
	}

	// Build filters
	filters := &repository.PublicTripFilters{
		Query:         req.Query,
		Cities:        req.Cities,
		Durations:     durations,
		Months:        req.Months,
		TravelerTypes: req.TravelerTypes,
		Sort:          req.Sort,
		Page:          req.Page,
		PageSize:      req.PageSize,
	}

	publicTrips, total, err := s.publicTripRepo.FindAll(ctx, filters)
	if err != nil {
		return nil, err
	}

	// Convert to DTOs
	summaries := make([]dto.PublicTripSummary, len(publicTrips))
	tripIDs := make([]string, len(publicTrips))
	for i, pt := range publicTrips {
		summaries[i] = s.toPublicTripSummary(&pt)
		tripIDs[i] = pt.ID
	}

	// If user is authenticated, batch check which trips they liked
	if userID != nil && len(tripIDs) > 0 {
		likedTripIDs, err := s.tripLikeRepo.GetUserLikedTripIDs(ctx, *userID, tripIDs)
		if err == nil {
			// Create a map for O(1) lookup
			likedMap := make(map[string]bool)
			for _, id := range likedTripIDs {
				likedMap[id] = true
			}

			// Update summaries with hasLiked
			for i := range summaries {
				hasLiked := likedMap[summaries[i].ID]
				summaries[i].HasLiked = &hasLiked
			}
		}
	}

	// Calculate if there are more pages
	hasMorePages := (req.Page * req.PageSize) < int(total)

	return &dto.ListPublicTripsResponse{
		Trips:        summaries,
		Total:        int(total),
		Page:         req.Page,
		PageSize:     req.PageSize,
		HasMorePages: hasMorePages,
	}, nil
}

func (s *publicTripService) GetPublicTrip(ctx context.Context, tripID string, userID *string) (*dto.PublicTripDetail, error) {
	publicTrip, err := s.publicTripRepo.FindByID(ctx, tripID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, utils.NewNotFoundError("Public trip")
		}
		return nil, err
	}

	detail := s.toPublicTripDetail(publicTrip)

	// If user is authenticated, check if they liked this trip
	if userID != nil {
		hasLiked, err := s.tripLikeRepo.HasLiked(ctx, *userID, tripID)
		if err == nil {
			detail.HasLiked = &hasLiked
		}
	}

	return detail, nil
}

func (s *publicTripService) ToggleVisibility(ctx context.Context, userID, tripID, visibility string) (*dto.PublicTripDetail, error) {
	// Validate visibility
	if visibility != "public" && visibility != "private" {
		return nil, utils.NewValidationError("visibility must be 'public' or 'private'")
	}

	// Toggle visibility
	if err := s.publicTripRepo.ToggleVisibility(ctx, tripID, userID, visibility); err != nil {
		return nil, err
	}

	// Get updated trip
	trip, err := s.tripRepo.FindByID(ctx, tripID, userID)
	if err != nil {
		return nil, err
	}

	return s.toPublicTripDetail(trip), nil
}

// Helper methods to convert models to DTOs
func (s *publicTripService) toPublicTripSummary(trip *models.Trip) dto.PublicTripSummary {
	// Extract origin cities from destinations
	originCities := make([]string, 0)
	for _, td := range trip.TripDestinations {
		if td.Destination != nil {
			originCities = append(originCities, td.Destination.City)
		}
	}

	// Parse trip dates (try RFC3339 first, then YYYY-MM-DD format)
	start, errStart := time.Parse(time.RFC3339, trip.StartDate)
	if errStart != nil {
		start, errStart = time.Parse("2006-01-02", trip.StartDate)
	}

	end, errEnd := time.Parse(time.RFC3339, trip.EndDate)
	if errEnd != nil {
		end, errEnd = time.Parse("2006-01-02", trip.EndDate)
	}

	// Calculate duration - always prefer date calculation if dates are valid
	durationDays := 0
	if errStart == nil && errEnd == nil {
		// Calculate from actual trip dates
		durationDays = int(end.Sub(start).Hours()/24) + 1
	} else if len(trip.DayPlans) > 0 {
		// Fallback to day plans count if date parsing failed
		durationDays = len(trip.DayPlans)
	}

	slug := ""
	if trip.Slug != nil {
		slug = *trip.Slug
	}

	// Get months - always prefer trip dates if available, otherwise use day plans
	startMonth := 1
	endMonth := 1

	if errStart == nil && errEnd == nil {
		// Use trip dates for months
		startMonth = int(start.Month())
		endMonth = int(end.Month())
	} else if len(trip.DayPlans) > 0 {
		// Fallback to day plans if date parsing failed
		firstDate, err := time.Parse(time.RFC3339, trip.DayPlans[0].Date)
		if err == nil {
			startMonth = int(firstDate.Month())
		}

		lastDate, err := time.Parse(time.RFC3339, trip.DayPlans[len(trip.DayPlans)-1].Date)
		if err == nil {
			endMonth = int(lastDate.Month())
		} else {
			endMonth = startMonth
		}
	}

	return dto.PublicTripSummary{
		ID:            trip.ID,
		Title:         trip.Name,
		Slug:          slug,
		CoverImageURL: trip.CoverImage,
		Summary:       trip.Summary,
		OriginCities:  originCities,
		DurationDays:  durationDays,
		StartMonth:    startMonth,
		EndMonth:      endMonth,
		TravelerType:  trip.TravelerType,
		UpdatedAt:     trip.UpdatedAt.Format(time.RFC3339),
		Likes:         trip.Likes,
	}
}

func (s *publicTripService) toPublicTripDetail(trip *models.Trip) *dto.PublicTripDetail {
	summary := s.toPublicTripSummary(trip)

	// Build author info
	author := dto.Author{
		Name: "Anonymous",
	}
	if trip.User != nil {
		// Prefer DisplayName over Name for public display
		if trip.User.DisplayName != nil && *trip.User.DisplayName != "" {
			author.Name = *trip.User.DisplayName
		} else {
			author.Name = trip.User.Name
		}
		author.AvatarURL = trip.User.AvatarURL
	}

	// Build metadata
	metadata := dto.Metadata{
		CreatedAt: trip.CreatedAt.Format(time.RFC3339),
		Likes:     trip.Likes,
	}

	return &dto.PublicTripDetail{
		PublicTripSummary: summary,
		Itinerary:         trip.DayPlans,
		Author:            author,
		Metadata:          metadata,
	}
}
