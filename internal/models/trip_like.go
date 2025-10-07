package models

import "time"

// TripLike represents a user's like on a public trip
type TripLike struct {
	ID        string    `json:"id" gorm:"primaryKey;size:64"`
	UserID    string    `json:"userId" gorm:"size:64;not null;index:idx_trip_likes_user"`
	TripID    string    `json:"tripId" gorm:"size:64;not null;index:idx_trip_likes_trip"`
	CreatedAt time.Time `json:"createdAt" gorm:"index:idx_trip_likes_created,sort:desc"`

	// Relations
	User *User `json:"-" gorm:"foreignKey:UserID"`
	Trip *Trip `json:"-" gorm:"foreignKey:TripID"`
}

// TableName specifies the table name for TripLike
func (TripLike) TableName() string {
	return "trip_likes"
}
