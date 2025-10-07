package models

import "time"

// User represents a user in the system
type User struct {
	ID          string    `json:"id" gorm:"primaryKey;size:64"`
	GoogleID    *string   `json:"googleId,omitempty" gorm:"size:100;uniqueIndex"`
	Name        string    `json:"name" gorm:"size:255;not null"`
	Email       string    `json:"email" gorm:"size:255;uniqueIndex;not null"`
	DisplayName string    `json:"displayName" gorm:"size:100"`
	AvatarURL   *string   `json:"avatarUrl" gorm:"type:text"`
	Locale      string    `json:"locale" gorm:"size:10;default:'en'"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
