package repository

import (
	"context"
	"strings"
	"time"
	"triply-server/internal/models"

	"gorm.io/gorm"
)

// PublicTripRepository defines the interface for public trip data operations
type PublicTripRepository interface {
	FindAll(ctx context.Context, filters *PublicTripFilters) ([]models.Trip, int64, error)
	FindByID(ctx context.Context, id string) (*models.Trip, error)
	FindBySlug(ctx context.Context, slug string) (*models.Trip, error)
	ToggleVisibility(ctx context.Context, tripID string, userID string, visibility string) error
}

// PublicTripFilters holds filter criteria for listing public trips
type PublicTripFilters struct {
	Query         *string
	Cities        []string
	MinDays       *int
	MaxDays       *int
	Months        []int
	Tags          []string
	TravelerTypes []string
	Sort          string
	Page          int
	PageSize      int
}

type publicTripRepository struct {
	db *gorm.DB
}

// NewPublicTripRepository creates a new public trip repository instance
func NewPublicTripRepository(db *gorm.DB) PublicTripRepository {
	return &publicTripRepository{db: db}
}

func (r *publicTripRepository) FindAll(ctx context.Context, filters *PublicTripFilters) ([]models.Trip, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.Trip{}).
		Where("trips.visibility = ?", "public")

	// Apply filters
	if filters.Query != nil && *filters.Query != "" {
		searchTerm := "%" + strings.ToLower(*filters.Query) + "%"
		query = query.Where(
			"LOWER(trips.name) LIKE ? OR LOWER(trips.summary) LIKE ? OR LOWER(trips.description) LIKE ?",
			searchTerm, searchTerm, searchTerm,
		)
	}

	if len(filters.Cities) > 0 {
		// Match against destination cities using subquery
		query = query.Where("EXISTS (SELECT 1 FROM trip_destinations td JOIN destinations d ON td.destination_id = d.id WHERE td.trip_id = trips.id AND d.city IN ?)", filters.Cities)
	}

	if filters.MinDays != nil {
		query = query.Where("EXTRACT(DAY FROM (trips.end_date::timestamp - trips.start_date::timestamp)) + 1 >= ?", *filters.MinDays)
	}

	if filters.MaxDays != nil {
		query = query.Where("EXTRACT(DAY FROM (trips.end_date::timestamp - trips.start_date::timestamp)) + 1 <= ?", *filters.MaxDays)
	}

	if len(filters.Tags) > 0 {
		for _, tag := range filters.Tags {
			query = query.Where("EXISTS (SELECT 1 FROM unnest(trips.tags) t WHERE t = ?)", tag)
		}
	}

	if len(filters.TravelerTypes) > 0 {
		query = query.Where("trips.traveler_type IN ?", filters.TravelerTypes)
	}

	// Count total - reuse the same query conditions
	var total int64
	countQuery := r.db.WithContext(ctx).Model(&models.Trip{})

	// Apply the base condition
	countQuery = countQuery.Where("trips.visibility = ?", "public")

	// Apply same filters to count
	if filters.Query != nil && *filters.Query != "" {
		searchTerm := "%" + strings.ToLower(*filters.Query) + "%"
		countQuery = countQuery.Where(
			"LOWER(trips.name) LIKE ? OR LOWER(trips.summary) LIKE ? OR LOWER(trips.description) LIKE ?",
			searchTerm, searchTerm, searchTerm,
		)
	}
	if len(filters.Cities) > 0 {
		countQuery = countQuery.Where("EXISTS (SELECT 1 FROM trip_destinations td JOIN destinations d ON td.destination_id = d.id WHERE td.trip_id = trips.id AND d.city IN ?)", filters.Cities)
	}

	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	switch filters.Sort {
	case "mostRecent":
		query = query.Order("trips.updated_at DESC")
	case "shortest":
		query = query.Order("trips.end_date - trips.start_date ASC")
	case "longest":
		query = query.Order("trips.end_date - trips.start_date DESC")
	default: // "featured"
		query = query.Order("trips.likes DESC, trips.updated_at DESC")
	}

	// Apply pagination
	offset := (filters.Page - 1) * filters.PageSize
	query = query.Offset(offset).Limit(filters.PageSize)

	var trips []models.Trip
	if err := query.
		Select("trips.*").
		Preload("TripDestinations", func(db *gorm.DB) *gorm.DB {
			return db.Order("trip_destinations.order_index ASC")
		}).
		Preload("TripDestinations.Destination").
		Preload("DayPlans", func(db *gorm.DB) *gorm.DB {
			return db.Order("day_plans.day_number ASC")
		}).
		Find(&trips).Error; err != nil {
		return nil, 0, err
	}

	return trips, total, nil
}

func (r *publicTripRepository) FindByID(ctx context.Context, id string) (*models.Trip, error) {
	var trip models.Trip
	err := r.db.WithContext(ctx).
		Select("trips.*").
		Preload("User").
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
		Where("id = ? AND visibility = ?", id, "public").
		First(&trip).Error
	if err != nil {
		return nil, err
	}
	return &trip, nil
}

func (r *publicTripRepository) FindBySlug(ctx context.Context, slug string) (*models.Trip, error) {
	var trip models.Trip
	err := r.db.WithContext(ctx).
		Select("trips.*").
		Preload("User").
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
		Where("slug = ? AND visibility = ?", slug, "public").
		First(&trip).Error
	if err != nil {
		return nil, err
	}
	return &trip, nil
}

func (r *publicTripRepository) ToggleVisibility(ctx context.Context, tripID string, userID string, visibility string) error {
	now := time.Now().UTC()
	updates := map[string]interface{}{
		"visibility": visibility,
		"updated_at": now,
	}

	// If making public for the first time, set published_at
	if visibility == "public" {
		var trip models.Trip
		if err := r.db.WithContext(ctx).Select("published_at").Where("id = ?", tripID).First(&trip).Error; err != nil {
			return err
		}
		if trip.PublishedAt == nil {
			updates["published_at"] = now
		}
	}

	return r.db.WithContext(ctx).
		Model(&models.Trip{}).
		Where("id = ? AND user_id = ?", tripID, userID).
		Updates(updates).Error
}
