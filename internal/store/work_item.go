package store

import (
	"time"

	"github.com/google/uuid"

	"github.com/liftedinit/mfx-migrator/internal/utils"
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

// Equal returns true if the WorkItem is equal to the other WorkItem
func (wi WorkItem) Equal(other WorkItem) bool {
	return wi.Status == other.Status &&
		utils.EqualTimePtr(wi.CreatedDate, other.CreatedDate) &&
		wi.UUID == other.UUID &&
		wi.ManyHash == other.ManyHash &&
		wi.ManifestAddress == other.ManifestAddress &&
		utils.EqualStringPtr(wi.ManifestHash, other.ManifestHash) &&
		utils.EqualTimePtr(wi.ManifestDatetime, other.ManifestDatetime) &&
		utils.EqualStringPtr(wi.Error, other.Error)
}
