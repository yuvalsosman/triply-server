package models

import "time"

// Destination represents a location that can be used across multiple trips
type Destination struct {
	ID string `json:"id" gorm:"primaryKey;size:64"`

	// Location
	City    string  `json:"city" gorm:"size:100;not null"`
	Region  *string `json:"region" gorm:"size:100"`
	Country string  `json:"country" gorm:"size:100;not null"`

	// Geographic
	Latitude  *float64 `json:"latitude" gorm:"type:decimal(10,8)"`
	Longitude *float64 `json:"longitude" gorm:"type:decimal(11,8)"`
	Timezone  *string  `json:"timezone" gorm:"size:50"`

	// Media
	HeroImage *string     `json:"heroImage" gorm:"type:text"`
	Images    StringArray `json:"images" gorm:"type:text"` // JSON array of image URLs

	// Metadata
	Description *string `json:"description" gorm:"type:text"`
	PlaceID     *string `json:"placeId" gorm:"size:255"` // Google Maps Place ID

	// Stats (denormalized for performance)
	TripCount       int     `json:"tripCount" gorm:"default:0"`
	PopularityScore float64 `json:"popularityScore" gorm:"type:decimal(5,2);default:0;index:idx_popularity,sort:desc"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// Relations
	TripDestinations    []TripDestination    `json:"-" gorm:"foreignKey:DestinationID"`
	DayPlanDestinations []DayPlanDestination `json:"-" gorm:"foreignKey:DestinationID"`
}

// TableName specifies the table name for Destination
func (Destination) TableName() string {
	return "destinations"
}
