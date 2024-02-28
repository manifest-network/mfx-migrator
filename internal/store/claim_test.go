package store_test

import (
	"embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"

	"github.com/liftedinit/mfx-migrator/internal/store"
)

const (
	uuidv4Regex  = "[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89aAbB][a-f0-9]{3}-[a-f0-9]{12}"
	rootUrl      = "http://fakeurl:3001/api/v1/"
	neighborhood = "1"
)

var (
	migrationsUrl = rootUrl + fmt.Sprintf("neighborhoods/%s/migrations/", neighborhood)
	migrationUrl  = migrationsUrl + uuidv4Regex
)

//go:embed testdata/work-items.json
//go:embed testdata/work-item.json
//go:embed testdata/work-item-update-success.json
//go:embed testdata/work-item-update-failure.json
//go:embed testdata/work-item-update-force.json
//go:embed testdata/work-item-force.json
var mockData embed.FS

type Endpoint struct {
	method string
	url    string
	data   string
	code   int
}
type testCase struct {
	desc      string
	endpoints []Endpoint
	check     func()
}

// createJsonResponderFromFile creates a new JSON responder from a file.
func createJsonResponderFromFile(filePath string, code int) (httpmock.Responder, error) {
	inputData, err := mockData.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var jsonData interface{}
	if err = json.Unmarshal(inputData, &jsonData); err != nil {
		return nil, err
	}

	return httpmock.NewJsonResponder(code, jsonData)
}

// setupMockResponder sets up a mock responder for a given method and URL.
func setupMockResponder(t *testing.T, method, url, filePath string, code int) {
	var responder httpmock.Responder
	var err error
	if filePath != "" {
		responder, err = createJsonResponderFromFile(filePath, code)
		require.NoError(t, err)
	} else {
		responder = httpmock.NewErrorResponder(fmt.Errorf("not found"))
	}

	slog.Info("Setting up mock responder", "method", method, "url", url, "code", code)
	httpmock.RegisterResponder(method, url, responder)
}

func TestStore_Claim(t *testing.T) {
	setup(t) // TODO: Move this to a test suite setup function
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	testUrl, _ := url.Parse(rootUrl)
	rClient := resty.New().SetBaseURL(testUrl.String()).SetPathParam("neighborhood", neighborhood)
	httpmock.ActivateNonDefault(rClient.GetClient())

	var tests = []testCase{
		{"success_queue", []Endpoint{
			{"GET", migrationsUrl, "testdata/work-items.json", http.StatusOK},
			{"GET", "=~^" + migrationUrl, "testdata/work-item.json", http.StatusOK},
			{"PUT", "=~^" + migrationUrl, "testdata/work-item-update-success.json", http.StatusOK},
		}, func() {
			item, err := store.ClaimWorkItemFromQueue(rClient)
			require.NotEqual(t, uuid.Nil, item.UUID)
			require.NoError(t, err)
			require.NotNil(t, item)
		}},
		{"failure_queue", []Endpoint{
			{"GET", migrationsUrl, "testdata/work-items.json", http.StatusOK},
			{"GET", "=~^" + migrationUrl, "testdata/work-item.json", http.StatusOK},
			{"PUT", "=~^" + migrationUrl, "testdata/work-item-update-failure.json", http.StatusOK},
		}, func() {
			item, err := store.ClaimWorkItemFromQueue(rClient)
			require.Error(t, err) // unable to claim the work item
			require.Nil(t, item)
		}},
		{"success_uuid", []Endpoint{
			{"GET", "=~^" + migrationUrl, "testdata/work-item.json", http.StatusOK},
			{"PUT", "=~^" + migrationUrl, "testdata/work-item-update-success.json", http.StatusOK},
		}, func() {
			myUUID := uuid.MustParse("5aa19d2a-4bdf-4687-a850-1804756b3f1f")
			item, err := store.ClaimWorkItemFromUUID(rClient, myUUID, false)
			require.Equal(t, myUUID, item.UUID)
			require.NoError(t, err)
			require.NotNil(t, item)
		}},
		{"failure_uuid", []Endpoint{
			{"GET", migrationsUrl, "testdata/work-items.json", http.StatusOK},
			{"GET", "=~^" + migrationUrl, "testdata/work-item.json", http.StatusOK},
			{"PUT", "=~^" + migrationUrl, "testdata/work-item-update-failure.json", http.StatusOK},
		}, func() {
			item, err := store.ClaimWorkItemFromUUID(rClient, uuid.MustParse("5aa19d2a-4bdf-4687-a850-1804756b3f1f"), false)
			require.Error(t, err) // unable to claim the work item
			require.Nil(t, item)
		}},
		{"failure_uuid_not_found", []Endpoint{
			{"GET", "=~^" + migrationUrl, "", http.StatusNotFound},
		}, func() {
			item, err := store.ClaimWorkItemFromUUID(rClient, uuid.New(), false)
			require.Error(t, err) // work item not found
			require.Nil(t, item)
		}},
		{"force_uuid_succeed", []Endpoint{
			{"GET", "=~^" + migrationUrl, "testdata/work-item-force.json", http.StatusOK},
			{"PUT", "=~^" + migrationUrl, "testdata/work-item-update-force.json", http.StatusOK},
		}, func() {
			myUUID := uuid.MustParse("c726e305-089a-4a50-b6b6-c707d45221f2")
			item, err := store.ClaimWorkItemFromUUID(rClient, myUUID, true)
			require.Equal(t, myUUID, item.UUID)
			require.NoError(t, err)
			require.NotNil(t, item)
		}},
		{"force_uuid_fail", []Endpoint{
			{"GET", "=~^" + migrationUrl, "testdata/work-item-force.json", http.StatusOK},
			{"PUT", "=~^" + migrationUrl, "testdata/work-item-update-failure.json", http.StatusOK},
		}, func() {
			item, err := store.ClaimWorkItemFromUUID(rClient, uuid.MustParse("5aa19d2a-4bdf-4687-a850-1804756b3f1f"), false)
			require.Error(t, err) // work item not in the correct state to be claimed and force is false
			require.Nil(t, item)
		}},
		{"invalid_all_work_items_url", []Endpoint{
			{"GET", migrationsUrl, "", http.StatusNotFound},
		}, func() {
			_, err := store.ClaimWorkItemFromQueue(rClient)
			require.Error(t, err) // unable to list work items
		}},
		{"invalid_update_work_item_url", []Endpoint{
			{"GET", migrationsUrl, "testdata/work-items.json", http.StatusOK},
			{"GET", "=~^" + migrationUrl, "testdata/work-item.json", http.StatusOK},
			{"PUT", "=~^" + migrationUrl, "", http.StatusNotFound},
		}, func() {
			item, err := store.ClaimWorkItemFromQueue(rClient)
			require.Error(t, err) // unable to claim the work item
			require.Nil(t, item)
		}},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			for _, endpoint := range tt.endpoints {
				setupMockResponder(t, endpoint.method, endpoint.url, endpoint.data, endpoint.code)
			}

			tt.check()
		})
		httpmock.Reset()
	}
}
