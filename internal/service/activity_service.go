package service

import (
	"context"
	"triply-server/internal/models"
	"triply-server/internal/repository"
)

// ActivityService defines the interface for activity operations
type ActivityService interface {
	GetActivitiesByDayPlan(ctx context.Context, dayPlanID string) ([]models.Activity, error)
	UpdateActivityOrders(ctx context.Context, dayPlanID string, activities []models.Activity) error
}

type activityService struct {
	activityRepo repository.ActivityRepository
}

// NewActivityService creates a new activity service instance
func NewActivityService(activityRepo repository.ActivityRepository) ActivityService {
	return &activityService{activityRepo: activityRepo}
}

func (s *activityService) GetActivitiesByDayPlan(ctx context.Context, dayPlanID string) ([]models.Activity, error) {
	return s.activityRepo.FindByDayPlanID(ctx, dayPlanID)
}

func (s *activityService) UpdateActivityOrders(ctx context.Context, dayPlanID string, activities []models.Activity) error {
	// Validate all activities belong to the same day plan
	for _, activity := range activities {
		if activity.DayPlanID != dayPlanID {
			activity.DayPlanID = dayPlanID
		}
	}

	return s.activityRepo.UpdateOrders(ctx, activities)
}
