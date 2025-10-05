package models

import "gorm.io/gorm"

// Coordinates represents geographic coordinates
type Coordinates struct {
	Lat float64 `json:"lat" gorm:"column:lat"`
	Lng float64 `json:"lng" gorm:"column:lng"`
}

// Activity represents a single activity within a day plan
type Activity struct {
	ID        string   `json:"id" gorm:"primaryKey;size:64"`
	DayPlanID string   `json:"-" gorm:"index"`
	Title     string   `json:"title"`
	TimeOfDay string   `json:"timeOfDay"` // start|mid|end
	Order     int      `json:"order"`
	Location  string   `json:"location"`
	Address   *string  `json:"address"`
	Type      string   `json:"type"` // transportation|culture|accommodation|meal|experience
	Cost      *string  `json:"cost"`
	PlaceID   *string  `json:"placeId"`
	Lat       *float64 `json:"-"`
	Lng       *float64 `json:"-"`

	// Transient JSON field to map coordinates to Lat/Lng
	Coordinates *Coordinates `json:"coordinates" gorm:"-:all"`
}

// BeforeSave hook to convert coordinates to lat/lng columns
func (a *Activity) BeforeSave(tx *gorm.DB) (err error) {
	if a.Coordinates != nil {
		a.Lat = &a.Coordinates.Lat
		a.Lng = &a.Coordinates.Lng
	} else {
		a.Lat = nil
		a.Lng = nil
	}
	return nil
}

// AfterFind hook to populate coordinates from lat/lng columns
func (a *Activity) AfterFind(tx *gorm.DB) (err error) {
	if a.Lat != nil && a.Lng != nil {
		a.Coordinates = &Coordinates{Lat: *a.Lat, Lng: *a.Lng}
	}
	return nil
}
