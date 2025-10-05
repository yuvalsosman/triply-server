package dto

import "triply-server/internal/models"

// LoginResponse represents the response after successful login
type LoginResponse struct {
	User  models.User `json:"user"`
	Token string      `json:"token,omitempty"`
}

// RefreshTokenRequest represents a token refresh request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// RefreshTokenResponse represents a token refresh response
type RefreshTokenResponse struct {
	Token string `json:"token"`
}

// UpdateProfileRequest represents a profile update request
type UpdateProfileRequest struct {
	DisplayName string `json:"displayName"`
}

// UpdateProfileResponse represents the response after updating profile
type UpdateProfileResponse struct {
	User models.User `json:"user"`
}
