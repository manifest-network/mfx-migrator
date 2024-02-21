package state

import (
	"time"

	"github.com/google/uuid"
)

type WorkItem struct {
	Status           Status     `json:"status"`
	UUID             uuid.UUID  `json:"uuid"`
	ManyHash         string     `json:"manyHash"`
	ManifestHash     *string    `json:"manifestHash"`
	ManifestDatetime *time.Time `json:"manifestDatetime"`
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
