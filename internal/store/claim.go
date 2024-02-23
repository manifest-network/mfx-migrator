package store

import (
	"fmt"
	"log/slog"
	"net/url"

	"github.com/google/uuid"
)

// ClaimWorkItemFromQueue retrieves a work item from the remote database work queue.
func (s *Store) ClaimWorkItemFromQueue() (*WorkItem, error) {
	// 0. Create the URL
	fullUrl, err := url.JoinPath(s.rootUrl.String(), GetMigrationsEndpoint())
	if err != nil {
		slog.Error(ErrorGeneratingURL, "error", err)
		return nil, err
	}

	// 1. Get all work items from remote
	response, err := s.client.Get(fullUrl, &WorkItems{})
	if err != nil {
		slog.Error(ErrorGettingWorkItems, "error", err)
		return nil, err
	}

	items := response.Result().(*WorkItems)
	slog.Debug("available work items", "items", items)

	// 2. Loop over all work items
	for _, item := range items.Items {
		// 2.1 Try to claim the work item
		claimedItem, err := s.tryToClaimWorkItem(&item, false)
		if err != nil {
			slog.Error(ErrorClaimingWorkItem, "error", err)
			return nil, err
		}
		if err != nil || claimedItem == nil {
			continue
		}
		return claimedItem, nil
	}

	slog.Debug("no work items available")
	return nil, nil
}

func (s *Store) ClaimWorkItemFromUUID(uuid uuid.UUID, force bool) (*WorkItem, error) {
	// 0. Create the URL
	fullUrl, err := url.JoinPath(s.rootUrl.String(), GetMigrationEndpoint(uuid.String()))
	if err != nil {
		slog.Error(ErrorGeneratingURL, "error", err)
		return nil, err
	}

	// 1. Get the work item from the remote database
	response, err := s.client.Get(fullUrl, &WorkItem{})
	if err != nil {
		slog.Error(ErrorGettingWorkItem, "error", err)
		return nil, err
	}

	item := response.Result().(*WorkItem)
	slog.Debug("work item", "item", item)

	// 2. Try to claim the work item
	claimedItem, err := s.tryToClaimWorkItem(item, force)
	if err != nil {
		slog.Error(ErrorClaimingWorkItem, "error", err)
		return nil, err
	}
	if claimedItem != nil {
		return claimedItem, nil
	}

	return nil, fmt.Errorf("unable to claim the work item: %s", uuid)
}

func (s *Store) tryToClaimWorkItem(item *WorkItem, force bool) (*WorkItem, error) {
	// 1. Try to claim the work item
	updateResponse, err := s.updateWorkItem(*item, CLAIMED, force)
	if err != nil {
		slog.Warn("unable to claim the work item", "msg", err)
		return nil, err
	}

	// 2. Check if the work item was claimed
	if updateResponse.Status == CLAIMED {
		slog.Debug("work item claimed", "uuid", item.UUID)
		return item, nil
	}

	slog.Debug("work item not claimed", "uuid", item.UUID)
	return nil, nil
}

// updateWorkItem updates the status of a work item in the remote database.
func (s *Store) updateWorkItem(item WorkItem, status WorkItemStatus, force bool) (*WorkItemUpdateResponse, error) {
	// 0. Check if the work item is in the correct state to be claimed
	if item.Status != CREATED {
		// If the work item is not in the correct state, we can't claim it, unless we're forcing it
		if !force {
			slog.Error("work item not in the correct state to be claimed", "uuid", item.UUID, "status", item.Status)
			return nil, fmt.Errorf("work item not in the correct state to be claimed: %s", item.Status)
		}
		slog.Warn("forcing re-claim of work item", "uuid", item.UUID, "status", item.Status)
	}

	// 1. Update the work item status
	//    We're updating a copy of the work item, so we don't modify the original
	//    TODO: Does this need to be a copy?
	item.Status = status

	// 2. Send the update request
	// 2.1 Create the URL
	fullUrl, err := url.JoinPath(s.rootUrl.String(), GetUpdateEndpoint(item.UUID))
	if err != nil {
		slog.Error(ErrorGeneratingURL, "error", err)
		return nil, err
	}

	// 2.2 Send the request
	response, err := s.client.Put(fullUrl, item, &WorkItemUpdateResponse{})
	if err != nil {
		slog.Error("error claiming work item", "error", err)
		return nil, err
	}

	// 3. Return the response
	updateResponse := response.Result().(*WorkItemUpdateResponse)
	return updateResponse, nil
}
