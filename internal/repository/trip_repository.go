package repository

import (
	"context"
	"triply-server/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
		Preload("Destinations").
		Preload("Destinations.DailyPlans").
		Preload("Destinations.DailyPlans.Activities").
		Find(&trips).Error
	if err != nil {
		return nil, err
	}
	return trips, nil
}

func (r *tripRepository) FindByID(ctx context.Context, tripID, userID string) (*models.Trip, error) {
	var trip models.Trip
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", tripID, userID).
		Preload("Destinations").
		Preload("Destinations.DailyPlans").
		Preload("Destinations.DailyPlans.Activities").
		First(&trip).Error
	if err != nil {
		return nil, err
	}
	return &trip, nil
}

func (r *tripRepository) Create(ctx context.Context, trip *models.Trip) error {
	return r.db.WithContext(ctx).
		Session(&gorm.Session{FullSaveAssociations: true}).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(trip).Error
}

func (r *tripRepository) Update(ctx context.Context, trip *models.Trip) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Check if trip exists
		var existing models.Trip
		if err := tx.Where("id = ? AND user_id = ?", trip.ID, trip.UserID).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// If not found, create it
				return tx.Session(&gorm.Session{FullSaveAssociations: true}).Create(trip).Error
			}
			return err
		}

		// Delete existing destinations and their children (cascade will handle the rest)
		if err := tx.Where("trip_id = ?", trip.ID).Delete(&models.Destination{}).Error; err != nil {
			return err
		}

		// Update main trip fields
		if err := tx.Model(&models.Trip{}).Where("id = ? AND user_id = ?", trip.ID, trip.UserID).Updates(map[string]interface{}{
			"name":           trip.Name,
			"description":    trip.Description,
			"traveler_count": trip.TravelerCount,
			"start_date":     trip.StartDate,
			"end_date":       trip.EndDate,
			"locale":         trip.Locale,
			"visibility":     trip.Visibility,
			"status":         trip.Status,
			"timezone":       trip.Timezone,
			"cover_image":    trip.CoverImage,
			"updated_at":     trip.UpdatedAt,
		}).Error; err != nil {
			return err
		}

		// Recreate associations
		for i := range trip.Destinations {
			trip.Destinations[i].TripID = trip.ID
			for j := range trip.Destinations[i].DailyPlans {
				trip.Destinations[i].DailyPlans[j].DestinationID = trip.Destinations[i].ID
				for k := range trip.Destinations[i].DailyPlans[j].Activities {
					trip.Destinations[i].DailyPlans[j].Activities[k].DayPlanID = trip.Destinations[i].DailyPlans[j].ID
				}
			}
		}

		if len(trip.Destinations) > 0 {
			if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).Create(&trip.Destinations).Error; err != nil {
				return err
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

func (r *tripRepository) FindByShadowUserID(ctx context.Context, shadowUserID string) ([]models.Trip, error) {
	var trips []models.Trip
	err := r.db.WithContext(ctx).
		Where("shadow_user_id = ?", shadowUserID).
		Preload("Destinations").
		Preload("Destinations.DailyPlans").
		Preload("Destinations.DailyPlans.Activities").
		Find(&trips).Error
	if err != nil {
		return nil, err
	}
	return trips, nil
}

func (r *tripRepository) FindByIDWithShadowUser(ctx context.Context, tripID, shadowUserID string) (*models.Trip, error) {
	var trip models.Trip
	err := r.db.WithContext(ctx).
		Where("id = ? AND shadow_user_id = ?", tripID, shadowUserID).
		Preload("Destinations").
		Preload("Destinations.DailyPlans").
		Preload("Destinations.DailyPlans.Activities").
		First(&trip).Error
	if err != nil {
		return nil, err
	}
	return &trip, nil
}

func (r *tripRepository) MigrateShadowTrips(ctx context.Context, shadowUserID, userID string) error {
	// Update all trips with shadow_user_id to have user_id and clear shadow_user_id
	return r.db.WithContext(ctx).
		Model(&models.Trip{}).
		Where("shadow_user_id = ?", shadowUserID).
		Updates(map[string]interface{}{
			"user_id":        userID,
			"shadow_user_id": nil,
		}).Error
}
