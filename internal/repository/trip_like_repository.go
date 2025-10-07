package repository

import (
	"context"
	"fmt"
	"math/rand"
	"time"
	"triply-server/internal/models"

	"gorm.io/gorm"
)

// TripLikeRepository defines the interface for trip like operations
type TripLikeRepository interface {
	// Check if user liked a trip
	HasLiked(ctx context.Context, userID, tripID string) (bool, error)

	// Get all trip IDs liked by user (for batch checking)
	GetUserLikedTripIDs(ctx context.Context, userID string, tripIDs []string) ([]string, error)

	// Toggle like (returns true if liked, false if unliked, and the new total count)
	ToggleLike(ctx context.Context, userID, tripID string) (bool, int, error)
}

type tripLikeRepository struct {
	db *gorm.DB
}

// NewTripLikeRepository creates a new trip like repository instance
func NewTripLikeRepository(db *gorm.DB) TripLikeRepository {
	return &tripLikeRepository{db: db}
}

func (r *tripLikeRepository) HasLiked(ctx context.Context, userID, tripID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.TripLike{}).
		Where("user_id = ? AND trip_id = ?", userID, tripID).
		Count(&count).Error

	return count > 0, err
}

func (r *tripLikeRepository) GetUserLikedTripIDs(ctx context.Context, userID string, tripIDs []string) ([]string, error) {
	if len(tripIDs) == 0 {
		return []string{}, nil
	}

	var likes []models.TripLike
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND trip_id IN ?", userID, tripIDs).
		Select("trip_id").
		Find(&likes).Error

	if err != nil {
		return nil, err
	}

	result := make([]string, len(likes))
	for i, like := range likes {
		result[i] = like.TripID
	}
	return result, nil
}

func (r *tripLikeRepository) ToggleLike(ctx context.Context, userID, tripID string) (bool, int, error) {
	var liked bool
	var totalLikes int

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Check if already liked
		var existingLike models.TripLike
		err := tx.Where("user_id = ? AND trip_id = ?", userID, tripID).
			First(&existingLike).Error

		if err == gorm.ErrRecordNotFound {
			// Create new like
			like := models.TripLike{
				ID:     generateLikeID(),
				UserID: userID,
				TripID: tripID,
			}
			if err := tx.Create(&like).Error; err != nil {
				return err
			}

			// Increment trip likes count
			if err := tx.Model(&models.Trip{}).
				Where("id = ?", tripID).
				UpdateColumn("likes", gorm.Expr("likes + 1")).Error; err != nil {
				return err
			}

			liked = true
		} else if err != nil {
			return err
		} else {
			// Unlike: delete existing like
			if err := tx.Delete(&existingLike).Error; err != nil {
				return err
			}

			// Decrement trip likes count (but don't go below 0)
			if err := tx.Model(&models.Trip{}).
				Where("id = ? AND likes > 0", tripID).
				UpdateColumn("likes", gorm.Expr("likes - 1")).Error; err != nil {
				return err
			}

			liked = false
		}

		// Get updated total likes count
		var trip models.Trip
		if err := tx.Select("likes").Where("id = ?", tripID).First(&trip).Error; err != nil {
			return err
		}
		totalLikes = trip.Likes

		return nil
	})

	return liked, totalLikes, err
}

// Helper function to generate like IDs
func generateLikeID() string {
	return fmt.Sprintf("like-%s", generateRandomString(16))
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
