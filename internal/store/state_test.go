package store_test

import (
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/liftedinit/mfx-migrator/testutils"

	"github.com/liftedinit/mfx-migrator/internal/store"
)

func TestSaveLoadState(t *testing.T) {
	tempDir := testutils.SetupTmpDir(t)
	defer os.RemoveAll(tempDir)

	someUUID := uuid.New()
	item := &store.WorkItem{
		Status:           store.CREATED,
		UUID:             someUUID,
		ManyHash:         "",
		ManifestHash:     nil,
		ManifestDatetime: nil,
	}
	err := store.SaveState(item)
	require.NoError(t, err)

	otherItem, err := store.LoadState(someUUID.String())
	require.NoError(t, err)
	require.Equal(t, item, otherItem)
}
