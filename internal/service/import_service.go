package service

import (
	"context"
	"fmt"
	"triply-server/internal/dto"
	"triply-server/internal/models"
	"triply-server/internal/repository"
	"triply-server/internal/utils"

	"github.com/google/uuid"
)

// ImportService defines the interface for import operations
type ImportService interface {
	ImportTripParts(ctx context.Context, userID string, req *dto.ImportTripRequest) (*dto.ImportTripResponse, error)
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

func (s *importService) ImportTripParts(ctx context.Context, userID string, req *dto.ImportTripRequest) (*dto.ImportTripResponse, error) {
	// 1. Fetch source public trip
	publicTrip, err := s.publicTripRepo.FindByID(ctx, req.SourceTripID)
	if err != nil {
		return nil, utils.NewNotFoundError("Source trip")
	}

	// 2. Get full source trip with details
	sourceTrip := publicTrip.Trip
	if sourceTrip == nil {
		return nil, fmt.Errorf("source trip not fully loaded")
	}

	// 3. Extract selected parts
	selectedDays := s.extractSelectedDays(sourceTrip, &req.Selection)

	// 4. Find target destination in user's trip
	// For now, we'll need the full target trip ID - this should be passed in the request
	// This is a simplified implementation
	// In a full implementation, you'd need to pass the target trip ID as well

	// Generate import ID for tracking
	importID := uuid.New().String()

	// Tag imported activities with metadata
	for i := range selectedDays {
		for j := range selectedDays[i].Activities {
			selectedDays[i].Activities[j].ID = utils.GenerateID("act")
			// Could add metadata field to track import source if model supports it
		}
	}

	// 5. Return the selected parts
	// Note: Actual insertion into target trip should be done by the trip service
	// This service just extracts the parts to be imported

	return &dto.ImportTripResponse{
		ImportID: importID,
		UpdatedTrip: models.Trip{
			ID: req.SourceTripID,
			// This would contain the updated trip after import
			// Implementation depends on how we want to handle the merge
		},
	}, nil
}

func (s *importService) extractSelectedDays(sourceTrip *models.Trip, selection *dto.ImportSelection) []models.DayPlan {
	selectedDays := []models.DayPlan{}

	if len(selection.DayIDs) == 0 && len(selection.ActivityIDs) == 0 {
		return selectedDays
	}

	dayIDSet := make(map[string]bool)
	for _, id := range selection.DayIDs {
		dayIDSet[id] = true
	}

	activityIDSet := make(map[string]bool)
	for _, id := range selection.ActivityIDs {
		activityIDSet[id] = true
	}

	// Extract days and activities
	for _, dest := range sourceTrip.Destinations {
		for _, day := range dest.DailyPlans {
			// If day is selected or has selected activities
			if dayIDSet[day.ID] {
				selectedDays = append(selectedDays, day)
			} else if len(activityIDSet) > 0 {
				// Check if any activities from this day are selected
				filteredActivities := []models.Activity{}
				for _, act := range day.Activities {
					if activityIDSet[act.ID] {
						filteredActivities = append(filteredActivities, act)
					}
				}
				if len(filteredActivities) > 0 {
					dayClone := day
					dayClone.Activities = filteredActivities
					selectedDays = append(selectedDays, dayClone)
				}
			}
		}
	}

	return selectedDays
}
