package localstate_test

import (
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/liftedinit/mfx-migrator/internal/localstate"
	"github.com/liftedinit/mfx-migrator/internal/store"
	"github.com/stretchr/testify/require"
)

func setup(t *testing.T) string {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "test")
	require.NoError(t, err)

	// Change the current working directory to the temporary directory
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	// Return the path of the temporary directory for cleanup purposes
	return tempDir
}

func TestSaveLoadState(t *testing.T) {
	tempDir := setup(t)

	defer os.RemoveAll(tempDir)

	someUUID := uuid.New()
	item := &store.WorkItem{
		Status:           store.CREATED,
		UUID:             someUUID,
		ManyHash:         "",
		ManifestHash:     nil,
		ManifestDatetime: nil,
	}
	err := localstate.SaveState(item)
	require.NoError(t, err)

	otherItem, err := localstate.LoadState(someUUID.String())
	require.NoError(t, err)
	require.Equal(t, item, otherItem)
}
