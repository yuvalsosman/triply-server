package models

import "time"

// User represents a user in the system
type User struct {
	ID          string    `json:"id" gorm:"primaryKey;size:64"`
	GoogleID    *string   `json:"googleId,omitempty" gorm:"uniqueIndex"`
	Name        string    `json:"name"`
	DisplayName *string   `json:"displayName,omitempty"`
	Email       string    `json:"email" gorm:"uniqueIndex"`
	Locale      string    `json:"locale" gorm:"size:5;default:en"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	Trips []Trip `json:"-" gorm:"foreignKey:UserID"`
}
