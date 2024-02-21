package store

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/liftedinit/mfx-migrator/internal/state"
	"github.com/liftedinit/mfx-migrator/internal/utils"
)

// updateWorkItem tries to update the status of a work item in the remote database.
// Returns an error if the update failed.
func updateWorkItem(updateURL string, item *state.WorkItem) (bool, error) {
	slog.Debug("updateWorkItem", "updateURL", updateURL, "item", item)

	buffer, err := json.Marshal(item)
	if err != nil {
		slog.Error("error marshalling json", "error", err)
		return false, err
	}

	dataBuffer := bytes.NewBuffer(buffer)
	resp, err := utils.MakeHTTPPutRequest(updateURL, dataBuffer)
	if err != nil {
		slog.Error("error updating work item", "error", err)
		return false, err
	}

	var maybeClaimed state.UpdateResponse
	err = json.Unmarshal(resp, &maybeClaimed)
	if err != nil {
		slog.Error("error unmarshalling json", "error", err)
		return false, err
	}

	return maybeClaimed.Status, nil
}

// updateWorkItemStatus updates the status of a work item and saves the state.
func updateWorkItemStatus(updateURL string, item *state.WorkItem, status state.Status) (bool, error) {
	slog.Debug("updateWorkItemStatus", "updateURL", updateURL, "item", item, "status", status, "statusStr", status.String())

	item.Status = status
	claimed, err := updateWorkItem(updateURL, item)
	if err != nil {
		return false, err
	}
	if !claimed {
		slog.Debug("work item was not claimed on the remote database", "uuid", item.UUID)
		return false, nil
	}

	s := state.NewState(item.UUID, status, time.Now())
	if err = s.Save(); err != nil {
		// The local and remote states are now out of sync
		// This is a critical error and should be handled by an operator
		slog.Error("could not update state", "error", err)
		return false, err
	}

	slog.Info("claimed work item", "uuid", item.UUID)
	return true, nil
}
