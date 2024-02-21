package store

import (
	"log/slog"

	"github.com/google/uuid"
	"github.com/liftedinit/mfx-migrator/internal/state"
)

// claimWorkItems tries to claim work items from the remote database.
func claimWorkItems(updateURL string, items *state.WorkItems, force bool) (bool, error) {
	for _, item := range items.Items {
		claimed, err := claimWorkItem(updateURL, &item, force)
		if err != nil {
			slog.Error("error claiming work item", "error", err)
			return false, err
		}

		if claimed {
			return true, nil
		}
	}

	// If we're here, no work items were available
	slog.Info("no work items available")
	return false, nil
}

// claimWorkItem attempts to claim a work item from the remote database.
// Returns true if the work item was claimed, false if it was not available or if an error happened.
// If the work item was claimed, the local state is updated.
func claimWorkItem(updateURL string, item *state.WorkItem, force bool) (bool, error) {
	slog.Debug("claimWorkItem", "updateURL", updateURL, "item", item)

	// Check if the work item is in the correct state to be claimed
	ok, err := checkWorkItemStatus(item, force)
	if err != nil {
		return false, err
	}

	// If the work item is not in the correct state, we can't claim it
	if !ok {
		return false, nil
	}

	return updateWorkItemStatus(updateURL, item, state.CLAIMED)
}

// ClaimWorkItemFromQueue retrieves a work item from the remote database work queue.
// Returns true if a work item was claimed, false if no work items are available or if an error happened.
func (s *RemoteStore) ClaimWorkItemFromQueue(rootURL string) (bool, error) {
	allWorkURL := s.Router.GetAllWorkURL(rootURL)
	slog.Debug("ClaimWorkItemFromQueue", "rootURL", rootURL, "allWorkURL", allWorkURL)

	// Get all work items
	items, err := getAllWorkItems(allWorkURL)
	if err != nil {
		slog.Error("error getting all work items", "error", err)
		return false, err
	}

	updateURL := s.Router.UpdateWorkURL(rootURL)
	return claimWorkItems(updateURL, items, false)
}

// ClaimWorkItemFromUUID retrieves a work item from the remote database by UUID.
// Returns true if a work item was claimed, false if no work items are available or if an error happened.
func (s *RemoteStore) ClaimWorkItemFromUUID(rootURL string, uuidStr uuid.UUID, force bool) (bool, error) {
	workURL := s.Router.GetWorkURL(rootURL)
	slog.Debug("ClaimWorkItemFromUUID", "rootURL", rootURL, "workURL", workURL, "uuid", uuidStr)

	// Get the work item from the remote database
	item, err := getWorkItem(workURL, uuidStr)
	if err != nil {
		slog.Error("error getting work item", "error", err)
		return false, err
	}

	// And try to claim it
	items := &state.WorkItems{Items: []state.WorkItem{*item}}
	updateURL := s.Router.UpdateWorkURL(rootURL)
	return claimWorkItems(updateURL, items, force)
}
