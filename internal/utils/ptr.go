package utils

import "time"

// EqualStringPtr returns true if the string pointers are equal
func EqualStringPtr(a, b *string) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

// EqualTimePtr returns true if the time pointers are equal
func EqualTimePtr(a, b *time.Time) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.Equal(*b)
}
