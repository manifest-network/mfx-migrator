package store

import (
	"time"

	"github.com/google/uuid"

	"github.com/liftedinit/mfx-migrator/internal/utils"
)

type Token struct {
	AccessToken string `json:"access_token"`
}

type WorkItemStatus int

const (
	CREATED WorkItemStatus = iota + 1
	CLAIMED
	MIGRATING
	COMPLETED
	FAILED
)

func (s WorkItemStatus) String() string {
	return [...]string{"created", "claimed", "migrating", "completed", "failed"}[s-1]
}

// EnumIndex returns the enum index of a LocalWorkItemStatus.
func (s WorkItemStatus) EnumIndex() int64 {
	return int64(s)
}

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

// IsNil returns true if the WorkItem is nil
func (wi WorkItem) IsNil() bool {
	return wi.UUID == uuid.Nil
}

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

// IsNil returns true if the WorkItems is nil
func (wi WorkItems) IsNil() bool {
	return len(wi.Items) == 0 && wi.Meta.TotalPages == 0 && wi.Meta.CurrentPage == 0 && wi.Meta.ItemsPerPage == 0 && wi.Meta.ItemCount == 0 && wi.Meta.TotalItems == 0
}

type ContextKey string
