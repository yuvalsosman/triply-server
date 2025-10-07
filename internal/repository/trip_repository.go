package repository

import (
	"context"
	"triply-server/internal/models"

	"gorm.io/gorm"
)

// TripRepository defines the interface for trip data operations
type TripRepository interface {
	FindByUserID(ctx context.Context, userID string) ([]models.Trip, error)
	FindByShadowUserID(ctx context.Context, shadowUserID string) ([]models.Trip, error)
	FindByID(ctx context.Context, tripID, userID string) (*models.Trip, error)
	FindByIDWithShadowUser(ctx context.Context, tripID, shadowUserID string) (*models.Trip, error)
	Create(ctx context.Context, trip *models.Trip) error
	Update(ctx context.Context, trip *models.Trip) error
	Delete(ctx context.Context, tripID, userID string) error
	MigrateShadowTrips(ctx context.Context, shadowUserID, userID string) error
}

type tripRepository struct {
	db *gorm.DB
}

// NewTripRepository creates a new trip repository instance
func NewTripRepository(db *gorm.DB) TripRepository {
	return &tripRepository{db: db}
}

func (r *tripRepository) FindByUserID(ctx context.Context, userID string) ([]models.Trip, error) {
	var trips []models.Trip
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Preload("TripDestinations", func(db *gorm.DB) *gorm.DB {
			return db.Order("trip_destinations.order_index ASC")
		}).
		Preload("TripDestinations.Destination").
		Preload("DayPlans", func(db *gorm.DB) *gorm.DB {
			return db.Order("day_plans.day_number ASC")
		}).
		Preload("DayPlans.DayPlanDestinations", func(db *gorm.DB) *gorm.DB {
			return db.Order("day_plan_destinations.order_index ASC")
		}).
		Preload("DayPlans.DayPlanDestinations.Destination").
		Preload("DayPlans.DayPlanActivities", func(db *gorm.DB) *gorm.DB {
			return db.Order("day_plan_activities.time_of_day, day_plan_activities.order_within_time ASC")
		}).
		Preload("DayPlans.DayPlanActivities.Activity").
		Order("trips.updated_at DESC").
		Find(&trips).Error
	if err != nil {
		return nil, err
	}
	return trips, nil
}

func (r *tripRepository) FindByShadowUserID(ctx context.Context, shadowUserID string) ([]models.Trip, error) {
	// Shadow trips are stored with the shadow user ID as the user_id
	return r.FindByUserID(ctx, shadowUserID)
}

func (r *tripRepository) FindByID(ctx context.Context, tripID, userID string) (*models.Trip, error) {
	var trip models.Trip
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", tripID, userID).
		Preload("TripDestinations", func(db *gorm.DB) *gorm.DB {
			return db.Order("trip_destinations.order_index ASC")
		}).
		Preload("TripDestinations.Destination").
		Preload("DayPlans", func(db *gorm.DB) *gorm.DB {
			return db.Order("day_plans.day_number ASC")
		}).
		Preload("DayPlans.DayPlanDestinations", func(db *gorm.DB) *gorm.DB {
			return db.Order("day_plan_destinations.order_index ASC")
		}).
		Preload("DayPlans.DayPlanDestinations.Destination").
		Preload("DayPlans.DayPlanActivities", func(db *gorm.DB) *gorm.DB {
			return db.Order("day_plan_activities.time_of_day, day_plan_activities.order_within_time ASC")
		}).
		Preload("DayPlans.DayPlanActivities.Activity").
		First(&trip).Error
	if err != nil {
		return nil, err
	}
	return &trip, nil
}

func (r *tripRepository) FindByIDWithShadowUser(ctx context.Context, tripID, shadowUserID string) (*models.Trip, error) {
	return r.FindByID(ctx, tripID, shadowUserID)
}

func (r *tripRepository) Create(ctx context.Context, trip *models.Trip) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create the trip
		if err := tx.Create(trip).Error; err != nil {
			return err
		}

		// If trip has day plans with nested data, save them
		if len(trip.DayPlans) > 0 {
			for i := range trip.DayPlans {
				trip.DayPlans[i].TripID = trip.ID
				if err := tx.Create(&trip.DayPlans[i]).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func (r *tripRepository) Update(ctx context.Context, trip *models.Trip) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Update trip basic info
		if err := tx.Model(&models.Trip{}).
			Where("id = ? AND user_id = ?", trip.ID, trip.UserID).
			Updates(trip).Error; err != nil {
			return err
		}

		// Delete existing day plans and related data (cascade will handle activities)
		if err := tx.Where("trip_id = ?", trip.ID).Delete(&models.DayPlan{}).Error; err != nil {
			return err
		}

		// Delete existing trip destinations
		if err := tx.Where("trip_id = ?", trip.ID).Delete(&models.TripDestination{}).Error; err != nil {
			return err
		}

		// Recreate day plans if provided
		if len(trip.DayPlans) > 0 {
			for i := range trip.DayPlans {
				trip.DayPlans[i].TripID = trip.ID
				if err := tx.Create(&trip.DayPlans[i]).Error; err != nil {
					return err
				}
			}
		}

		// Recreate trip destinations if provided
		if len(trip.TripDestinations) > 0 {
			for i := range trip.TripDestinations {
				trip.TripDestinations[i].TripID = trip.ID
				if err := tx.Create(&trip.TripDestinations[i]).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func (r *tripRepository) Delete(ctx context.Context, tripID, userID string) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", tripID, userID).
		Delete(&models.Trip{}).Error
}

func (r *tripRepository) MigrateShadowTrips(ctx context.Context, shadowUserID, userID string) error {
	return r.db.WithContext(ctx).
		Model(&models.Trip{}).
		Where("user_id = ?", shadowUserID).
		Update("user_id", userID).Error
}
