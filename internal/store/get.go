package store

import (
	"log/slog"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// GetWorkItem retrieves a work item from the remote database by UUID.
func GetWorkItem(r *resty.Client, itemUUID uuid.UUID) (*WorkItem, error) {
	req := r.R().
		SetPathParam("uuid", itemUUID.String()).
		SetResult(&WorkItem{})
	response, err := req.Get("neighborhoods/{neighborhood}/migrations/{uuid}")
	if err != nil {
		slog.Error(ErrorGettingWorkItem, "error", err)
		return nil, err
	}

	item := response.Result().(*WorkItem)
	if item == nil || (item != nil && item.IsNil()) {
		slog.Error("error unmarshalling work item")
		return nil, errors.New("error unmarshalling work item")
	}
	slog.Debug("work item", "item", item)
	return item, nil
}

// GetAllWorkItems retrieves all work items from the remote database.
func GetAllWorkItems(r *resty.Client) (*WorkItems, error) {
	req := r.R().SetResult(&WorkItems{})
	response, err := req.Get("neighborhoods/{neighborhood}/migrations/")
	if err != nil {
		slog.Error(ErrorGettingWorkItems, "error", err)
		return nil, err
	}

	items := response.Result().(*WorkItems)
	if items == nil || (items != nil && items.IsNil()) {
		slog.Error("error unmarshalling work items")
		return nil, errors.New("error unmarshalling work items")
	}
	slog.Debug("work items", "items", items)
	return items, nil
}
