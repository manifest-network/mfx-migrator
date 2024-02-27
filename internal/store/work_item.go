package store

import (
	"time"

	"github.com/google/uuid"
)

type WorkItem struct {
	Status           WorkItemStatus `json:"status"`
	CreatedDate      *time.Time     `json:"createdDate"`
	UUID             uuid.UUID      `json:"uuid"`
	ManyHash         string         `json:"manyHash"`
	ManifestAddress  string         `json:"manifestAddress"`
	ManifestHash     *string        `json:"manifestHash"`
	ManifestDatetime *time.Time     `json:"manifestDatetime"`
	Error            *string        `json:"error"`
}

func (wi WorkItem) Equal(other WorkItem) bool {
	if wi.Status != other.Status {
		return false
	}
	if (wi.CreatedDate == nil) != (other.CreatedDate == nil) || (wi.CreatedDate != nil && !wi.CreatedDate.Equal(*other.CreatedDate)) {
		return false
	}
	if wi.UUID != other.UUID {
		return false
	}
	if wi.ManyHash != other.ManyHash {
		return false
	}
	if wi.ManifestAddress != other.ManifestAddress {
		return false
	}
	if (wi.ManifestHash == nil) != (other.ManifestHash == nil) || (wi.ManifestHash != nil && *wi.ManifestHash != *other.ManifestHash) {
		return false
	}
	if (wi.ManifestDatetime == nil) != (other.ManifestDatetime == nil) || (wi.ManifestDatetime != nil && !wi.ManifestDatetime.Equal(*other.ManifestDatetime)) {
		return false
	}
	if (wi.Error == nil) != (other.Error == nil) || (wi.Error != nil && *wi.Error != *other.Error) {
		return false
	}

	return true
}
