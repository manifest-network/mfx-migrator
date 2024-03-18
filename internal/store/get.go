package store

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

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
		return nil, errors.WithMessage(err, ErrorGettingWorkItem)
	}

	if response == nil {
		return nil, fmt.Errorf("response is nil")
	}

	statusCode := response.StatusCode()
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("response status code: %d", statusCode)
	}

	item := response.Result().(*WorkItem)
	if item == nil || (item != nil && item.IsNil()) {
		return nil, fmt.Errorf("error unmarshalling work item")
	}
	slog.Debug("work item", "item", item)
	return item, nil
}

// GetAllWorkItems retrieves all work items from the remote database.
func GetAllWorkItems(r *resty.Client, status *WorkItemStatus) (*WorkItems, error) {
	req := r.R().SetResult(&WorkItems{})
	if status != nil {
		req.SetQueryParam("status", strconv.FormatInt(status.EnumIndex(), 10))
	}
	response, err := req.Get("neighborhoods/{neighborhood}/migrations/")
	if err != nil {
		return nil, errors.WithMessage(err, ErrorGettingWorkItems)
	}

	if response == nil {
		return nil, fmt.Errorf("response is nil")
	}

	statusCode := response.StatusCode()
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("response status code: %d", statusCode)
	}

	items := response.Result().(*WorkItems)
	if items == nil || (items != nil && items.IsNil()) {
		return nil, fmt.Errorf("error unmarshalling work items")
	}
	slog.Debug("work items", "items", items)
	return items, nil
}
