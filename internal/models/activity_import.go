package models

import "time"

// ActivityImport tracks when activities are imported from one trip to another
// Useful for analytics and recommendations
type ActivityImport struct {
	ID string `json:"id" gorm:"primaryKey;size:64"`

	SourceTripID     *string `json:"sourceTripId" gorm:"size:64;index"` // can be null if activity is from library
	TargetTripID     string  `json:"targetTripId" gorm:"size:64;not null;index"`
	ActivityID       string  `json:"activityId" gorm:"size:64;not null;index"`
	ImportedByUserID string  `json:"importedByUserId" gorm:"size:64;not null"`

	ImportedAt time.Time `json:"importedAt"`

	// Relations
	SourceTrip     *Trip     `json:"-" gorm:"foreignKey:SourceTripID"`
	TargetTrip     *Trip     `json:"-" gorm:"foreignKey:TargetTripID"`
	Activity       *Activity `json:"-" gorm:"foreignKey:ActivityID"`
	ImportedByUser *User     `json:"-" gorm:"foreignKey:ImportedByUserID"`
}

// TableName specifies the table name
func (ActivityImport) TableName() string {
	return "activity_imports"
}
