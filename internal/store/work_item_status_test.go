package store_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/liftedinit/mfx-migrator/internal/store"
)

func TestStore_WorkItemStatus(t *testing.T) {
	tests := []struct {
		desc string
		s    store.WorkItemStatus
		i    int
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
