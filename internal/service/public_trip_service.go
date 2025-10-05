package service

import (
	"context"
	"fmt"
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
	PublishTrip(ctx context.Context, userID, tripID string) (*models.PublicTrip, error)
	UnpublishTrip(ctx context.Context, userID, tripID string) error
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
		Seasons:       req.Seasons,
		BudgetLevels:  req.BudgetLevels,
		Paces:         req.Paces,
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

func (s *publicTripService) PublishTrip(ctx context.Context, userID, tripID string) (*models.PublicTrip, error) {
	// Get trip
	trip, err := s.tripRepo.FindByID(ctx, tripID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, utils.NewNotFoundError("Trip")
		}
		return nil, err
	}

	// Update trip visibility
	trip.Visibility = "public"
	if err := s.tripRepo.Update(ctx, trip); err != nil {
		return nil, err
	}

	// Check if public trip record already exists
	existingPT, err := s.publicTripRepo.FindByTripID(ctx, tripID)
	if err == nil {
		// Update existing
		existingPT.UpdatedAt = time.Now()
		if err := s.publicTripRepo.Update(ctx, existingPT); err != nil {
			return nil, err
		}
		return existingPT, nil
	}

	// Create new public trip record
	publicTrip := s.generatePublicTripFromTrip(trip)
	if err := s.publicTripRepo.Create(ctx, publicTrip); err != nil {
		return nil, err
	}

	return publicTrip, nil
}

func (s *publicTripService) UnpublishTrip(ctx context.Context, userID, tripID string) error {
	// Get trip
	trip, err := s.tripRepo.FindByID(ctx, tripID, userID)
	if err != nil {
		return err
	}

	// Update trip visibility
	trip.Visibility = "private"
	return s.tripRepo.Update(ctx, trip)
}

func (s *publicTripService) ToggleVisibility(ctx context.Context, userID, tripID, visibility string) (*dto.PublicTripDetail, error) {
	// Get trip
	trip, err := s.tripRepo.FindByID(ctx, tripID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, utils.NewNotFoundError("Trip")
		}
		return nil, err
	}

	// Update visibility
	trip.Visibility = visibility
	if err := s.tripRepo.Update(ctx, trip); err != nil {
		return nil, err
	}

	// If making public, ensure public trip record exists
	if visibility == "public" {
		_, err := s.PublishTrip(ctx, userID, tripID)
		if err != nil {
			return nil, err
		}
	}

	// Get public trip
	publicTrip, err := s.publicTripRepo.FindByTripID(ctx, tripID)
	if err != nil {
		// If not found, create a minimal response
		return &dto.PublicTripDetail{
			PublicTripSummary: dto.PublicTripSummary{
				ID:    tripID,
				Title: trip.Name,
			},
			Metadata: models.PublicTripMetadata{
				Visibility: visibility,
				UpdatedAt:  time.Now(),
			},
		}, nil
	}

	return s.toPublicTripDetail(publicTrip), nil
}

func (s *publicTripService) generatePublicTripFromTrip(trip *models.Trip) *models.PublicTrip {
	// Extract origin cities from destinations
	cities := make([]string, 0, len(trip.Destinations))
	for _, dest := range trip.Destinations {
		cities = append(cities, dest.City)
	}

	// Calculate duration
	// This is simplified - in reality you'd parse the dates
	duration := 7 // Default

	return &models.PublicTrip{
		ID:     utils.GenerateID("pt"),
		TripID: trip.ID,
		Slug:   fmt.Sprintf("%s-%s", trip.Name, trip.ID),
		HeroImageURL: func() string {
			if trip.CoverImage != nil {
				return *trip.CoverImage
			}
			return ""
		}(),
		OriginCities:  cities,
		DurationDays:  duration,
		StartMonth:    1, // Parse from trip.StartDate
		Seasons:       []string{"spring"},
		BudgetLevel:   "moderate",
		Pace:          "balanced",
		Tags:          []string{},
		TravelerTypes: []string{},
		Likes:         0,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

func (s *publicTripService) toPublicTripSummary(pt *models.PublicTrip) dto.PublicTripSummary {
	title := pt.Slug
	if pt.Trip != nil {
		title = pt.Trip.Name
	}

	return dto.PublicTripSummary{
		ID:            pt.ID,
		Title:         title,
		Slug:          pt.Slug,
		HeroImageURL:  pt.HeroImageURL,
		Summary:       pt.Summary,
		OriginCities:  pt.OriginCities,
		DurationDays:  pt.DurationDays,
		StartMonth:    pt.StartMonth,
		Seasons:       pt.Seasons,
		BudgetLevel:   pt.BudgetLevel,
		Pace:          pt.Pace,
		Tags:          pt.Tags,
		TravelerTypes: pt.TravelerTypes,
		UpdatedAt:     pt.UpdatedAt.Format(time.RFC3339),
		Likes:         pt.Likes,
	}
}

func (s *publicTripService) toPublicTripDetail(pt *models.PublicTrip) *dto.PublicTripDetail {
	summary := s.toPublicTripSummary(pt)

	itinerary := []models.DayPlan{}
	if pt.Trip != nil {
		for _, dest := range pt.Trip.Destinations {
			itinerary = append(itinerary, dest.DailyPlans...)
		}
	}

	detail := &dto.PublicTripDetail{
		PublicTripSummary: summary,
		Highlights:        pt.Highlights,
		Itinerary:         itinerary,
		Author: models.PublicTripAuthor{
			Name:      pt.AuthorName,
			AvatarURL: pt.AuthorAvatarURL,
			HomeCity:  pt.AuthorHomeCity,
		},
		Metadata: models.PublicTripMetadata{
			Visibility: func() string {
				if pt.Trip != nil {
					return pt.Trip.Visibility
				}
				return "public"
			}(),
			CreatedAt: pt.CreatedAt,
			UpdatedAt: pt.UpdatedAt,
		},
	}

	if pt.EstimatedCostCurrency != nil && pt.EstimatedCostAmount != nil {
		detail.EstimatedCost = &models.PublicTripCost{
			Currency: *pt.EstimatedCostCurrency,
			Amount:   *pt.EstimatedCostAmount,
		}
	}

	return detail
}
