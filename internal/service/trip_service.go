package service

import (
	"context"
	"time"
	"triply-server/internal/models"
	"triply-server/internal/repository"
	"triply-server/internal/utils"

	"gorm.io/gorm"
)

// TripService defines the interface for trip operations
type TripService interface {
	ListTrips(ctx context.Context, userID string) ([]models.Trip, error)
	GetTrip(ctx context.Context, tripID, userID string) (*models.Trip, error)
	CreateTrip(ctx context.Context, trip *models.Trip) (*models.Trip, error)
	UpdateTrip(ctx context.Context, trip *models.Trip) (*models.Trip, error)
	DeleteTrip(ctx context.Context, tripID, userID string) error
	ClonePublicTrip(ctx context.Context, publicTripID, userID, newTripName string) (*models.Trip, error)
	MigrateShadowTrips(ctx context.Context, shadowUserID, userID string) error
	GetShadowUserTrips(ctx context.Context, shadowUserID string) ([]models.Trip, error)
	CreateShadowTrip(ctx context.Context, trip *models.Trip, shadowUserID string) (*models.Trip, error)
	UpdateShadowTrip(ctx context.Context, trip *models.Trip, shadowUserID string) (*models.Trip, error)
	DeleteShadowTrip(ctx context.Context, tripID, shadowUserID string) error
}

type tripService struct {
	tripRepo       repository.TripRepository
	publicTripRepo repository.PublicTripRepository
}

// NewTripService creates a new trip service instance
func NewTripService(tripRepo repository.TripRepository, publicTripRepo repository.PublicTripRepository) TripService {
	return &tripService{
		tripRepo:       tripRepo,
		publicTripRepo: publicTripRepo,
	}
}

func (s *tripService) ListTrips(ctx context.Context, userID string) ([]models.Trip, error) {
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

func (s *tripService) CreateTrip(ctx context.Context, trip *models.Trip) (*models.Trip, error) {
	now := time.Now()

	// Generate ID if not provided
	if trip.ID == "" {
		trip.ID = utils.GenerateID("trip")
	}

	trip.CreatedAt = now
	trip.UpdatedAt = now

	// Set default visibility
	if trip.Visibility == "" {
		trip.Visibility = "private"
	}

	// Set default status
	if trip.Status == "" {
		trip.Status = "planning"
	}

	// Generate IDs for day plans if provided
	for i := range trip.DayPlans {
		if trip.DayPlans[i].ID == "" {
			trip.DayPlans[i].ID = utils.GenerateID("day")
		}
		trip.DayPlans[i].TripID = trip.ID
		trip.DayPlans[i].CreatedAt = now
		trip.DayPlans[i].UpdatedAt = now
	}

	if err := s.tripRepo.Create(ctx, trip); err != nil {
		return nil, err
	}

	return trip, nil
}

func (s *tripService) UpdateTrip(ctx context.Context, trip *models.Trip) (*models.Trip, error) {
	trip.UpdatedAt = time.Now()

	// Generate IDs for day plans if provided
	now := time.Now()
	for i := range trip.DayPlans {
		if trip.DayPlans[i].ID == "" {
			trip.DayPlans[i].ID = utils.GenerateID("day")
		}
		trip.DayPlans[i].TripID = trip.ID
		trip.DayPlans[i].UpdatedAt = now
	}

	if err := s.tripRepo.Update(ctx, trip); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, utils.NewNotFoundError("Trip")
		}
		return nil, err
	}

	return trip, nil
}

func (s *tripService) DeleteTrip(ctx context.Context, tripID, userID string) error {
	return s.tripRepo.Delete(ctx, tripID, userID)
}

func (s *tripService) MigrateShadowTrips(ctx context.Context, shadowUserID, userID string) error {
	return s.tripRepo.MigrateShadowTrips(ctx, shadowUserID, userID)
}

func (s *tripService) GetShadowUserTrips(ctx context.Context, shadowUserID string) ([]models.Trip, error) {
	return s.tripRepo.FindByShadowUserID(ctx, shadowUserID)
}

func (s *tripService) CreateShadowTrip(ctx context.Context, trip *models.Trip, shadowUserID string) (*models.Trip, error) {
	// Shadow trips use shadow user ID as the user_id
	trip.UserID = shadowUserID
	return s.CreateTrip(ctx, trip)
}

