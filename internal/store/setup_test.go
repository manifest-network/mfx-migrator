package store_test

import (
	"embed"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/liftedinit/mfx-migrator/internal/common"
	"github.com/stretchr/testify/require"

	"github.com/liftedinit/mfx-migrator/internal/testutils"
)

//go:embed testdata/work-items.json
//go:embed testdata/work-item.json
//go:embed testdata/work-item-update-success.json
//go:embed testdata/work-item-update-failure.json
var mockData embed.FS

var (
	itemUUID = uuid.MustParse("5aa19d2a-4bdf-4687-a850-1804756b3f1f")
)

const (
	workItemsPath             = "testdata/work-items.json"
	workItemPath              = "testdata/work-item.json"
	workItemUpdateSuccessPath = "testdata/work-item-update-success.json"
	workItemUpdateFailurePath = "testdata/work-item-update-failure.json"
)

func setup(t *testing.T) *httptest.Server {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)
	routes := map[string]http.HandlerFunc{
		"/get-all-work": func(rw http.ResponseWriter, req *http.Request) {
			data, err := mockData.ReadFile(workItemsPath)
			require.NoError(t, err)
			rw.Header().Set(common.ContentType, common.ContentTypeJSON)
			_, err = rw.Write(data)
			require.NoError(t, err)
		},
		"/get-work": func(rw http.ResponseWriter, req *http.Request) {
			data, err := mockData.ReadFile(workItemPath)
			require.NoError(t, err)
			rw.Header().Set(common.ContentType, common.ContentTypeJSON)
			_, err = rw.Write(data)
			require.NoError(t, err)
		},
		"/update-work": func(rw http.ResponseWriter, req *http.Request) {
			if req.Method != http.MethodPut {
				http.Error(rw, "Invalid request method", http.StatusMethodNotAllowed)
				return
			}
			data, err := mockData.ReadFile(workItemUpdateSuccessPath)
			require.NoError(t, err)
			rw.Header().Set(common.ContentType, common.ContentTypeJSON)
			_, err = rw.Write(data)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}
		},
		"/update-work-failure": func(rw http.ResponseWriter, req *http.Request) {
			if req.Method != http.MethodPut {
				http.Error(rw, "Invalid request method", http.StatusMethodNotAllowed)
				return
			}
			data, err := mockData.ReadFile(workItemUpdateFailurePath)
			require.NoError(t, err)
			rw.Header().Set(common.ContentType, common.ContentTypeJSON)
			_, err = rw.Write(data)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}
		},
	}

	return testutils.CreateHTTPTestServer(routes)
}
