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

	// 2. Loop over all work items
	for _, item := range items.Items {
		// 2.0 Check if the work item is in the correct state to be claimed
		if !itemCanBeClaimed(&item, false) {
			slog.Warn("unable to claim work item, invalid state", "uuid", item.UUID, "status", item.Status.String())
			continue
		}

		// 2.1 Try claiming the work item
		return claimItem(r, &item)
	}

	// No work items available
	return nil, nil
}

func ClaimWorkItemFromUUID(r *resty.Client, uuid uuid.UUID, force bool) (*WorkItem, error) {
	// 1. Get the work item from the remote database
	item, err := GetWorkItem(r, uuid)
	if err != nil {
		slog.Error(ErrorGettingWorkItem, "error", err)
		return nil, err
	}

	// 2. Check if the work item is in the correct state to be claimed
	if !itemCanBeClaimed(item, force) {
		return nil, fmt.Errorf("unable to claim work item, invalid state: %s", &item.Status)
	}

	// 3. Try to claim the work item
	return claimItem(r, item)
}

func claimItem(r *resty.Client, item *WorkItem) (*WorkItem, error) {
	// 1. Try to claim the work item
	newItem := *item
	newItem.Status = CLAIMED
	if err := UpdateWorkItemAndSaveState(r, newItem); err != nil {
		slog.Error(ErrorClaimingWorkItem, "error", err)
		return nil, err
	}

	return &newItem, nil
}

func itemCanBeClaimed(item *WorkItem, force bool) bool {
	if item.Status != CREATED {
		if force {
			slog.Warn("forcing re-claim of work item", "uuid", item.UUID, "status", item.Status)
			return true
		}
		slog.Debug("work item not in the correct state to be claimed", "uuid", item.UUID, "status", item.Status)
		return false
	}
	return true
}