func (s *tripService) UpdateShadowTrip(ctx context.Context, trip *models.Trip, shadowUserID string) (*models.Trip, error) {
	trip.UserID = shadowUserID
	return s.UpdateTrip(ctx, trip)
}

func (s *tripService) DeleteShadowTrip(ctx context.Context, tripID, shadowUserID string) error {
	return s.tripRepo.Delete(ctx, tripID, shadowUserID)
}

func (s *tripService) ClonePublicTrip(ctx context.Context, publicTripID, userID, newTripName string) (*models.Trip, error) {
	// 1. Fetch the original public trip with all nested data
	originalTrip, err := s.publicTripRepo.FindByID(ctx, publicTripID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, utils.NewNotFoundError("Public trip")
		}
		return nil, err
	}

	// 2. Verify the trip is public (only public trips can be cloned)
	if originalTrip.Visibility != "public" {
		return nil, utils.NewValidationError("Only public trips can be cloned")
	}

	// 3. Create a new trip structure with cloned data
	now := time.Now()
	clonedTrip := &models.Trip{
		ID:            utils.GenerateID("trip"),
		UserID:        userID,
		Name:          newTripName,
		Description:   originalTrip.Description,
		TravelerCount: originalTrip.TravelerCount,
		StartDate:     originalTrip.StartDate,
		EndDate:       originalTrip.EndDate,
		Timezone:      originalTrip.Timezone,
		CoverImage:    originalTrip.CoverImage,
		HeroImage:     originalTrip.HeroImage,
		Visibility:    "private",  // Always start as private
		Status:        "planning", // Reset to planning status
		TravelerType:  originalTrip.TravelerType,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// 4. Clone all day plans with their nested data
	clonedTrip.DayPlans = make([]models.DayPlan, len(originalTrip.DayPlans))
	for i, originalDay := range originalTrip.DayPlans {
		clonedDay := models.DayPlan{
			ID:        utils.GenerateID("day"),
			TripID:    clonedTrip.ID,
			Date:      originalDay.Date,
			DayNumber: originalDay.DayNumber,
			Notes:     originalDay.Notes,
			CreatedAt: now,
			UpdatedAt: now,
		}

		// Clone day plan destinations
		if len(originalDay.DayPlanDestinations) > 0 {
			clonedDay.DayPlanDestinations = make([]models.DayPlanDestination, len(originalDay.DayPlanDestinations))
			for j, originalDest := range originalDay.DayPlanDestinations {
				clonedDay.DayPlanDestinations[j] = models.DayPlanDestination{
					ID:            utils.GenerateID("dpd"),
					DayPlanID:     clonedDay.ID,
					DestinationID: originalDest.DestinationID,
					OrderIndex:    originalDest.OrderIndex,
				}
			}
		}

		// Clone day plan activities
		if len(originalDay.DayPlanActivities) > 0 {
			clonedDay.DayPlanActivities = make([]models.DayPlanActivity, len(originalDay.DayPlanActivities))
			for j, originalAct := range originalDay.DayPlanActivities {
				clonedDay.DayPlanActivities[j] = models.DayPlanActivity{
					ID:              utils.GenerateID("dpa"),
					DayPlanID:       clonedDay.ID,
					ActivityID:      originalAct.ActivityID,
					TimeOfDay:       originalAct.TimeOfDay,
					OrderWithinTime: originalAct.OrderWithinTime,
				}
			}
		}

		clonedTrip.DayPlans[i] = clonedDay
	}

	// 5. Clone trip destinations
	if len(originalTrip.TripDestinations) > 0 {
		clonedTrip.TripDestinations = make([]models.TripDestination, len(originalTrip.TripDestinations))
		for i, originalTripDest := range originalTrip.TripDestinations {
			clonedTrip.TripDestinations[i] = models.TripDestination{
				ID:            utils.GenerateID("td"),
				TripID:        clonedTrip.ID,
				DestinationID: originalTripDest.DestinationID,
				OrderIndex:    originalTripDest.OrderIndex,
			}
		}
	}

	// 6. Save the cloned trip (this will cascade to all nested entities)
	if err := s.tripRepo.Create(ctx, clonedTrip); err != nil {
		return nil, err
	}

	// 7. Increment the clone count on the original trip (async, don't block on error)
	go func() {
		if err := s.publicTripRepo.IncrementCloneCount(context.Background(), publicTripID); err != nil {
			// Log error but don't fail the clone operation
			// TODO: Add proper logging
		}
	}()

	return clonedTrip, nil
}
