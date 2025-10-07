package models

import "time"

// DayPlanDestination represents destinations visited during a specific day
// Supports multiple destinations per day (e.g., start in Tokyo, end in Kyoto)
type DayPlanDestination struct {
	ID            string `json:"id" gorm:"primaryKey;size:64"`
	DayPlanID     string `json:"dayPlanId" gorm:"size:64;not null;index"`
	DestinationID string `json:"destinationId" gorm:"size:64;not null"`

	// Order (e.g., 0 = start of day, 1 = end of day)
	OrderIndex int `json:"orderIndex" gorm:"not null"`

	// Timing
	PartOfDay *string `json:"partOfDay" gorm:"size:20"` // morning, afternoon, evening, all-day

	CreatedAt time.Time `json:"createdAt"`

	// Relations
	DayPlan     *DayPlan     `json:"-" gorm:"foreignKey:DayPlanID"`
	Destination *Destination `json:"destination,omitempty" gorm:"foreignKey:DestinationID"`
}

// TableName specifies the table name
func (DayPlanDestination) TableName() string {
	return "day_plan_destinations"
}
