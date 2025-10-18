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
	Name string  `json:"name" gorm:"size:255;not null"`
	Slug *string `json:"slug" gorm:"size:255;uniqueIndex"`

	// Trip Details
	TravelerCount int    `json:"travelerCount" gorm:"default:1"` // Total count (computed from Adults + len(ChildrenAges))
	Adults        int    `json:"adults" gorm:"default:2"`
	ChildrenAges  string `json:"childrenAges" gorm:"type:text"` // JSON array of child ages, e.g., "[5,8,12]"
	StartDate     string `json:"startDate" gorm:"type:date;not null"`
	EndDate       string `json:"endDate" gorm:"type:date;not null"`

	// Media
	CoverImage string `json:"coverImage" gorm:"type:text;not null"`

	// Visibility & Status
	Visibility string `json:"visibility" gorm:"size:20;default:'public';index"` // private, public
	Status     string `json:"status" gorm:"size:20;default:'active'"`           // active, completed

	// Public Trip Metadata (only for public trips)
	Summary      *string `json:"summary" gorm:"type:text"`
	TravelerType string  `json:"travelerType" gorm:"size:50;not null;default:''"` // סולו, זוג, משפחה, חברים (single value)
	Likes        int     `json:"likes" gorm:"default:0"`
	CloneCount   int     `json:"cloneCount" gorm:"default:0"` // Number of times this trip has been cloned

	// Timestamps
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// Relations
	User             *User             `json:"-" gorm:"foreignKey:UserID"`
	DayPlans         []DayPlan         `json:"dayPlans,omitempty" gorm:"foreignKey:TripID;constraint:OnDelete:CASCADE"`
	TripDestinations []TripDestination `json:"-" gorm:"foreignKey:TripID;constraint:OnDelete:CASCADE"`
}
