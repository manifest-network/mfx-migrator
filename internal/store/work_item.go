package store

import (
	"time"

	"github.com/google/uuid"
)

type WorkItem struct {
	Status           WorkItemStatus `json:"status"`
	UUID             uuid.UUID      `json:"uuid"`
	ManyHash         string         `json:"manyHash"`
	ManifestHash     *string        `json:"manifestHash"`
	ManifestDatetime *time.Time     `json:"manifestDatetime"`
}
