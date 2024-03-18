package store_test

import (
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
	endpoints []testutils.HttpResponder
	check     func()
}

func TestStore_Claim(t *testing.T) {
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}

	testUrl, _ := url.Parse(testutils.RootUrl)
	rClient := resty.New().SetBaseURL(testUrl.String()).SetPathParam("neighborhood", testutils.Neighborhood)
	httpmock.ActivateNonDefault(rClient.GetClient())

	var tests = []testCase{
		// Successfully claim a work item from the queue
		{"success_queue", []testutils.HttpResponder{
			{Method: "GET", Url: testutils.MigrationsUrl, Responder: testutils.MustAllMigrationsGetResponder(1, store.CREATED)},
			{Method: "GET", Url: "=~^" + testutils.MigrationUrl, Responder: testutils.MustMigrationGetResponder(store.CLAIMED)},
			{Method: "PUT", Url: "=~^" + testutils.MigrationUrl, Responder: testutils.MigrationUpdateResponder},
		}, func() {
			item, err := store.ClaimWorkItemFromQueue(rClient)
			require.NotEqual(t, uuid.Nil, item.UUID)
			require.NoError(t, err)
			require.NotNil(t, item)
		}},
		// Fail to claim a work item from the queue (work item update failure)
		{"failure_queue", []testutils.HttpResponder{
			{Method: "GET", Url: testutils.MigrationsUrl, Responder: testutils.MustAllMigrationsGetResponder(1, store.CREATED)},
			{Method: "GET", Url: "=~^" + testutils.MigrationUrl, Responder: testutils.MustMigrationGetResponder(store.CREATED)},
			{Method: "PUT", Url: "=~^" + testutils.MigrationUrl, Responder: testutils.NotFoundResponder},
		}, func() {
			item, err := store.ClaimWorkItemFromQueue(rClient)
			require.Error(t, err)
			require.ErrorContains(t, err, "could not claim work item")
			require.ErrorContains(t, err, "status code: 404")
			require.Nil(t, item)
		}},
		// No work items available in the queue
		{"no_item_queue", []testutils.HttpResponder{
			{Method: "GET", Url: testutils.MigrationsUrl, Responder: testutils.MustAllMigrationsGetResponder(0, store.CREATED)},
		}, func() {
			item, err := store.ClaimWorkItemFromQueue(rClient)
			require.NoError(t, err) // no work items available
			require.Nil(t, item)
		}},
		// Successfully claim a work item by UUID
		{"success_uuid", []testutils.HttpResponder{
			{Method: "GET", Url: "=~^" + testutils.MigrationUrl, Responder: testutils.MustMigrationGetResponder(store.CREATED)},
			{Method: "PUT", Url: "=~^" + testutils.MigrationUrl, Responder: testutils.MigrationUpdateResponder},
		}, func() {
			myUUID := uuid.MustParse("5aa19d2a-4bdf-4687-a850-1804756b3f1f")
			item, err := store.ClaimWorkItemFromUUID(rClient, myUUID, false)
			require.Equal(t, myUUID, item.UUID)
			require.NoError(t, err)
			require.NotNil(t, item)
		}},
		// Fail to claim a work item by UUID (work item update failure)
		{"failure_uuid", []testutils.HttpResponder{
			{Method: "GET", Url: testutils.MigrationsUrl, Responder: testutils.MustAllMigrationsGetResponder(1, store.CREATED)},
			{Method: "GET", Url: "=~^" + testutils.MigrationUrl, Responder: testutils.MustMigrationGetResponder(store.CREATED)},
			{Method: "PUT", Url: "=~^" + testutils.MigrationUrl, Responder: testutils.NotFoundResponder},
		}, func() {
			item, err := store.ClaimWorkItemFromUUID(rClient, uuid.MustParse("5aa19d2a-4bdf-4687-a850-1804756b3f1f"), false)
			require.Error(t, err) // unable to claim the work item
			require.ErrorContains(t, err, "could not claim work item")
			require.ErrorContains(t, err, "status code: 404")
			require.Nil(t, item)
		}},
		// Fail to claim a work item by UUID (work item UUID not found)
		{"failure_uuid_not_found", []testutils.HttpResponder{
			{Method: "GET", Url: "=~^" + testutils.MigrationUrl, Responder: testutils.NotFoundResponder},
		}, func() {
			item, err := store.ClaimWorkItemFromUUID(rClient, uuid.New(), false)
			require.Error(t, err) // work item not found
			require.ErrorContains(t, err, "could not get work item")
			require.ErrorContains(t, err, "status code: 404")
			require.Nil(t, item)
		}},
		// Successfully claim a work item by UUID (forced claim, work item already claimed)
		{"force_uuid_succeed", []testutils.HttpResponder{
			{Method: "GET", Url: "=~^" + testutils.MigrationUrl, Responder: testutils.MustMigrationGetResponder(store.MIGRATING)},
			{Method: "PUT", Url: "=~^" + testutils.MigrationUrl, Responder: testutils.MigrationUpdateResponder},
		}, func() {
			myUUID := uuid.MustParse("5aa19d2a-4bdf-4687-a850-1804756b3f1f")
			item, err := store.ClaimWorkItemFromUUID(rClient, myUUID, true)
			require.Equal(t, myUUID, item.UUID)
			require.Equal(t, item.Status, store.CLAIMED)
			require.NoError(t, err)
			require.NotNil(t, item)
		}},
		// Test force claim of a failed item clears the previous error
		{"force_uuid_clear_previous", []testutils.HttpResponder{
			{Method: "GET", Url: "=~^" + testutils.MigrationUrl, Responder: testutils.MustMigrationGetResponder(store.FAILED)},
			{Method: "PUT", Url: "=~^" + testutils.MigrationUrl, Responder: testutils.MigrationUpdateResponder},
		}, func() {
			item, err := store.ClaimWorkItemFromUUID(rClient, uuid.MustParse("5aa19d2a-4bdf-4687-a850-1804756b3f1f"), true)
			require.Equal(t, store.CLAIMED, item.Status)
			require.Nil(t, item.Error)
			require.NoError(t, err)
			require.NotNil(t, item)
		}},
		// Fail to claim a work item by UUID (forced claim, work item update failure)
		{"force_uuid_fail", []testutils.HttpResponder{
			{Method: "GET", Url: "=~^" + testutils.MigrationUrl, Responder: testutils.MustMigrationGetResponder(store.MIGRATING)},
			{Method: "PUT", Url: "=~^" + testutils.MigrationUrl, Responder: testutils.MigrationUpdateResponder},
		}, func() {
			item, err := store.ClaimWorkItemFromUUID(rClient, uuid.MustParse("5aa19d2a-4bdf-4687-a850-1804756b3f1f"), false)
			require.Error(t, err)
			require.ErrorContains(t, err, "unable to claim work item, invalid state")
			require.Nil(t, item)
		}},
		// Fail to claim a work item by UUID (response is not a work item)
		{"invalid_work_item", []testutils.HttpResponder{
			{Method: "GET", Url: "=~^" + testutils.MigrationUrl, Responder: testutils.GarbageResponder},
		}, func() {
			item, err := store.ClaimWorkItemFromUUID(rClient, uuid.New(), false)
			require.Error(t, err)
			require.ErrorContains(t, err, "cannot unmarshal")
			require.Nil(t, item)
		}},
		// Fail to claim a work item from the queue (work item list is invalid)
		{"invalid_work_items", []testutils.HttpResponder{
			{Method: "GET", Url: testutils.MigrationsUrl, Responder: testutils.GarbageResponder},
		}, func() {
			item, err := store.ClaimWorkItemFromQueue(rClient)
			require.Error(t, err)
			require.ErrorContains(t, err, "cannot unmarshal")
			require.Nil(t, item)
		}},
		// Fail to claim a work item from the queue (invalid work item list URL)
		{"invalid_all_work_items_url", []testutils.HttpResponder{
			{Method: "GET", Url: testutils.MigrationsUrl, Responder: testutils.NotFoundResponder},
		}, func() {
			_, err := store.ClaimWorkItemFromQueue(rClient)
			require.Error(t, err) // unable to list work items
			require.ErrorContains(t, err, "could not get all work items")
			require.ErrorContains(t, err, "status code: 404")
		}},
		// Fail to claim a work item from the queue (invalid work item update URL)
		{"invalid_update_work_item_url", []testutils.HttpResponder{
			{Method: "GET", Url: testutils.MigrationsUrl, Responder: testutils.MustAllMigrationsGetResponder(1, store.CREATED)},
			{Method: "GET", Url: "=~^" + testutils.MigrationUrl, Responder: testutils.MustMigrationGetResponder(store.CREATED)},
			{Method: "PUT", Url: "=~^" + testutils.MigrationUrl, Responder: testutils.NotFoundResponder},
		}, func() {
			item, err := store.ClaimWorkItemFromQueue(rClient)
			require.Error(t, err)
			require.ErrorContains(t, err, "could not claim work item")
			require.ErrorContains(t, err, "error updating remote work item")
			require.ErrorContains(t, err, "status code: 404")
			require.Nil(t, item)
		}},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			for _, endpoint := range tt.endpoints {
				httpmock.RegisterResponder(endpoint.Method, endpoint.Url, endpoint.Responder)
			}

			tt.check()
		})
		httpmock.Reset()
	}
}
