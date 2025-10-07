package service

import (
	"context"
	"triply-server/internal/dto"
	"triply-server/internal/repository"
)

// TripLikeService defines the interface for trip like business logic
type TripLikeService interface {
	ToggleLike(ctx context.Context, userID, tripID string) (*dto.LikeToggleResponse, error)
}

type tripLikeService struct {
	tripLikeRepo repository.TripLikeRepository
}

// NewTripLikeService creates a new trip like service instance
func NewTripLikeService(tripLikeRepo repository.TripLikeRepository) TripLikeService {
	return &tripLikeService{
		tripLikeRepo: tripLikeRepo,
	}
}

func (s *tripLikeService) ToggleLike(ctx context.Context, userID, tripID string) (*dto.LikeToggleResponse, error) {
	liked, totalLikes, err := s.tripLikeRepo.ToggleLike(ctx, userID, tripID)
	if err != nil {
		return nil, err
	}

	return &dto.LikeToggleResponse{
		Liked:      liked,
		TotalLikes: totalLikes,
	}, nil
}
