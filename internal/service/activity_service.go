package service

import (
	"context"
	"triply-server/internal/models"
	"triply-server/internal/repository"
)

// ActivityService defines the interface for activity operations
type ActivityService interface {
	GetActivitiesByDayPlan(ctx context.Context, dayPlanID string) ([]models.DayPlanActivity, error)
	UpdateActivityOrders(ctx context.Context, dayPlanID string, dayPlanActivities []models.DayPlanActivity) error
}

type activityService struct {
	activityRepo repository.ActivityRepository
}

// NewActivityService creates a new activity service instance
func NewActivityService(activityRepo repository.ActivityRepository) ActivityService {
	return &activityService{activityRepo: activityRepo}
}

func (s *activityService) GetActivitiesByDayPlan(ctx context.Context, dayPlanID string) ([]models.DayPlanActivity, error) {
	return s.activityRepo.FindByDayPlanID(ctx, dayPlanID)
}

func (s *activityService) UpdateActivityOrders(ctx context.Context, dayPlanID string, dayPlanActivities []models.DayPlanActivity) error {
	// Ensure all activities belong to the specified day plan
	for i := range dayPlanActivities {
		dayPlanActivities[i].DayPlanID = dayPlanID
	}

	return s.activityRepo.UpdateOrders(ctx, dayPlanActivities)
}
