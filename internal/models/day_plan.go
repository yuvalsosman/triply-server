package models

import "time"

// DayPlan represents a single day in a trip
type DayPlan struct {
	ID     string `json:"id" gorm:"primaryKey;size:64"`
	TripID string `json:"tripId" gorm:"size:64;not null;index"`

	Date      string  `json:"date" gorm:"type:date;not null"`
	DayNumber int     `json:"dayNumber" gorm:"not null"` // 1, 2, 3, etc.
	Notes     *string `json:"notes" gorm:"type:text"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// Relations
	Trip                 *Trip                 `json:"-" gorm:"foreignKey:TripID"`
	DayPlanDestinations  []DayPlanDestination  `json:"destinations,omitempty" gorm:"foreignKey:DayPlanID;constraint:OnDelete:CASCADE"`
	DayPlanActivities    []DayPlanActivity     `json:"activities,omitempty" gorm:"foreignKey:DayPlanID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for DayPlan
func (DayPlan) TableName() string {
	return "day_plans"
}
