package repository

import (
	"context"
	"triply-server/internal/models"

	"gorm.io/gorm"
)

// ActivityRepository defines the interface for activity data operations
type ActivityRepository interface {
	FindByDayPlanID(ctx context.Context, dayPlanID string) ([]models.DayPlanActivity, error)
	UpdateOrders(ctx context.Context, dayPlanActivities []models.DayPlanActivity) error
}

type activityRepository struct {
	db *gorm.DB
}

// NewActivityRepository creates a new activity repository instance
func NewActivityRepository(db *gorm.DB) ActivityRepository {
	return &activityRepository{db: db}
}

func (r *activityRepository) FindByDayPlanID(ctx context.Context, dayPlanID string) ([]models.DayPlanActivity, error) {
	var dayPlanActivities []models.DayPlanActivity
	err := r.db.WithContext(ctx).
		Preload("Activity").
		Where("day_plan_id = ?", dayPlanID).
		Order("time_of_day, order_within_time").
		Find(&dayPlanActivities).Error
	if err != nil {
		return nil, err
	}
	return dayPlanActivities, nil
}

func (r *activityRepository) UpdateOrders(ctx context.Context, dayPlanActivities []models.DayPlanActivity) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, dpa := range dayPlanActivities {
			if err := tx.Model(&models.DayPlanActivity{}).
				Where("id = ?", dpa.ID).
				Updates(map[string]interface{}{
					"time_of_day":       dpa.TimeOfDay,
					"order_within_time": dpa.OrderWithinTime,
				}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
