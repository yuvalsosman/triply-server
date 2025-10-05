package models

import "time"

// Trip represents a travel itinerary
type Trip struct {
	ID            string        `json:"id" gorm:"primaryKey;size:64"`
	UserID        string        `json:"-" gorm:"index"`
	ShadowUserID  *string       `json:"-" gorm:"index;size:64"` // For unauthenticated users
	Name          string        `json:"name"`
	Description   *string       `json:"description"`
	TravelerCount int           `json:"travelerCount"`
	StartDate     string        `json:"startDate"` // ISO yyyy-mm-dd
	EndDate       string        `json:"endDate"`
	Locale        string        `json:"locale" gorm:"size:5"`
	Visibility    string        `json:"visibility" gorm:"size:16"` // private|public|unlisted
	Status        string        `json:"status" gorm:"size:16"`     // draft|active|completed
	Timezone      *string       `json:"timezone"`
	CoverImage    *string       `json:"coverImage"`
	CreatedAt     time.Time     `json:"createdAt"`
	UpdatedAt     time.Time     `json:"updatedAt"`
	Destinations  []Destination `json:"destinations" gorm:"constraint:OnDelete:CASCADE;"`
}
