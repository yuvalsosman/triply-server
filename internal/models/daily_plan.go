package models

// DayPlan represents activities planned for a specific day
type DayPlan struct {
	ID            string     `json:"id" gorm:"primaryKey;size:64"`
	DestinationID string     `json:"-" gorm:"index"`
	Date          string     `json:"date"`
	Notes         *string    `json:"notes"`
	Activities    []Activity `json:"activities" gorm:"constraint:OnDelete:CASCADE;"`
}
