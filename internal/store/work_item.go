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
}
