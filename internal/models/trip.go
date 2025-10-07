package models

import "time"

// TravelerType represents the type of travelers
type TravelerType string

const (
	TravelerTypeSolo    TravelerType = "סולו"
	TravelerTypeCouple  TravelerType = "זוג"
	TravelerTypeFamily  TravelerType = "משפחה"
	TravelerTypeFriends TravelerType = "חברים"
)

// Trip represents a trip (both private and public)
type Trip struct {
	ID     string `json:"id" gorm:"primaryKey;size:64"`
	UserID string `json:"userId" gorm:"size:64;not null;index"`

	// Basic Info
	Name        string  `json:"name" gorm:"size:255;not null"`
	Description *string `json:"description" gorm:"type:text"`
	Slug        *string `json:"slug" gorm:"size:255;uniqueIndex"`

	// Trip Details
	TravelerCount int     `json:"travelerCount" gorm:"default:1"`
	StartDate     string  `json:"startDate" gorm:"type:date;not null"`
	EndDate       string  `json:"endDate" gorm:"type:date;not null"`
	Timezone      *string `json:"timezone" gorm:"size:50"`

	// Media
	CoverImage *string `json:"coverImage" gorm:"type:text"`
	HeroImage  *string `json:"heroImage" gorm:"type:text"`

	// Visibility & Status
	Visibility string `json:"visibility" gorm:"size:20;default:'private';index"` // private, unlisted, public
	Status     string `json:"status" gorm:"size:20;default:'planning'"`          // planning, active, completed, archived

	// Public Trip Metadata (only for public trips)
	Summary      *string     `json:"summary" gorm:"type:text"`
	Tags         StringArray `json:"tags" gorm:"type:text"`                           // culture, foodie, adventure
	TravelerType string      `json:"travelerType" gorm:"size:50;not null;default:''"` // סולו, זוג, משפחה, חברים (single value)
	Likes        int         `json:"likes" gorm:"default:0"`

	// Timestamps
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	PublishedAt *time.Time `json:"publishedAt"`

	// Relations
	User             *User             `json:"-" gorm:"foreignKey:UserID"`
	DayPlans         []DayPlan         `json:"dayPlans,omitempty" gorm:"foreignKey:TripID;constraint:OnDelete:CASCADE"`
	TripDestinations []TripDestination `json:"-" gorm:"foreignKey:TripID;constraint:OnDelete:CASCADE"`
}
