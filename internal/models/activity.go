package models

import "time"

// Activity represents a reusable activity that can be used in multiple trips
type Activity struct {
	ID string `json:"id" gorm:"primaryKey;size:64"`

	// Basic Info
	Title       string  `json:"title" gorm:"size:255;not null"`
	Description *string `json:"description" gorm:"type:text"`
	Type        string  `json:"type" gorm:"size:50;not null;index"` // meal, culture, transportation, accommodation, etc.

	// Location
	Location  *string  `json:"location" gorm:"size:255"`
	Address   *string  `json:"address" gorm:"type:text"`
	Latitude  *float64 `json:"latitude" gorm:"type:decimal(10,8)"`
	Longitude *float64 `json:"longitude" gorm:"type:decimal(11,8)"`
	PlaceID   *string  `json:"placeId" gorm:"size:255"` // Google Maps Place ID

	// Details
	DurationMinutes       *int    `json:"durationMinutes"`
	EstimatedCostAmount   *int    `json:"estimatedCostAmount"`
	EstimatedCostCurrency *string `json:"estimatedCostCurrency" gorm:"size:3"`

	// Media
	ImageURL *string     `json:"imageUrl" gorm:"type:text"`
	Images   StringArray `json:"images" gorm:"type:text"` // JSON array

	// Metadata
	Tags StringArray `json:"tags" gorm:"type:text"` // photography, family-friendly, must-see
	URL  *string     `json:"url" gorm:"type:text"`  // external link

	// Stats (for recommendations)
	UsageCount    int     `json:"usageCount" gorm:"default:0;index:idx_usage,sort:desc"`
	AverageRating float64 `json:"averageRating" gorm:"type:decimal(3,2);default:0"`

	// Created by (optional - for user-generated activities)
	CreatedByUserID *string `json:"createdByUserId" gorm:"size:64"`
	IsVerified      bool    `json:"isVerified" gorm:"default:false"` // curated activities

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// Relations
	CreatedByUser     *User               `json:"-" gorm:"foreignKey:CreatedByUserID"`
	DayPlanActivities []DayPlanActivity   `json:"-" gorm:"foreignKey:ActivityID"`
	ActivityImports   []ActivityImport    `json:"-" gorm:"foreignKey:ActivityID"`
}

// TableName specifies the table name for Activity
func (Activity) TableName() string {
	return "activities"
}
