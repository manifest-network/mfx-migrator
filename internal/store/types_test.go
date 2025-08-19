package store_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/manifest-network/mfx-migrator/internal/store"
)

func TestTypes_WorkItemStatus(t *testing.T) {
	tests := []struct {
		desc string
		s    store.WorkItemStatus
		i    int64
	}{
		{"created", store.CREATED, 1},
		{"claimed", store.CLAIMED, 2},
		{"migrating", store.MIGRATING, 3},
		{"completed", store.COMPLETED, 4},
		{"failed", store.FAILED, 5},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			require.Equal(t, tt.desc, tt.s.String())
			require.Equal(t, tt.i, tt.s.EnumIndex())
		})
	}
}

func TestTypes_WorkItem(t *testing.T) {
	now := time.Now()
	hash := "foobar"
	wi := store.WorkItem{
		Status:           store.CREATED,
		ManifestAddress:  "foo",
		ManifestHash:     &hash,
		ManifestDatetime: &now,
	}

	require.True(t, wi.Equal(wi))
	require.False(t, wi.Equal(store.WorkItem{}))
}
