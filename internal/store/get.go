package store

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/liftedinit/mfx-migrator/internal/state"
	"github.com/liftedinit/mfx-migrator/internal/utils"
)

// getAllWorkItems retrieves all work items from the remote database.
func getAllWorkItems(url string) (*state.WorkItems, error) {
	slog.Debug("getAllWorkItems", "url", url)
	body, err := utils.MakeHTTPGetRequest(url)
	if err != nil {
		return nil, err
	}

	var workItems state.WorkItems
	err = json.Unmarshal(body, &workItems)
	if err != nil {
		slog.Error("error unmarshalling json", "error", err)
		return nil, err
	}

	slog.Debug("getAllWorkItems", "allWorkItems", workItems)

	return &workItems, nil
}

// getWorkItem retrieves a work item from the remote database by UUID.
func getWorkItem(url string, uuidStr uuid.UUID) (*state.WorkItem, error) {
	slog.Debug("getWorkItem", "url", url, "uuid", uuidStr)
	body, err := utils.MakeHTTPGetRequest(url)
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

	slog.Debug("getWorkItem", "workItem", workItem)

	return &workItem, nil
}
