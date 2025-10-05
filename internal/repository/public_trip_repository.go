package repository

import (
	"context"
	"strings"
	"triply-server/internal/models"

	"gorm.io/gorm"
)

// PublicTripRepository defines the interface for public trip data operations
type PublicTripRepository interface {
	FindAll(ctx context.Context, filters *PublicTripFilters) ([]models.PublicTrip, int64, error)
	FindByID(ctx context.Context, id string) (*models.PublicTrip, error)
	FindByTripID(ctx context.Context, tripID string) (*models.PublicTrip, error)
	Create(ctx context.Context, publicTrip *models.PublicTrip) error
	Update(ctx context.Context, publicTrip *models.PublicTrip) error
	Delete(ctx context.Context, id string) error
}

// PublicTripFilters holds filter criteria for listing public trips
type PublicTripFilters struct {
	Query         *string
	Cities        []string
	MinDays       *int
	MaxDays       *int
	Months        []int
	Seasons       []string
	BudgetLevels  []string
	Paces         []string
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

func (r *publicTripRepository) FindAll(ctx context.Context, filters *PublicTripFilters) ([]models.PublicTrip, int64, error) {
	query := r.db.WithContext(ctx).
		Joins("JOIN trips ON trips.id = public_trips.trip_id").
		Where("trips.visibility = ?", "public")

	// Apply filters
	if filters.Query != nil && *filters.Query != "" {
		searchTerm := "%" + strings.ToLower(*filters.Query) + "%"
		query = query.Where(
			"LOWER(trips.name) LIKE ? OR LOWER(public_trips.summary) LIKE ?",
			searchTerm, searchTerm,
		)
	}

	if len(filters.Cities) > 0 {
		// Match against origin_cities array
		for _, city := range filters.Cities {
			query = query.Where("public_trips.origin_cities LIKE ?", "%"+city+"%")
		}
	}

	if filters.MinDays != nil {
		query = query.Where("public_trips.duration_days >= ?", *filters.MinDays)
	}

	if filters.MaxDays != nil {
		query = query.Where("public_trips.duration_days <= ?", *filters.MaxDays)
	}

	if len(filters.Months) > 0 {
		query = query.Where("public_trips.start_month IN ?", filters.Months)
	}

	if len(filters.Seasons) > 0 {
		for _, season := range filters.Seasons {
			query = query.Where("public_trips.seasons LIKE ?", "%"+season+"%")
		}
	}

	if len(filters.BudgetLevels) > 0 {
		query = query.Where("public_trips.budget_level IN ?", filters.BudgetLevels)
	}

	if len(filters.Paces) > 0 {
		query = query.Where("public_trips.pace IN ?", filters.Paces)
	}

	if len(filters.Tags) > 0 {
		for _, tag := range filters.Tags {
			query = query.Where("public_trips.tags LIKE ?", "%"+tag+"%")
		}
	}

	if len(filters.TravelerTypes) > 0 {
		for _, ttype := range filters.TravelerTypes {
			query = query.Where("public_trips.traveler_types LIKE ?", "%"+ttype+"%")
		}
	}

	// Count total
	var total int64
	if err := query.Model(&models.PublicTrip{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	switch filters.Sort {
	case "mostRecent":
		query = query.Order("public_trips.updated_at DESC")
	case "shortest":
		query = query.Order("public_trips.duration_days ASC")
	case "longest":
		query = query.Order("public_trips.duration_days DESC")
	default: // "featured"
		query = query.Order("public_trips.likes DESC, public_trips.updated_at DESC")
	}

	// Apply pagination
	offset := (filters.Page - 1) * filters.PageSize
	query = query.Offset(offset).Limit(filters.PageSize)

	var trips []models.PublicTrip
	if err := query.Preload("Trip").Find(&trips).Error; err != nil {
		return nil, 0, err
	}

	return trips, total, nil
}

func (r *publicTripRepository) FindByID(ctx context.Context, id string) (*models.PublicTrip, error) {
	var trip models.PublicTrip
	err := r.db.WithContext(ctx).
		Preload("Trip").
		Preload("Trip.Destinations").
		Preload("Trip.Destinations.DailyPlans").
		Preload("Trip.Destinations.DailyPlans.Activities").
		Where("id = ?", id).
		First(&trip).Error
	if err != nil {
		return nil, err
	}
	return &trip, nil
}

func (r *publicTripRepository) FindByTripID(ctx context.Context, tripID string) (*models.PublicTrip, error) {
	var trip models.PublicTrip
	err := r.db.WithContext(ctx).
		Preload("Trip").
		Where("trip_id = ?", tripID).
		First(&trip).Error
	if err != nil {
		return nil, err
	}
	return &trip, nil
}

func (r *publicTripRepository) Create(ctx context.Context, publicTrip *models.PublicTrip) error {
	return r.db.WithContext(ctx).Create(publicTrip).Error
}

func (r *publicTripRepository) Update(ctx context.Context, publicTrip *models.PublicTrip) error {
	return r.db.WithContext(ctx).Save(publicTrip).Error
}

func (r *publicTripRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.PublicTrip{}).Error
}
