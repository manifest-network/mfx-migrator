package utils

import "github.com/google/uuid"

// IsValidUUID checks if a string is a valid UUID.
func IsValidUUID(uuidStr string) bool {
	_, err := uuid.Parse(uuidStr)
	return err == nil
}
