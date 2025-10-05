package models

// Destination represents a location within a trip
type Destination struct {
	ID                string    `json:"id" gorm:"primaryKey;size:64"`
	TripID            string    `json:"-" gorm:"index"`
	City              string    `json:"city"`
	Region            *string   `json:"region"`
	HeroImage         *string   `json:"heroImage"`
	StartDate         string    `json:"startDate"`
	EndDate           string    `json:"endDate"`
	PlaceholderLocked *bool     `json:"placeholderLocked"`
	DailyPlans        []DayPlan `json:"dailyPlans" gorm:"constraint:OnDelete:CASCADE;"`
}
