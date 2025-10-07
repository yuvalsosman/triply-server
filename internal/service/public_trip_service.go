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
	ListPublicTrips(ctx context.Context, req *dto.ListPublicTripsRequest) (*dto.ListPublicTripsResponse, error)
	GetPublicTrip(ctx context.Context, tripID string) (*dto.PublicTripDetail, error)
	ToggleVisibility(ctx context.Context, userID, tripID, visibility string) (*dto.PublicTripDetail, error)
}

type publicTripService struct {
	publicTripRepo repository.PublicTripRepository
	tripRepo       repository.TripRepository
}

// NewPublicTripService creates a new public trip service instance
func NewPublicTripService(publicTripRepo repository.PublicTripRepository, tripRepo repository.TripRepository) PublicTripService {
	return &publicTripService{
		publicTripRepo: publicTripRepo,
		tripRepo:       tripRepo,
	}
}

func (s *publicTripService) ListPublicTrips(ctx context.Context, req *dto.ListPublicTripsRequest) (*dto.ListPublicTripsResponse, error) {
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

	// Build filters
	filters := &repository.PublicTripFilters{
		Query:         req.Query,
		Cities:        req.Cities,
		MinDays:       req.MinDays,
		MaxDays:       req.MaxDays,
		Months:        req.Months,
		Tags:          req.Tags,
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
	for i, pt := range publicTrips {
		summaries[i] = s.toPublicTripSummary(&pt)
	}

	return &dto.ListPublicTripsResponse{
		Trips:    summaries,
		Total:    int(total),
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (s *publicTripService) GetPublicTrip(ctx context.Context, tripID string) (*dto.PublicTripDetail, error) {
	publicTrip, err := s.publicTripRepo.FindByID(ctx, tripID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, utils.NewNotFoundError("Public trip")
		}
		return nil, err
	}

	return s.toPublicTripDetail(publicTrip), nil
}

func (s *publicTripService) ToggleVisibility(ctx context.Context, userID, tripID, visibility string) (*dto.PublicTripDetail, error) {
	// Validate visibility
	if visibility != "public" && visibility != "private" && visibility != "unlisted" {
		return nil, utils.NewValidationError("visibility must be 'public', 'private', or 'unlisted'")
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

	// Calculate duration from day plans count (more reliable than parsing dates)
	durationDays := len(trip.DayPlans)
	if durationDays == 0 {
		// Fallback to date calculation if no day plans
		start, _ := time.Parse("2006-01-02", trip.StartDate)
		end, _ := time.Parse("2006-01-02", trip.EndDate)
		if !start.IsZero() && !end.IsZero() {
			durationDays = int(end.Sub(start).Hours()/24) + 1
		}
	}

	slug := ""
	if trip.Slug != nil {
		slug = *trip.Slug
	}

	heroImage := ""
	if trip.HeroImage != nil {
		heroImage = *trip.HeroImage
	}

	// Get start month from first day plan if available
	startMonth := 1
	if len(trip.DayPlans) > 0 {
		firstDate, err := time.Parse(time.RFC3339, trip.DayPlans[0].Date)
		if err == nil {
			startMonth = int(firstDate.Month())
		}
	}

	return dto.PublicTripSummary{
		ID:           trip.ID,
		Title:        trip.Name,
		Slug:         slug,
		HeroImageURL: heroImage,
		Summary:      trip.Summary,
		OriginCities: originCities,
		DurationDays: durationDays,
		StartMonth:   startMonth,
		Tags:         trip.Tags,
		TravelerType: trip.TravelerType,
		UpdatedAt:    trip.UpdatedAt.Format(time.RFC3339),
		Likes:        trip.Likes,
	}
}

func (s *publicTripService) toPublicTripDetail(trip *models.Trip) *dto.PublicTripDetail {
	summary := s.toPublicTripSummary(trip)

	// Build author info
	author := dto.Author{
		Name: "Anonymous",
	}
	if trip.User != nil {
		author.Name = trip.User.Name
		author.AvatarURL = trip.User.AvatarURL
	}

	// Build metadata
	publishedAt := ""
	if trip.PublishedAt != nil {
		publishedAt = trip.PublishedAt.Format(time.RFC3339)
	}

	metadata := dto.Metadata{
		CreatedAt:   trip.CreatedAt.Format(time.RFC3339),
		PublishedAt: publishedAt,
		Likes:       trip.Likes,
	}

	return &dto.PublicTripDetail{
		PublicTripSummary: summary,
		Itinerary:         trip.DayPlans,
		Author:            author,
		Metadata:          metadata,
	}
}
