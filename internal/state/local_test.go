package state_test

import (
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/liftedinit/mfx-migrator/internal/state"
)

func setupTest(t *testing.T) (string, func()) {
	// Create a temporary directory.
	tempDir, err := os.MkdirTemp("", "test")
	require.NoError(t, err)

	// Change the current working directory to the temporary directory so that the state files are written there
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	return tempDir, func() {
		// Cleanup function to remove the temporary directory
		os.RemoveAll(tempDir)
	}
}

func newState() (*state.LocalState, uuid.UUID, time.Time) {
	someUUID := uuid.New()
	now := time.Now()
	s := state.NewState(someUUID, state.CLAIMING, now)
	return s, someUUID, now
}

func TestLocalState(t *testing.T) {
	s, someUUID, now := newState()

	// Check that the state was created correctly.
	require.Equal(t, state.Version, s.Version)
	require.Equal(t, someUUID, s.UUID)
	require.Equal(t, state.CLAIMING, s.Status)
	require.Equal(t, now.Truncate(time.Millisecond), s.Timestamp)
}

func TestLocalState_SaveDelete(t *testing.T) {
	_, cleanup := setupTest(t)
	defer cleanup()

	s, someUUID, _ := newState()
	filename := someUUID.String() + ".uuid"

	err := s.Save()
	require.NoError(t, err)
	require.FileExists(t, filename)

	// Delete the state file.
	err = s.Delete()
	require.NoError(t, err)
	require.NoFileExists(t, filename)
}

func TestLocalState_SaveLoad(t *testing.T) {
	_, cleanup := setupTest(t)
	defer cleanup()

	s, someUUID, _ := newState()
	err := s.Save()
	require.NoError(t, err)

	// Load the state from the file.
	loaded, err := state.LoadState(someUUID.String())
	require.NoError(t, err)
	require.Equal(t, s, loaded)
}

func TestLocalState_Update(t *testing.T) {
	_, cleanup := setupTest(t)
	defer cleanup()

	s, someUUID, _ := newState()
	err := s.Save()
	require.NoError(t, err)

	// Update the state.
	newStatus := state.CLAIMED
	newTimestamp := time.Now()
	s.Update(newStatus, newTimestamp)
	err = s.Save()
	require.NoError(t, err)

	// Load the state from the file.
	loaded, err := state.LoadState(someUUID.String())
	require.NoError(t, err)
	require.Equal(t, newStatus, loaded.Status)
	require.Equal(t, newTimestamp.Truncate(time.Millisecond), loaded.Timestamp)
}

func TestLocalStatus(t *testing.T) {
	s, _, _ := newState()
	require.Equal(t, s.Status.String(), "claiming")
	require.Equal(t, s.Status.EnumIndex(), 1)

	s.Update(state.CLAIMED, time.Now())
	require.Equal(t, s.Status.String(), "claimed")
	require.Equal(t, s.Status.EnumIndex(), 2)

	s.Update(state.IN_PROGRESS, time.Now())
	require.Equal(t, s.Status.String(), "in progress")
	require.Equal(t, s.Status.EnumIndex(), 3)
}
