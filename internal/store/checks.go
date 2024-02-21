package store

import (
	"fmt"
	"log/slog"

	"github.com/liftedinit/mfx-migrator/internal/state"
)

// checkWorkItemStatus checks if a work item is in the correct state to be claimed.
// If force is true, the check is skipped.
func checkWorkItemStatus(item *state.WorkItem, force bool) (bool, error) {
	if item == nil {
		slog.Error("work item is nil")
		return false, fmt.Errorf("work item is nil")
	}

	if item.Status != state.CREATED {
		statusStr := item.Status.String()

		if force {
			slog.Info("forcing claim of work item", "status", item.Status, "statusStr", statusStr)
		} else {
			slog.Warn("unable to claim work item", "status", item.Status, "statusStr", statusStr)
			return false, nil
		}
	}

	return true, nil
}
