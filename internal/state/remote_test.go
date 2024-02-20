package state_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/liftedinit/mfx-migrator/internal/state"
)

func TestRemoteStatus(t *testing.T) {
	s := state.CREATED
	require.Equal(t, s.String(), "created")
	require.Equal(t, s.EnumIndex(), 1)

	s = state.FAILED
	require.Equal(t, s.String(), "failed")
	require.Equal(t, s.EnumIndex(), 2)
}
