package utils

import (
	"strconv"
	"time"
)

// GenerateID generates a simple ID based on timestamp
func GenerateID(prefix string) string {
	return prefix + "-" + strconv.FormatInt(time.Now().UnixNano(), 36)
}

// EnsureTripIDs ensures all entities in a trip have IDs
func EnsureTripIDs(trip interface{}) {
	// This will be implemented with actual trip entity
	// For now, this is a placeholder
}
