package store

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/google/uuid"

	"github.com/liftedinit/mfx-migrator/internal/state"
	"github.com/liftedinit/mfx-migrator/internal/utils"
)

// GetAllWorkItems retrieves all work items from the remote database.
func GetAllWorkItems(requestURL string) (*state.WorkItems, error) {
	slog.Debug("GetAllWorkItems", "requestURL", requestURL)
	body, err := utils.MakeHTTPRequest(requestURL)
	if err != nil {
		return nil, err
	}

	var workItems state.WorkItems
	err = json.Unmarshal(body, &workItems)
	if err != nil {
		slog.Error("error unmarshalling json", "error", err)
		return nil, err
	}

	slog.Debug("GetAllWorkItems", "allWorkItems", workItems)

	return &workItems, nil
}

// GetWorkItem retrieves a work item from the remote database by UUID.
func GetWorkItem(requestURL string, uuidStr uuid.UUID) (*state.WorkItem, error) {
	slog.Debug("GetWorkItem", "requestURL", requestURL, "uuid", uuidStr)
	body, err := utils.MakeHTTPRequest(requestURL)
	if err != nil {
		return nil, err
	}

	var workItem state.WorkItem
	err = json.Unmarshal(body, &workItem)
	if err != nil {
		slog.Error("error unmarshalling json", "error", err)
		return nil, err
	}

	if workItem.UUID != uuidStr {
		slog.Error("uuid mismatch", "uuid", uuidStr, "workItemUUID", workItem.UUID)
		return nil, fmt.Errorf("uuid mismatch")

	}

	slog.Debug("GetWorkItem", "workItem", workItem)

	return &workItem, nil
}

// ClaimWorkItemFromQueue retrieves a work item from the remote database work queue.
// Returns true if a work item was claimed, false if no work items are available or if an error happened.
func ClaimWorkItemFromQueue(requestURL string) (bool, error) {
	slog.Debug("ClaimWorkItemFromQueue", "requestURL", requestURL)

	// Get all work items
	items, err := GetAllWorkItems(requestURL)
	if err != nil {
		slog.Error("error getting all work items", "error", err)
		return false, err
	}

	return claimWorkItems(requestURL, items, false)
}

// ClaimWorkItemFromUUID retrieves a work item from the remote database by UUID.
// Returns true if a work item was claimed, false if no work items are available or if an error happened.
func ClaimWorkItemFromUUID(requestURL string, uuidStr uuid.UUID, force bool) (bool, error) {
	slog.Debug("ClaimWorkItemFromUUID", "requestURL", requestURL, "uuid", uuidStr)

	// Get the work item from the remote database
	item, err := GetWorkItem(requestURL, uuidStr)
	if err != nil {
		slog.Error("error getting work item", "error", err)
		return false, err
	}

	// And try to claim it
	items := &state.WorkItems{Items: []state.WorkItem{*item}}
	return claimWorkItems(requestURL, items, force)
}

// claimWorkItems tries to claim work items from the remote database.
func claimWorkItems(requestURL string, items *state.WorkItems, force bool) (bool, error) {
	for _, item := range items.Items {
		claimed, err := claimWorkItem(requestURL, &item, force)
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

// checkWorkItemStatus checks if a work item is in the correct state to be claimed.
// If force is true, the check is skipped.
func checkWorkItemStatus(item *state.WorkItem, force bool) (bool, error) {
	if item == nil {
		slog.Error("work item is nil")
		return false, fmt.Errorf("work item is nil")
	}

	if item.Status != state.CREATED {
		if force {
			slog.Info("forcing claim of work item", "status", item.Status)
		} else {
			slog.Error("work item is not in the correct state", "status", item.Status)
			return false, fmt.Errorf("work item is not in the correct state")
		}
	}

	return true, nil
}

// claimWorkItem does the actual work of claiming a work item from the remote database.
func claimWorkItem(requestURL string, item *state.WorkItem, force bool) (bool, error) {
	slog.Debug("claimWorkItem", "requestURL", requestURL, "item", item)

	ok, err := checkWorkItemStatus(item, force)
	if err != nil || !ok {
		return false, err
	}

	// TODO: This is a temporary implementation to simulate the remote database
	// 		 The claim is successful 50% of the time
	maybeClaimed := rand.Intn(2) == 0

	if maybeClaimed {
		s := state.NewState(item.UUID, state.CLAIMED, time.Now())
		if err = s.Save(); err != nil {
			// The local and remote states are now out of sync
			// This is a critical error and should be handled by an operator
			slog.Error("could not update state", "error", err)
			return false, err
		}

		slog.Info("claimed work item", "uuid", item.UUID)
		return true, nil
	}

	slog.Debug("work item not claimed", "uuid", item.UUID)
	return false, nil
}
