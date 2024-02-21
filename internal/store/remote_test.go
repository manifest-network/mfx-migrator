package store_test

import (
	"embed"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/liftedinit/mfx-migrator/internal/testutils"

	"github.com/liftedinit/mfx-migrator/internal/state"
	"github.com/liftedinit/mfx-migrator/internal/store"
)

//go:embed testdata/work-items.json
//go:embed testdata/work-item.json
var mockData embed.FS

func setup(t *testing.T, filePath string, contentType string) *httptest.Server {
	return testutils.CreateHTTPTestServer(t, filePath, contentType, mockData)
}

func TestGetAllWorkItems(t *testing.T) {
	server := setup(t, "testdata/work-items.json", "application/json")
	defer server.Close()

	items, err := store.GetAllWorkItems(server.URL)
	require.NoError(t, err)

	workItems := items.Items
	meta := items.Meta

	require.Equal(t, 10, len(workItems))

	for _, item := range workItems {
		require.Equal(t, state.CREATED, item.Status)
		require.NotEmpty(t, item.UUID)
		require.NotEmpty(t, item.ManyHash)
		require.Empty(t, item.ManifestDatetime)
		require.Empty(t, item.ManifestHash)
	}

	require.Equal(t, 11, meta.TotalItems)
	require.Equal(t, 10, meta.ItemCount)
	require.Equal(t, 10, meta.ItemsPerPage)
	require.Equal(t, 2, meta.TotalPages)
	require.Equal(t, 1, meta.CurrentPage)
}

func TestGetAllWorkItems_NotJSON(t *testing.T) {
	server := setup(t, "testdata/work-items.json", "text/plain")
	defer server.Close()

	_, err := store.GetAllWorkItems(server.URL)
	require.Error(t, err)
}

func TestGetWorkItem(t *testing.T) {
	server := setup(t, "testdata/work-item.json", "application/json")
	defer server.Close()

	item, err := store.GetWorkItem(server.URL, uuid.MustParse("5aa19d2a-4bdf-4687-a850-1804756b3f1f"))
	require.NoError(t, err)

	require.Equal(t, state.CREATED, item.Status)
	require.NotEmpty(t, item.UUID)
	require.NotEmpty(t, item.ManyHash)
	require.Empty(t, item.ManifestDatetime)
	require.Empty(t, item.ManifestHash)
}

func TestGetWorkItem_NotJSON(t *testing.T) {
	server := setup(t, "testdata/work-item.json", "text/plain")
	defer server.Close()

	_, err := store.GetWorkItem(server.URL, uuid.MustParse("5aa19d2a-4bdf-4687-a850-1804756b3f1f"))
	require.Error(t, err)
}

// TODO: This test might currently fail because the result of claiming a work item is random.
func TestClaimWorkItemFromQueue(t *testing.T) {
	_, cleanup := testutils.SetupTempDir(t)
	defer cleanup()

	server := setup(t, "testdata/work-items.json", "application/json")
	defer server.Close()

	claimed, err := store.ClaimWorkItemFromQueue(server.URL)
	require.NoError(t, err)
	require.True(t, claimed)
}

// TODO: This test will currently fail half the time because the result of claiming a work item is random.
func TestClaimWorkItemFromUUID(t *testing.T) {
	_, cleanup := testutils.SetupTempDir(t)
	defer cleanup()

	server := setup(t, "testdata/work-item.json", "application/json")
	defer server.Close()

	claimed, err := store.ClaimWorkItemFromUUID(server.URL, uuid.MustParse("5aa19d2a-4bdf-4687-a850-1804756b3f1f"), false)
	require.NoError(t, err)
	require.True(t, claimed)
}

func TestClaimWorkItemFromUUID_UUIDNotFound(t *testing.T) {
	_, cleanup := testutils.SetupTempDir(t)
	defer cleanup()

	server := setup(t, "testdata/work-item.json", "application/json")
	defer server.Close()

	claimed, err := store.ClaimWorkItemFromUUID(server.URL, uuid.MustParse("00000000-0000-0000-0000-000000000000"), false)
	require.Error(t, err)
	require.False(t, claimed)
}
