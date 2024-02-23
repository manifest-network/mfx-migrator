package store_test

import (
	"testing"

	"github.com/liftedinit/mfx-migrator/internal/store"
	"github.com/stretchr/testify/require"
)

func TestStore_WorkItemStatus(t *testing.T) {
	tests := []struct {
		desc string
		s    store.WorkItemStatus
		i    int
	}{
		{"created", store.CREATED, 1},
		{"claimed", store.CLAIMED, 2},
		{"processing", store.PROCESSING, 3},
		{"failed", store.FAILED, 4},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			require.Equal(t, tt.desc, tt.s.String())
			require.Equal(t, tt.i, tt.s.EnumIndex())
		})
	}
}
