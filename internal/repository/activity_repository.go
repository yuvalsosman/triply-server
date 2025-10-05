package repository

import (
	"context"
	"triply-server/internal/models"

	"gorm.io/gorm"
)

// ActivityRepository defines the interface for activity data operations
type ActivityRepository interface {
	FindByDayPlanID(ctx context.Context, dayPlanID string) ([]models.Activity, error)
	UpdateOrders(ctx context.Context, activities []models.Activity) error
}

type activityRepository struct {
	db *gorm.DB
}

// NewActivityRepository creates a new activity repository instance
func NewActivityRepository(db *gorm.DB) ActivityRepository {
	return &activityRepository{db: db}
}

func (r *activityRepository) FindByDayPlanID(ctx context.Context, dayPlanID string) ([]models.Activity, error) {
	var activities []models.Activity
	err := r.db.WithContext(ctx).
		Where("day_plan_id = ?", dayPlanID).
		Order("time_of_day, \"order\"").
		Find(&activities).Error
	if err != nil {
		return nil, err
	}
	return activities, nil
}

func (r *activityRepository) UpdateOrders(ctx context.Context, activities []models.Activity) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, activity := range activities {
			if err := tx.Model(&models.Activity{}).
				Where("id = ?", activity.ID).
				Updates(map[string]interface{}{
					"time_of_day": activity.TimeOfDay,
					"order":       activity.Order,
				}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
