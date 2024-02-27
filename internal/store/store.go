package store

import (
	"fmt"
	"log/slog"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
)

// ClaimWorkItemFromQueue retrieves a work item from the remote database work queue.
func ClaimWorkItemFromQueue(r *resty.Client) (*WorkItem, error) {
	// 1. Get all work items from remote
	items, err := GetAllWorkItems(r)
	if err != nil {
		slog.Error(ErrorGettingWorkItems, "error", err)
		return nil, err
	}
	slog.Debug("available work items", "items", items)

	// 2. Loop over all work items
	for _, item := range items.Items {
		// 2.0 Check if the work item is in the correct state to be claimed
		if item.Status != CREATED {
			slog.Debug("work item not in the correct state to be claimed", "uuid", item.UUID, "status", item.Status)

			// If the work item is not in the correct state, we can't claim it. Continue to the next one
			continue
		}

		// 2.1 Try to claim the work item
		claimedItem, err := tryToClaimWorkItem(r, &item)
		if err != nil {
			slog.Error(ErrorClaimingWorkItem, "error", err)
			return nil, err
		}

		// 2.2 Unable to claim the work item, continue to the next one
		if claimedItem == nil {
			continue
		}

		return claimedItem, nil
	}

	slog.Debug("no work items available")
	return nil, nil
}

func ClaimWorkItemFromUUID(r *resty.Client, uuid uuid.UUID, force bool) (*WorkItem, error) {
	// 1. Get the work item from the remote database
	item, err := GetWorkItem(r, uuid)
	if err != nil {
		slog.Error(ErrorGettingWorkItem, "error", err)
		return nil, err
	}
	slog.Debug("work item", "item", item)

	// 2. Check if the work item is in the correct state to be claimed
	if item.Status != CREATED {
		// If the work item is not in the correct state, we can't claim it, unless we're forcing it
		if !force {
			slog.Error("work item not in the correct state to be claimed", "uuid", item.UUID, "status", item.Status)
			return nil, fmt.Errorf("work item not in the correct state to be claimed: %s, %s", item.UUID, item.Status.String())
		}
		slog.Warn("forcing re-claim of work item", "uuid", item.UUID, "status", item.Status)
	}

	// 3. Try to claim the work item
	claimedItem, err := tryToClaimWorkItem(r, item)
	if err != nil {
		slog.Error(ErrorClaimingWorkItem, "error", err)
		return nil, err
	}
	if claimedItem != nil {
		return claimedItem, nil
	}

	return nil, fmt.Errorf("unable to claim the work item: %s", uuid)
}

func tryToClaimWorkItem(r *resty.Client, item *WorkItem) (*WorkItem, error) {
	// 1. Try to claim the work item
	updateResponse, err := UpdateWorkItem(r, *item, CLAIMED)
	if err != nil {
		slog.Warn("unable to claim the work item", "msg", err)
		return nil, err
	}

	// 1.1 Make sure we have a response
	if updateResponse == nil {
		slog.Error("no update response returned when claiming work item")
		return nil, fmt.Errorf("no update response returned when claiming work item: %s", item.UUID)
	}

	// 2. Check if the work item was claimed
	if updateResponse.Status == CLAIMED {
		slog.Debug("work item claimed", "uuid", item.UUID)
		item.Status = CLAIMED
		return item, nil
	}

	slog.Debug("work item not claimed", "uuid", item.UUID)
	return nil, nil
}

// GetWorkItem retrieves a work item from the remote database by UUID.
func GetWorkItem(r *resty.Client, uuid uuid.UUID) (*WorkItem, error) {
	req := r.R().
		SetPathParam("uuid", uuid.String()).
		SetResult(&WorkItem{})
	response, err := req.Get("neighborhoods/{neighborhood}/migrations/{uuid}")
	if err != nil {
		slog.Error(ErrorGettingWorkItem, "error", err)
		return nil, err
	}

	item := response.Result().(*WorkItem)
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
	return items, nil
}

// UpdateWorkItem updates the status of a work item in the remote database.
func UpdateWorkItem(r *resty.Client, item WorkItem, status WorkItemStatus) (*WorkItemUpdateResponse, error) {
	// 1. Create an update request
	updateRequest := WorkItemUpdateRequest{
		Status:           status,
		ManifestDatetime: item.ManifestDatetime,
		ManifestHash:     item.ManifestHash,
		Error:            item.Error,
	}

	// 2. Send the update request
	// 2.1 Send the request
	req := r.R().
		SetPathParam("uuid", item.UUID.String()).
		SetBody(&updateRequest).
		SetResult(&WorkItemUpdateResponse{})
	response, err := req.Put("neighborhoods/{neighborhood}/migrations/{uuid}")
	if err != nil {
		slog.Error("error claiming work item", "error", err)
		return nil, err
	}

	if response == nil {
		slog.Error("no response returned when claiming work item")
		return nil, fmt.Errorf("no response returned when claiming work item: %s", item.UUID)
	}

	slog.Debug("update response",
		"status_code", response.StatusCode(),
		"status", response.Status(),
		"proto", response.Proto(),
		"time", response.Time(),
		"received_at", response.ReceivedAt(),
		"body", response.String())

	// 3. Return the response
	updateResponse := response.Result().(*WorkItemUpdateResponse)
	return updateResponse, nil
}
