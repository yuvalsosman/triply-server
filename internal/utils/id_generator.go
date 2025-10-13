package utils

import (
	"crypto/rand"
	"encoding/base32"
	"strconv"
	"strings"
	"time"
)

// GenerateID generates a unique ID with timestamp and random component
func GenerateID(prefix string) string {
	// Timestamp component (base36)
	timestamp := strconv.FormatInt(time.Now().UnixNano(), 36)

	// Random component (4 bytes = 6-7 chars in base32)
	randomBytes := make([]byte, 4)
	rand.Read(randomBytes)
	randomStr := strings.ToLower(strings.TrimRight(base32.StdEncoding.EncodeToString(randomBytes), "="))

	return prefix + "-" + timestamp + randomStr[:6]
}

// EnsureTripIDs ensures all entities in a trip have IDs
func EnsureTripIDs(trip interface{}) {
	// This will be implemented with actual trip entity
	// For now, this is a placeholder
}
