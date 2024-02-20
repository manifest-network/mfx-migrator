package state

import (
	"time"

	"github.com/google/uuid"
)

type RemoteStatus int

const (
	CREATED RemoteStatus = iota + 1
	FAILED
)

type WorkItem struct {
	Status           RemoteStatus `json:"status"`
	UUID             uuid.UUID    `json:"uuid"`
	ManyHash         string       `json:"manyHash"`
	ManifestHash     *string      `json:"manifestHash"`
	ManifestDatetime *time.Time   `json:"manifestDatetime"`
}

type Meta struct {
	TotalItems   int `json:"totalItems"`
	ItemCount    int `json:"itemCount"`
	ItemsPerPage int `json:"itemsPerPage"`
	TotalPages   int `json:"totalPages"`
	CurrentPage  int `json:"currentPage"`
}

type WorkItems struct {
	Items []WorkItem `json:"items"`
	Meta  Meta       `json:"meta"`
}
