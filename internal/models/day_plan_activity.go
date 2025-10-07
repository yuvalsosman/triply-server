package models

import "time"

// DayPlanActivity represents activities within a specific day with ordering
type DayPlanActivity struct {
	ID         string `json:"id" gorm:"primaryKey;size:64"`
	DayPlanID  string `json:"dayPlanId" gorm:"size:64;not null;index:idx_day_time_order"`
	ActivityID string `json:"activityId" gorm:"size:64;not null"`

	// Ordering
	TimeOfDay      string `json:"timeOfDay" gorm:"size:20;not null;index:idx_day_time_order"`   // start, mid, end (or morning, afternoon, evening)
	OrderWithinTime int    `json:"orderWithinTime" gorm:"not null;index:idx_day_time_order"` // order within the time_of_day section

	// Optional overrides (user can customize imported activities)
	CustomTitle *string    `json:"customTitle" gorm:"size:255"`
	CustomNotes *string    `json:"customNotes" gorm:"type:text"`
	CustomTime  *time.Time `json:"customTime"` // specific time if user wants

	// Status
	Completed bool `json:"completed" gorm:"default:false"`
	Skipped   bool `json:"skipped" gorm:"default:false"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// Relations
	DayPlan  *DayPlan  `json:"-" gorm:"foreignKey:DayPlanID"`
	Activity *Activity `json:"activity,omitempty" gorm:"foreignKey:ActivityID"`
}

// TableName specifies the table name
func (DayPlanActivity) TableName() string {
	return "day_plan_activities"
}
