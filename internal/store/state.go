package store

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
)

func SaveState(item *WorkItem) error {
	slog.Debug("saving state", "item", item)

	// Convert the WorkItem to JSON
	data, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("failed to marshal work item: %w", err)
	}

	// Create a new file with the UUID of the WorkItem as the filename
	file, err := os.Create(fmt.Sprintf("%s.json", item.UUID))
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Write the JSON data to the file
	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}

func LoadState(uuid string) (*WorkItem, error) {
	slog.Debug("loading state", "uuid", uuid)

	// Open the file with the UUID as the filename
	file, err := os.Open(fmt.Sprintf("%s.json", uuid))
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read the file content
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Convert the JSON data to a WorkItem
	var item WorkItem
	err = json.Unmarshal(data, &item)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal work item: %w", err)
	}

	return &item, nil
}
