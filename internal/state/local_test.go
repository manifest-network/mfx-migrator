package state_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/liftedinit/mfx-migrator/internal/testutils"

	"github.com/liftedinit/mfx-migrator/internal/state"
)

func newState() (*state.LocalState, uuid.UUID, time.Time) {
	someUUID := uuid.New()
	now := time.Now()
	s := state.NewState(someUUID, state.CREATED, now)
	return s, someUUID, now
}

func TestLocalState(t *testing.T) {
	s, someUUID, now := newState()

	// Check that the state was created correctly.
	require.Equal(t, state.Version, s.Version)
	require.Equal(t, someUUID, s.UUID)
	require.Equal(t, state.CREATED, s.Status)
	require.Equal(t, now.UTC().Truncate(time.Millisecond), s.Timestamp)
}

func TestLocalState_SaveDelete(t *testing.T) {
	_, cleanup := testutils.SetupTempDir(t)
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
	_, cleanup := testutils.SetupTempDir(t)
	defer cleanup()

	s, someUUID, _ := newState()
	err := s.Save()
	require.NoError(t, err)

	// Load the state from the file.
	loaded, err := state.LoadState(someUUID)
	require.NoError(t, err)
	require.Equal(t, s, loaded)
}

func TestLocalState_Update(t *testing.T) {
	_, cleanup := testutils.SetupTempDir(t)
	defer cleanup()

	s, someUUID, _ := newState()
	err := s.Save()
	require.NoError(t, err)

	// Update the state.
	newStatus := state.PROCESSING
	newTimestamp := time.Now().UTC()
	s.Update(newStatus, newTimestamp)
	err = s.Save()
	require.NoError(t, err)

	// Load the state from the file.
	loaded, err := state.LoadState(someUUID)
	require.NoError(t, err)
	require.Equal(t, newStatus, loaded.Status)
	require.Equal(t, newTimestamp.Truncate(time.Millisecond), loaded.Timestamp)
}

func TestLocalStatus(t *testing.T) {
	s, _, _ := newState()
	require.Equal(t, s.Status.String(), "created")
	require.Equal(t, s.Status.EnumIndex(), 1)

	s.Update(state.CLAIMED, time.Now())
	require.Equal(t, s.Status.String(), "claimed")
	require.Equal(t, s.Status.EnumIndex(), 2)

	s.Update(state.PROCESSING, time.Now())
	require.Equal(t, s.Status.String(), "processing")
	require.Equal(t, s.Status.EnumIndex(), 3)

	s.Update(state.FAILED, time.Now())
	require.Equal(t, s.Status.String(), "failed")
	require.Equal(t, s.Status.EnumIndex(), 4)
}
