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

type LocalStatus int

const (
	CLAIMING LocalStatus = iota + 1 // Work is available, but not yet claimed
	CLAIMED                         // Work has been claimed by a worker
	IN_PROGRESS
)

// String returns the string representation of a LocalStatus.
func (s LocalStatus) String() string {
	return [...]string{"created", "claimed", "in progress"}[s-1]
}

// EnumIndex returns the enum index of a LocalStatus.
func (s LocalStatus) EnumIndex() int {
	return int(s)
}

// LocalState represents the state of a migration.
// Add additional fields as necessary.
type LocalState struct {
	Version   int         `json:"version"`
	UUID      uuid.UUID   `json:"uuid"`
	Status    LocalStatus `json:"status"`
	Timestamp time.Time   `json:"timestamp"`
}

// NewState creates a new migration state.
func NewState(uuid uuid.UUID, status LocalStatus, timestamp time.Time) *LocalState {
	slog.Debug("NewState", "uuid", uuid, "status", status, "timestamp", timestamp)
	return &LocalState{
		Version:   Version,
		UUID:      uuid,
		Status:    status,
		Timestamp: timestamp,
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
func LoadState(uuid string) (*LocalState, error) {
	slog.Debug("LoadState", "uuid", uuid)
	fileName := fmt.Sprintf("%s.uuid", uuid)
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
func (s *LocalState) Update(status LocalStatus, timestamp time.Time) error {
	slog.Debug("Update", "status", status, "timestamp", timestamp)
	s.Status = status
	s.Timestamp = timestamp
	return s.Save()
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
