package store

import (
	"fmt"
	"log/slog"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

// ClaimWorkItemFromQueue retrieves a work item from the remote database work queue.
// Items will be claimed in parallel using goroutine.
// The maximum number of items that can be claimed in parallel is defined by the pagination of the remote database.
func ClaimWorkItemFromQueue(r *resty.Client) ([]*WorkItem, error) {
	// 1. Get all work items from remote
	status := CREATED
	items, err := GetAllWorkItems(r, &status)
	if err != nil {
		return nil, errors.WithMessage(err, ErrorGettingWorkItems)
	}

	var g errgroup.Group
	claimedItems := make([]*WorkItem, 0)

	// 2. Loop over all work items
	for _, item := range items.Items {
		item := item
		g.Go(func() error {
			// 2.0 Check if the work item is in the correct state to be claimed
			if !itemCanBeClaimed(&item, false) {
				slog.Warn("unable to claim work item, invalid state", "uuid", item.UUID, "status", item.Status.String())
				return nil
			}

			claimedItem, err := claimItem(r, &item)
			if err != nil {
				return errors.WithMessage(err, ErrorClaimingWorkItem)
			}
			claimedItems = append(claimedItems, claimedItem)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, errors.WithMessage(err, ErrorClaimingWorkItem)
	}

	// No work items available
	return claimedItems, nil
}

func ClaimWorkItemFromUUID(r *resty.Client, uuid uuid.UUID, force bool) (*WorkItem, error) {
	// 1. Get the work item from the remote database
	item, err := GetWorkItem(r, uuid)
	if err != nil {
		return nil, errors.WithMessage(err, ErrorGettingWorkItem)
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
	if item.Error != nil {
		slog.Info("clearing previous error", "uuid", item.UUID, "error", item.Error)
		newItem.Error = nil
	}
	if err := UpdateWorkItemAndSaveState(r, newItem); err != nil {
		return nil, errors.WithMessage(err, ErrorClaimingWorkItem)
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
