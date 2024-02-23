package store

import (
	"time"

	"github.com/google/uuid"
)

type WorkItemUpdateResponse struct {
	Status           WorkItemStatus `json:"status"`
	UUID             uuid.UUID      `json:"uuid"`
	ManifestDatetime *time.Time     `json:"manifestDatetime"`
	ManifestHash     *string        `json:"manifestHash"`
}
