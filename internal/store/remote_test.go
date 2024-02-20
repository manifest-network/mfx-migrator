package store_test

import (
	"embed"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/liftedinit/mfx-migrator/internal/state"
	"github.com/liftedinit/mfx-migrator/internal/store"
)

//go:embed testdata/work-items.json
var mockData embed.FS

// Function to create a mock server
func createMockServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

func createTestServer(t *testing.T, filePath string, contentType string) *httptest.Server {
	data, err := mockData.ReadFile(filePath)
	require.NoError(t, err)

	server := createMockServer(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", contentType)
		_, err = rw.Write(data)
		require.NoError(t, err)
	})

	return server
}

func TestGetAllWorkItems(t *testing.T) {
	server := createTestServer(t, "testdata/work-items.json", "application/json")
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
	server := createTestServer(t, "testdata/work-items.json", "text/plain")
	defer server.Close()

	_, err := store.GetAllWorkItems(server.URL)
	require.Error(t, err)
}
