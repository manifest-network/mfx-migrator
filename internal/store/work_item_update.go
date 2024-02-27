package store

import (
	"time"
)

type WorkItemUpdateRequest struct {
	Status           WorkItemStatus `json:"status"`
	ManifestDatetime *time.Time     `json:"manifestDatetime"`
	ManifestHash     *string        `json:"manifestHash"`
	Error            *string        `json:"error"`
}

type WorkItemUpdateResponse struct {
	Status           WorkItemStatus `json:"status"`
	ManifestDatetime *time.Time     `json:"manifestDatetime"`
	ManifestHash     *string        `json:"manifestHash"`
	Error            *string        `json:"error"`
}
