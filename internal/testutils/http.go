package testutils

import (
	"embed"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

// Function to create a mock server
func createMockServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

func CreateHTTPTestServer(t *testing.T, filePath string, contentType string, mockData embed.FS) *httptest.Server {
	data, err := mockData.ReadFile(filePath)
	require.NoError(t, err)

	server := createMockServer(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", contentType)
		_, err = rw.Write(data)
		require.NoError(t, err)
	})

	return server
}
