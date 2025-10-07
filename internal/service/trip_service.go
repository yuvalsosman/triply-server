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
	MigrateShadowTrips(ctx context.Context, shadowUserID, userID string) error
	GetShadowUserTrips(ctx context.Context, shadowUserID string) ([]models.Trip, error)
	CreateShadowTrip(ctx context.Context, trip *models.Trip, shadowUserID string) (*models.Trip, error)
	UpdateShadowTrip(ctx context.Context, trip *models.Trip, shadowUserID string) (*models.Trip, error)
	DeleteShadowTrip(ctx context.Context, tripID, shadowUserID string) error
}

type tripService struct {
	tripRepo repository.TripRepository
}

// NewTripService creates a new trip service instance
func NewTripService(tripRepo repository.TripRepository) TripService {
	return &tripService{tripRepo: tripRepo}
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
