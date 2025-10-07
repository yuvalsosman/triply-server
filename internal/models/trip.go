package models

import "time"

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
	Summary       *string     `json:"summary" gorm:"type:text"`
	Highlights    StringArray `json:"highlights" gorm:"type:text"`
	BudgetLevel   *string     `json:"budgetLevel" gorm:"size:20"`      // budget, moderate, premium, luxury
	Pace          *string     `json:"pace" gorm:"size:20"`             // relaxed, balanced, packed
	Tags          StringArray `json:"tags" gorm:"type:text"`           // culture, foodie, adventure
	TravelerTypes StringArray `json:"travelerTypes" gorm:"type:text"`  // solo, couple, family, friends
	Seasons       StringArray `json:"seasons" gorm:"type:text"`        // spring, summer, autumn, winter
	Likes         int         `json:"likes" gorm:"default:0"`
	ViewCount     int         `json:"viewCount" gorm:"default:0"`

	// Estimated Cost
	EstimatedCostAmount   *int    `json:"estimatedCostAmount"`
	EstimatedCostCurrency *string `json:"estimatedCostCurrency" gorm:"size:3"` // JPY, USD, EUR

	// Timestamps
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	PublishedAt *time.Time `json:"publishedAt"`

	// Relations
	User               *User               `json:"-" gorm:"foreignKey:UserID"`
	DayPlans           []DayPlan           `json:"dayPlans,omitempty" gorm:"foreignKey:TripID;constraint:OnDelete:CASCADE"`
	TripDestinations   []TripDestination   `json:"-" gorm:"foreignKey:TripID;constraint:OnDelete:CASCADE"`
}
