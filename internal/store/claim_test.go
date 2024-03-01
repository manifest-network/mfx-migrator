package store_test

import (
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"

	"github.com/liftedinit/mfx-migrator/testutils"

	"github.com/liftedinit/mfx-migrator/internal/store"
)

type testCase struct {
	desc      string
	endpoints []testutils.Endpoint
	check     func()
}

func TestStore_Claim(t *testing.T) {
	tempdir := testutils.SetupTmpDir(t)
	defer os.RemoveAll(tempdir)

	testUrl, _ := url.Parse(testutils.RootUrl)
	rClient := resty.New().SetBaseURL(testUrl.String()).SetPathParam("neighborhood", testutils.Neighborhood)
	httpmock.ActivateNonDefault(rClient.GetClient())

	var tests = []testCase{
		// Successfully claim a work item from the queue
		{"success_queue", []testutils.Endpoint{
			{Method: "GET", Url: testutils.MigrationsUrl, Data: "testdata/work-items.json", Code: http.StatusOK},
			{Method: "GET", Url: "=~^" + testutils.MigrationUrl, Data: "testdata/work-item.json", Code: http.StatusOK},
			{Method: "PUT", Url: "=~^" + testutils.MigrationUrl, Data: "testdata/work-item-update-success.json", Code: http.StatusOK},
		}, func() {
			item, err := store.ClaimWorkItemFromQueue(rClient)
			require.NotEqual(t, uuid.Nil, item.UUID)
			require.NoError(t, err)
			require.NotNil(t, item)
		}},
		// Fail to claim a work item from the queue (work item update failure)
		{"failure_queue", []testutils.Endpoint{
			{Method: "GET", Url: testutils.MigrationsUrl, Data: "testdata/work-items.json", Code: http.StatusOK},
			{Method: "GET", Url: "=~^" + testutils.MigrationUrl, Data: "testdata/work-item.json", Code: http.StatusOK},
			{Method: "PUT", Url: "=~^" + testutils.MigrationUrl, Data: "testdata/work-item-update-failure.json", Code: http.StatusOK},
		}, func() {
			item, err := store.ClaimWorkItemFromQueue(rClient)
			require.Error(t, err) // unable to claim the work item
			require.Nil(t, item)
		}},
		// No work items available in the queue
		{"no_item_queue", []testutils.Endpoint{
			{Method: "GET", Url: testutils.MigrationsUrl, Data: "testdata/no-work-items-available.json", Code: http.StatusOK},
		}, func() {
			item, err := store.ClaimWorkItemFromQueue(rClient)
			require.NoError(t, err) // no work items available
			require.Nil(t, item)
		}},
		// Successfully claim a work item by UUID
		{"success_uuid", []testutils.Endpoint{
			{Method: "GET", Url: "=~^" + testutils.MigrationUrl, Data: "testdata/work-item.json", Code: http.StatusOK},
			{Method: "PUT", Url: "=~^" + testutils.MigrationUrl, Data: "testdata/work-item-update-success.json", Code: http.StatusOK},
		}, func() {
			myUUID := uuid.MustParse("5aa19d2a-4bdf-4687-a850-1804756b3f1f")
			item, err := store.ClaimWorkItemFromUUID(rClient, myUUID, false)
			require.Equal(t, myUUID, item.UUID)
			require.NoError(t, err)
			require.NotNil(t, item)
		}},
		// Fail to claim a work item by UUID (work item update failure)
		{"failure_uuid", []testutils.Endpoint{
			{Method: "GET", Url: testutils.MigrationsUrl, Data: "testdata/work-items.json", Code: http.StatusOK},
			{Method: "GET", Url: "=~^" + testutils.MigrationUrl, Data: "testdata/work-item.json", Code: http.StatusOK},
			{Method: "PUT", Url: "=~^" + testutils.MigrationUrl, Data: "testdata/work-item-update-failure.json", Code: http.StatusOK},
		}, func() {
			item, err := store.ClaimWorkItemFromUUID(rClient, uuid.MustParse("5aa19d2a-4bdf-4687-a850-1804756b3f1f"), false)
			require.Error(t, err) // unable to claim the work item
			require.Nil(t, item)
		}},
		// Fail to claim a work item by UUID (work item UUID not found)
		{"failure_uuid_not_found", []testutils.Endpoint{
			{Method: "GET", Url: "=~^" + testutils.MigrationUrl, Code: http.StatusNotFound},
		}, func() {
			item, err := store.ClaimWorkItemFromUUID(rClient, uuid.New(), false)
			require.Error(t, err) // work item not found
			require.Nil(t, item)
		}},
		// Successfully claim a work item by UUID (forced claim, work item already claimed)
		{"force_uuid_succeed", []testutils.Endpoint{
			{Method: "GET", Url: "=~^" + testutils.MigrationUrl, Data: "testdata/work-item-force.json", Code: http.StatusOK},
			{Method: "PUT", Url: "=~^" + testutils.MigrationUrl, Data: "testdata/work-item-update-force.json", Code: http.StatusOK},
		}, func() {
			myUUID := uuid.MustParse("c726e305-089a-4a50-b6b6-c707d45221f2")
			item, err := store.ClaimWorkItemFromUUID(rClient, myUUID, true)
			require.Equal(t, myUUID, item.UUID)
			require.NoError(t, err)
			require.NotNil(t, item)
		}},
		// Fail to claim a work item by UUID (forced claim, work item update failure)
		{"force_uuid_fail", []testutils.Endpoint{
			{Method: "GET", Url: "=~^" + testutils.MigrationUrl, Data: "testdata/work-item-force.json", Code: http.StatusOK},
			{Method: "PUT", Url: "=~^" + testutils.MigrationUrl, Data: "testdata/work-item-update-failure.json", Code: http.StatusOK},
		}, func() {
			item, err := store.ClaimWorkItemFromUUID(rClient, uuid.MustParse("5aa19d2a-4bdf-4687-a850-1804756b3f1f"), false)
			require.Error(t, err) // work item not in the correct state to be claimed and force is false
			require.Nil(t, item)
		}},
		// Fail to claim a work item by UUID (response is not a work item)
		{"invalid_work_item", []testutils.Endpoint{
			{Method: "GET", Url: "=~^" + testutils.MigrationUrl, Data: "testdata/garbage.json", Code: http.StatusOK},
		}, func() {
			item, err := store.ClaimWorkItemFromUUID(rClient, uuid.New(), false)
			slog.Info("item", "item", item)
			require.Error(t, err) // work item is invalid
			require.Nil(t, item)
		}},
		// Fail to claim a work item from the queue (work item list is invalid)
		{"invalid_work_items", []testutils.Endpoint{
			{Method: "GET", Url: testutils.MigrationsUrl, Data: "testdata/garbage.json", Code: http.StatusOK},
		}, func() {
			item, err := store.ClaimWorkItemFromQueue(rClient)
			require.Error(t, err) // unable to list work items
			require.Nil(t, item)
		}},
		// Fail to claim a work item from the queue (invalid work item list URL)
		{"invalid_all_work_items_url", []testutils.Endpoint{
			{Method: "GET", Url: testutils.MigrationsUrl, Code: http.StatusNotFound},
		}, func() {
			_, err := store.ClaimWorkItemFromQueue(rClient)
			require.Error(t, err) // unable to list work items
		}},
		// Fail to claim a work item from the queue (invalid work item update URL)
		{"invalid_update_work_item_url", []testutils.Endpoint{
			{Method: "GET", Url: testutils.MigrationsUrl, Data: "testdata/work-items.json", Code: http.StatusOK},
			{Method: "GET", Url: "=~^" + testutils.MigrationUrl, Data: "testdata/work-item.json", Code: http.StatusOK},
			{Method: "PUT", Url: "=~^" + testutils.MigrationUrl, Code: http.StatusNotFound},
		}, func() {
			item, err := store.ClaimWorkItemFromQueue(rClient)
			require.Error(t, err) // unable to claim the work item
			require.Nil(t, item)
		}},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			for _, endpoint := range tt.endpoints {
				testutils.SetupMockResponder(t, endpoint.Method, endpoint.Url, endpoint.Data, endpoint.Code)
			}

			tt.check()
		})
		httpmock.Reset()
	}
}
