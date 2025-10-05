package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// StringArray is a custom type for handling string arrays in PostgreSQL
type StringArray []string

// Scan implements the sql.Scanner interface
func (s *StringArray) Scan(value interface{}) error {
	if value == nil {
		*s = []string{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, s)
	case string:
		return json.Unmarshal([]byte(v), s)
	default:
		return errors.New("unsupported type for StringArray")
	}
}

// Value implements the driver.Valuer interface
func (s StringArray) Value() (driver.Value, error) {
	if len(s) == 0 {
		return "[]", nil
	}
	return json.Marshal(s)
}

// PublicTrip represents a publicly shared trip with additional metadata
type PublicTrip struct {
	ID                    string      `json:"id" gorm:"primaryKey;size:64"`
	TripID                string      `json:"tripId" gorm:"index"`
	Slug                  string      `json:"slug" gorm:"uniqueIndex"`
	HeroImageURL          string      `json:"heroImageUrl"`
	Summary               *string     `json:"summary"`
	Highlights            StringArray `json:"highlights" gorm:"type:text"`
	OriginCities          StringArray `json:"originCities" gorm:"type:text"`
	DurationDays          int         `json:"durationDays"`
	StartMonth            int         `json:"startMonth"`
	Seasons               StringArray `json:"seasons" gorm:"type:text"`
	BudgetLevel           string      `json:"budgetLevel"`
	Pace                  string      `json:"pace"`
	Tags                  StringArray `json:"tags" gorm:"type:text"`
	TravelerTypes         StringArray `json:"travelerTypes" gorm:"type:text"`
	Likes                 int         `json:"likes" gorm:"default:0"`
	AuthorName            string      `json:"-"`
	AuthorAvatarURL       *string     `json:"-"`
	AuthorHomeCity        *string     `json:"-"`
	EstimatedCostCurrency *string     `json:"-"`
	EstimatedCostAmount   *int        `json:"-"`
	CreatedAt             time.Time   `json:"createdAt"`
	UpdatedAt             time.Time   `json:"updatedAt"`

	Trip *Trip `json:"-" gorm:"foreignKey:TripID"`
}

// PublicTripAuthor represents the author information for a public trip
type PublicTripAuthor struct {
	Name      string  `json:"name"`
	AvatarURL *string `json:"avatarUrl,omitempty"`
	HomeCity  *string `json:"homeCity,omitempty"`
}

// PublicTripCost represents the estimated cost of a trip
type PublicTripCost struct {
	Currency string `json:"currency"`
	Amount   int    `json:"amount"`
}

// PublicTripMetadata represents metadata for a public trip
type PublicTripMetadata struct {
	Visibility string    `json:"visibility"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}
