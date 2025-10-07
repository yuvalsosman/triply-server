package models

import "time"

// TripDestination represents the many-to-many relationship between trips and destinations
type TripDestination struct {
	ID            string `json:"id" gorm:"primaryKey;size:64"`
	TripID        string `json:"tripId" gorm:"size:64;not null;index:idx_trip_dest,unique"`
	DestinationID string `json:"destinationId" gorm:"size:64;not null;index:idx_trip_dest,unique"`

	// Order in trip itinerary
	OrderIndex int `json:"orderIndex" gorm:"not null;index:idx_order"`

	// Date range in this destination
	StartDate *string `json:"startDate" gorm:"type:date"`
	EndDate   *string `json:"endDate" gorm:"type:date"`

	// Optional overrides
	CustomNotes     *string `json:"customNotes" gorm:"type:text"`
	CustomHeroImage *string `json:"customHeroImage" gorm:"type:text"`

	CreatedAt time.Time `json:"createdAt"`

	// Relations
	Trip        *Trip        `json:"-" gorm:"foreignKey:TripID"`
	Destination *Destination `json:"destination,omitempty" gorm:"foreignKey:DestinationID"`
}

// TableName specifies the table name
func (TripDestination) TableName() string {
	return "trip_destinations"
}
