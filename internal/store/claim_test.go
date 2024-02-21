package store_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/liftedinit/mfx-migrator/internal/store"
	"github.com/liftedinit/mfx-migrator/internal/testutils"
	"github.com/stretchr/testify/require"
)

func TestClaimWorkItemFromQueue(t *testing.T) {
	_, cleanup := testutils.SetupTempDir(t)
	defer cleanup()

	server := setup(t)
	defer server.Close()

	remoteStore := store.NewRemoteStore(&store.RouterImpl{})
	claimed, err := remoteStore.ClaimWorkItemFromQueue(server.URL)
	require.NoError(t, err)
	require.True(t, claimed)
}

func TestClaimWorkItemFromUUID(t *testing.T) {
	_, cleanup := testutils.SetupTempDir(t)
	defer cleanup()

	server := setup(t)
	defer server.Close()

	remoteStore := store.NewRemoteStore(&store.RouterImpl{})
	claimed, err := remoteStore.ClaimWorkItemFromUUID(server.URL, itemUUID, false)
	require.NoError(t, err)
	require.True(t, claimed)
}

func TestClaimWorkItemFromUUID_UUIDNotFound(t *testing.T) {
	_, cleanup := testutils.SetupTempDir(t)
	defer cleanup()

	server := setup(t)
	defer server.Close()

	remoteStore := store.NewRemoteStore(&store.RouterImpl{})
	claimed, err := remoteStore.ClaimWorkItemFromUUID(server.URL, uuid.Nil, false)
	require.Error(t, err)
	require.False(t, claimed)
}

func TestClaimWorkItem_UpdateFailure(t *testing.T) {
	_, cleanup := testutils.SetupTempDir(t)
	defer cleanup()

	server := setup(t)
	defer server.Close()

	remoteStore := store.NewRemoteStore(&MockRouterImpl{})
	claimed, err := remoteStore.ClaimWorkItemFromUUID(server.URL, itemUUID, false)
	require.NoError(t, err)
	require.False(t, claimed)
}
