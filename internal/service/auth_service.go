package service

import (
	"context"
	"strings"
	"time"
	"triply-server/internal/models"
	"triply-server/internal/repository"
	"triply-server/internal/utils"

	"golang.org/x/oauth2"
)

// AuthService defines the interface for authentication operations
type AuthService interface {
	GetOrCreateUserFromGoogle(ctx context.Context, googleUser *GoogleUserInfo) (*models.User, error)
	GetUserByID(ctx context.Context, userID string) (*models.User, error)
	GenerateToken(user *models.User, jwtSecret string) (string, error)
	UpdateDisplayName(ctx context.Context, userID, displayName string) (*models.User, error)
}

// GoogleUserInfo represents user info from Google OAuth
type GoogleUserInfo struct {
	ID            string
	Email         string
	VerifiedEmail bool
	Name          string
	GivenName     string
	FamilyName    string
	Picture       string
	Locale        string
}

type authService struct {
	userRepo repository.UserRepository
}

// NewAuthService creates a new auth service instance
func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &authService{userRepo: userRepo}
}

func (s *authService) GetOrCreateUserFromGoogle(ctx context.Context, googleUser *GoogleUserInfo) (*models.User, error) {
	// Try to find existing user by Google ID
	existingUser, err := s.userRepo.FindByGoogleID(ctx, googleUser.ID)
	if err == nil {
		// Update user info
		existingUser.Name = nonEmpty(googleUser.Name, googleUser.Email)
		existingUser.UpdatedAt = time.Now()
		if err := s.userRepo.Update(ctx, existingUser); err != nil {
			return nil, err
		}
		return existingUser, nil
	}

	// Try to find by email
	existingUser, err = s.userRepo.FindByEmail(ctx, googleUser.Email)
	if err == nil {
		// Update with Google ID
		gid := googleUser.ID
		existingUser.GoogleID = &gid
		existingUser.Name = nonEmpty(googleUser.Name, googleUser.Email)
		existingUser.UpdatedAt = time.Now()
		if err := s.userRepo.Update(ctx, existingUser); err != nil {
			return nil, err
		}
		return existingUser, nil
	}

	// Create new user
	now := time.Now()
	googleID := googleUser.ID
	user := &models.User{
		ID:        "user-" + googleUser.ID,
		GoogleID:  &googleID,
		Name:      nonEmpty(googleUser.Name, googleUser.Email),
		Email:     googleUser.Email,
		Locale:    fallbackLocale(googleUser.Locale),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *authService) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	return s.userRepo.FindByID(ctx, userID)
}

func (s *authService) GenerateToken(user *models.User, jwtSecret string) (string, error) {
	return utils.GenerateJWT(user.ID, user.Email, jwtSecret)
}

// Helper functions

func nonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func fallbackLocale(loc string) string {
	if loc == "he" || strings.HasPrefix(loc, "he") {
		return "he"
	}
	return "en"
}

// ExchangeCodeForToken exchanges OAuth code for token
func ExchangeCodeForToken(config *oauth2.Config, code string) (*oauth2.Token, error) {
	return config.Exchange(context.Background(), code)
}

// UpdateDisplayName updates the user's display name
func (s *authService) UpdateDisplayName(ctx context.Context, userID, displayName string) (*models.User, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	user.DisplayName = &displayName
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}
