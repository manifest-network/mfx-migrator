package store_test

import (
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/manifest-network/mfx-migrator/internal/store"
)

func TestSaveLoadState(t *testing.T) {
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}

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
