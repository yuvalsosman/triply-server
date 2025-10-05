package service

import (
	"context"
	"time"
	"triply-server/internal/models"
	"triply-server/internal/repository"
	"triply-server/internal/utils"

	"gorm.io/gorm"
)

// TripService defines the interface for trip business logic
type TripService interface {
	GetUserTrips(ctx context.Context, userID string) ([]models.Trip, error)
	GetShadowUserTrips(ctx context.Context, shadowUserID string) ([]models.Trip, error)
	GetTrip(ctx context.Context, tripID, userID string) (*models.Trip, error)
	GetShadowUserTrip(ctx context.Context, tripID, shadowUserID string) (*models.Trip, error)
	CreateTrip(ctx context.Context, userID string, trip *models.Trip) (*models.Trip, error)
	CreateShadowTrip(ctx context.Context, shadowUserID string, trip *models.Trip) (*models.Trip, error)
	UpdateTrip(ctx context.Context, userID string, trip *models.Trip) (*models.Trip, error)
	UpdateShadowTrip(ctx context.Context, shadowUserID string, trip *models.Trip) (*models.Trip, error)
	DeleteTrip(ctx context.Context, userID string, tripID string) error
	MigrateShadowTrips(ctx context.Context, shadowUserID, userID string) error
}

type tripService struct {
	tripRepo repository.TripRepository
}

// NewTripService creates a new trip service instance
func NewTripService(tripRepo repository.TripRepository) TripService {
	return &tripService{tripRepo: tripRepo}
}

func (s *tripService) GetUserTrips(ctx context.Context, userID string) ([]models.Trip, error) {
	return s.tripRepo.FindByUserID(ctx, userID)
}

func (s *tripService) GetTrip(ctx context.Context, tripID, userID string) (*models.Trip, error) {
	trip, err := s.tripRepo.FindByID(ctx, tripID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, utils.NewNotFoundError("Trip")
		}
		return nil, err
	}
	return trip, nil
}

func (s *tripService) CreateTrip(ctx context.Context, userID string, trip *models.Trip) (*models.Trip, error) {
	// Set user ID
	trip.UserID = userID

	// Generate IDs if missing
	now := time.Now()
	if trip.ID == "" {
		trip.ID = utils.GenerateID("trip")
	}
	if trip.CreatedAt.IsZero() {
		trip.CreatedAt = now
	}
	trip.UpdatedAt = now

	// Generate IDs for nested entities
	for i := range trip.Destinations {
		dest := &trip.Destinations[i]
		if dest.ID == "" {
			dest.ID = utils.GenerateID("dest")
		}
		for j := range dest.DailyPlans {
			plan := &dest.DailyPlans[j]
			if plan.ID == "" {
				plan.ID = utils.GenerateID("day")
			}
			for k := range plan.Activities {
				act := &plan.Activities[k]
				if act.ID == "" {
					act.ID = utils.GenerateID("act")
				}
			}
		}
	}

	// Validate
	if err := s.validateTrip(trip); err != nil {
		return nil, err
	}

	// Create in repository
	if err := s.tripRepo.Create(ctx, trip); err != nil {
		return nil, err
	}

	return trip, nil
}

func (s *tripService) UpdateTrip(ctx context.Context, userID string, trip *models.Trip) (*models.Trip, error) {
	// Ensure user ID matches
	trip.UserID = userID
	trip.UpdatedAt = time.Now()

	// Validate
	if err := s.validateTrip(trip); err != nil {
		return nil, err
	}

	// Update in repository
	if err := s.tripRepo.Update(ctx, trip); err != nil {
		return nil, err
	}

	// Fetch updated trip
	return s.tripRepo.FindByID(ctx, trip.ID, userID)
}

func (s *tripService) DeleteTrip(ctx context.Context, userID string, tripID string) error {
	return s.tripRepo.Delete(ctx, tripID, userID)
}

func (s *tripService) validateTrip(trip *models.Trip) error {
	if trip.Name == "" {
		return utils.NewValidationError("trip name is required")
	}
	if trip.TravelerCount < 1 {
		return utils.NewValidationError("traveler count must be at least 1")
	}
	if trip.StartDate == "" || trip.EndDate == "" {
		return utils.NewValidationError("start and end dates are required")
	}
	if trip.StartDate > trip.EndDate {
		return utils.NewValidationError("start date must be before end date")
	}
	return nil
}

func (s *tripService) GetShadowUserTrips(ctx context.Context, shadowUserID string) ([]models.Trip, error) {
	return s.tripRepo.FindByShadowUserID(ctx, shadowUserID)
}

func (s *tripService) GetShadowUserTrip(ctx context.Context, tripID, shadowUserID string) (*models.Trip, error) {
	trip, err := s.tripRepo.FindByIDWithShadowUser(ctx, tripID, shadowUserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, utils.NewNotFoundError("Trip")
		}
		return nil, err
	}
	return trip, nil
}

func (s *tripService) CreateShadowTrip(ctx context.Context, shadowUserID string, trip *models.Trip) (*models.Trip, error) {
	// Set shadow user ID
	trip.ShadowUserID = &shadowUserID
	trip.UserID = "" // Clear user ID

	// Generate IDs if missing
	if trip.ID == "" {
		trip.ID = utils.GenerateID("trip")
	}

	for i := range trip.Destinations {
		if trip.Destinations[i].ID == "" {
			trip.Destinations[i].ID = utils.GenerateID("dest")
		}
		trip.Destinations[i].TripID = trip.ID
		for j := range trip.Destinations[i].DailyPlans {
			if trip.Destinations[i].DailyPlans[j].ID == "" {
				trip.Destinations[i].DailyPlans[j].ID = utils.GenerateID("day")
			}
			trip.Destinations[i].DailyPlans[j].DestinationID = trip.Destinations[i].ID
			for k := range trip.Destinations[i].DailyPlans[j].Activities {
				if trip.Destinations[i].DailyPlans[j].Activities[k].ID == "" {
					trip.Destinations[i].DailyPlans[j].Activities[k].ID = utils.GenerateID("act")
				}
				trip.Destinations[i].DailyPlans[j].Activities[k].DayPlanID = trip.Destinations[i].DailyPlans[j].ID
			}
		}
	}

	// Set timestamps
	now := time.Now()
	trip.CreatedAt = now
	trip.UpdatedAt = now

	// Validate
	if err := s.validateTrip(trip); err != nil {
		return nil, err
	}

	// Create in database
	if err := s.tripRepo.Create(ctx, trip); err != nil {
		return nil, err
	}

	return trip, nil
}

func (s *tripService) UpdateShadowTrip(ctx context.Context, shadowUserID string, trip *models.Trip) (*models.Trip, error) {
	// Verify ownership
	existing, err := s.tripRepo.FindByIDWithShadowUser(ctx, trip.ID, shadowUserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, utils.NewNotFoundError("Trip")
		}
		return nil, err
	}

	// Preserve shadow user ID
	trip.ShadowUserID = existing.ShadowUserID
	trip.UserID = "" // Clear user ID

	// Update timestamps
	trip.CreatedAt = existing.CreatedAt
	trip.UpdatedAt = time.Now()

	// Validate
	if err := s.validateTrip(trip); err != nil {
		return nil, err
	}

	// Update in database
	if err := s.tripRepo.Update(ctx, trip); err != nil {
		return nil, err
	}

	return trip, nil
}

func (s *tripService) MigrateShadowTrips(ctx context.Context, shadowUserID, userID string) error {
	return s.tripRepo.MigrateShadowTrips(ctx, shadowUserID, userID)
}
