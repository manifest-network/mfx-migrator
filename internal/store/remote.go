package store

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/liftedinit/mfx-migrator/internal/state"
)

// GetAllWorkItems retrieves all work items from the remote database.
func GetAllWorkItems(requestURL string) (*state.WorkItems, error) {
	slog.Debug("GetAllWorkItems", "requestURL", requestURL)

	res, err := http.Get(requestURL)
	if err != nil {
		slog.Error("error performing http get", "error", err)
		return nil, err
	}

	contentType := res.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") {
		slog.Error("unexpected content type", "Content-Type", contentType)
		return nil, fmt.Errorf("unexpected content type: %s", contentType)
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		slog.Error("http response error reading body", "error", err)
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
