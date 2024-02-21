package state

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"
)

// Version is the version of the state file format.
const Version = 1

// LocalState represents the state of a migration.
// Add additional fields as necessary.
type LocalState struct {
	Version   int       `json:"version"`
	UUID      uuid.UUID `json:"uuid"`
	Status    Status    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// NewState creates a new migration state.
func NewState(itemUUID uuid.UUID, status Status, timestamp time.Time) *LocalState {
	utc := timestamp.UTC()
	// Truncate the timestamp to avoid issues with JSON serialization.
	slog.Debug("Truncating timestamp to the millisecond", "timestamp", utc)
	utc = utc.Truncate(time.Millisecond)

	slog.Debug("NewState", "uuid", itemUUID, "status", status, "timestamp", utc)
	return &LocalState{
		Version:   Version,
		UUID:      itemUUID,
		Status:    status,
		Timestamp: utc,
	}
}

// fileName generates the file name for the state based on its UUID.
func (s *LocalState) fileName() string {
	return fmt.Sprintf("%s.uuid", s.UUID)
}

// Save writes the state to its .uuid file.
func (s *LocalState) Save() error {
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}
	fileName := s.fileName()
	return os.WriteFile(fileName, data, 0644)
}

// LoadState loads the state from a given .uuid file.
func LoadState(itemUUID uuid.UUID) (*LocalState, error) {
	slog.Debug("LoadState", "uuid", itemUUID)
	fileName := fmt.Sprintf("%s.uuid", itemUUID)
	data, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	var state LocalState
	err = json.Unmarshal(data, &state)
	if err != nil {
		return nil, err
	}
	return &state, nil
}

// Update updates the state and saves it.
// Modify this method according to what needs to be updated.
func (s *LocalState) Update(status Status, timestamp time.Time) {
	utc := timestamp.UTC()
	slog.Debug("Update", "status", status, "timestamp", utc)
	s.Status = status
	s.Timestamp = utc.Truncate(time.Millisecond)
}

// Delete removes the .uuid file associated with the state.
func (s *LocalState) Delete() error {
	fileName := s.fileName()
	err := os.Remove(fileName)
	if err != nil {
		return err
	}
	return nil
}
