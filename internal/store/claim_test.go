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
			{Method: "PUT", Url: testutils.ClaimUrl, Responder: testutils.MigrationClaimResponder(1, store.CLAIMED)},
		}, func() {
			items, err := store.ClaimWorkItemFromQueue(rClient)
			require.NotEmpty(t, items)
			require.NotEqual(t, uuid.Nil, items[0].UUID)
			require.NoError(t, err)
			require.NotNil(t, items)
		}},
		// No work items available in the queue
		{"no_item_queue", []testutils.HttpResponder{
			{Method: "PUT", Url: testutils.ClaimUrl, Responder: testutils.MigrationClaimResponder(0, store.CLAIMED)},
		}, func() {
			item, err := store.ClaimWorkItemFromQueue(rClient)
			require.NoError(t, err) // no work items available
			require.Empty(t, item)
		}},
		// Successfully claim a work item by UUID
		{"success_uuid", []testutils.HttpResponder{
			{Method: "PUT", Url: "=~^" + testutils.ClaimUuidUrl, Responder: testutils.MigrationClaimOneResponder(store.CLAIMED)},
		}, func() {
			myUUID := uuid.MustParse("5aa19d2a-4bdf-4687-a850-1804756b3f1f")
			item, err := store.ClaimWorkItemFromUUID(rClient, myUUID, false)
			require.NoError(t, err)
			require.NotNil(t, item)
			require.Equal(t, myUUID, item.UUID)
		}},
		// Fail to claim a work item by UUID (work item UUID not found)
		{"failure_uuid_not_found", []testutils.HttpResponder{
			{Method: "PUT", Url: "=~^" + testutils.ClaimUuidUrl, Responder: testutils.NotFoundResponder},
		}, func() {
			item, err := store.ClaimWorkItemFromUUID(rClient, uuid.New(), false)
			require.Error(t, err) // work item not found
			require.ErrorContains(t, err, "error claiming work item")
			require.ErrorContains(t, err, "status code: 404")
			require.Nil(t, item)
		}},
		// Fail to claim a work item by UUID (response is not a work item)
		{"invalid_work_item", []testutils.HttpResponder{
			{Method: "PUT", Url: "=~^" + testutils.ClaimUuidUrl, Responder: testutils.GarbageResponder},
		}, func() {
			item, err := store.ClaimWorkItemFromUUID(rClient, uuid.New(), false)
			require.Error(t, err)
			require.ErrorContains(t, err, "cannot unmarshal")
			require.Nil(t, item)
		}},
		// Fail to claim a work item from the queue (work item list is invalid)
		{"invalid_work_items", []testutils.HttpResponder{
			{Method: "PUT", Url: testutils.ClaimUrl, Responder: testutils.GarbageResponder},
		}, func() {
			item, err := store.ClaimWorkItemFromQueue(rClient)
			require.Error(t, err)
			require.ErrorContains(t, err, "cannot unmarshal")
			require.Nil(t, item)
		}},
		// Fail to claim a work item from the queue (invalid work item list URL)
		{"invalid_all_work_items_url", []testutils.HttpResponder{
			{Method: "PUT", Url: testutils.ClaimUrl, Responder: testutils.NotFoundResponder},
		}, func() {
			_, err := store.ClaimWorkItemFromQueue(rClient)
			require.Error(t, err) // unable to list work items
			require.ErrorContains(t, err, "error claiming work items")
			require.ErrorContains(t, err, "status code: 404")
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
