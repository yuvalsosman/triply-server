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

// ImportService defines the interface for importing trip data
type ImportService interface {
	ImportTripParts(ctx context.Context, userID string, req *dto.ImportTripRequest) (*models.Trip, error)
}

type importService struct {
	publicTripRepo repository.PublicTripRepository
	tripRepo       repository.TripRepository
}

// NewImportService creates a new import service instance
func NewImportService(publicTripRepo repository.PublicTripRepository, tripRepo repository.TripRepository) ImportService {
	return &importService{
		publicTripRepo: publicTripRepo,
		tripRepo:       tripRepo,
	}
}

func (s *importService) ImportTripParts(ctx context.Context, userID string, req *dto.ImportTripRequest) (*models.Trip, error) {
	// Get the public trip (source)
	publicTrip, err := s.publicTripRepo.FindByID(ctx, req.SourceTripID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, utils.NewNotFoundError("Public trip")
		}
		return nil, err
	}

	// For MVP, create a new trip that's a copy of the public trip
	// TODO: Implement selective day/activity import based on req.Selection and req.Target
	now := time.Now()
	newTrip := &models.Trip{
		ID:            utils.GenerateID("trip"),
		UserID:        userID,
		Name:          "Copy of " + publicTrip.Name,
		Description:   publicTrip.Description,
		TravelerCount: publicTrip.TravelerCount,
		StartDate:     publicTrip.StartDate,
		EndDate:       publicTrip.EndDate,
		Timezone:      publicTrip.Timezone,
		CoverImage:    publicTrip.HeroImage,
		Visibility:    "private",
		Status:        "planning",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.tripRepo.Create(ctx, newTrip); err != nil {
		return nil, err
	}

	return newTrip, nil
}
