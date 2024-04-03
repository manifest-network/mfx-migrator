package testutils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jarcoal/httpmock"

	"github.com/liftedinit/mfx-migrator/internal/many"
	"github.com/liftedinit/mfx-migrator/internal/store"
)

// TODO: Randomize parameters. Fuzzy testing.

const (
	Uuid            = "5aa19d2a-4bdf-4687-a850-1804756b3f1f"
	ManyFrom        = "maffbahksdwaqeenayy2gxke32hgb7aq4ao4wt745lsfs6wijp"
	ManySymbol      = "dummy"
	ManyHash        = "d1e60bf3bbbe497448498f942d340b872a89046854827dc43dd703ccbf7a8c78"
	ManifestAddress = "manifest1jjzy5en2000728mzs3wn86a6u6jpygzajj2fg2"
)

var CreatedDate = time.Date(2024, time.March, 1, 16, 54, 2, 651000000, time.UTC) // "2024-03-01T16:54:02.651Z"

type HttpResponder struct {
	Method    string
	Url       string
	Responder func(r *http.Request) (*http.Response, error)
}

var AuthResponder, _ = httpmock.NewJsonResponder(http.StatusOK, map[string]string{"access_token": "ya29.Gl0UBZ3"})

func MustNewTransactionResponseResponder(amount string) httpmock.Responder {
	response := many.TxInfo{Arguments: many.Arguments{
		From:   ManyFrom,
		To:     many.IllegalAddr,
		Amount: amount,
		Symbol: ManySymbol,
		Memo:   []string{Uuid, ManifestAddress},
	}}
	var transactionResponseResponder, err = httpmock.NewJsonResponder(http.StatusOK, response)
	if err != nil {
		panic(err)
	}
	return transactionResponseResponder
}

func getClaimedItems(nb uint, status store.WorkItemStatus) []*store.WorkItem {
	var items []*store.WorkItem = nil
	for i := 0; i < int(nb); i++ {
		items = append(items, &store.WorkItem{
			Status:           status,
			CreatedDate:      &CreatedDate,
			UUID:             uuid.MustParse(Uuid),
			ManyHash:         ManyHash,
			ManifestAddress:  ManifestAddress,
			ManifestHash:     nil,
			ManifestDatetime: nil,
		})
	}

	return items
}

func MigrationClaimResponder(nb uint, status store.WorkItemStatus) httpmock.Responder {
	if nb > 1 {
		panic("nb must be <= 1")
	}
	items := getClaimedItems(nb, status)

	var migrationClaimResponder, err = httpmock.NewJsonResponder(http.StatusOK, items)
	if err != nil {
		panic(err)
	}
	return migrationClaimResponder
}

func MigrationClaimOneResponder(status store.WorkItemStatus) httpmock.Responder {
	items := getClaimedItems(1, status)
	var migrationClaimResponder, err = httpmock.NewJsonResponder(http.StatusOK, items[0])
	if err != nil {
		panic(err)
	}
	return migrationClaimResponder
}

var NotFoundResponder, _ = httpmock.NewJsonResponder(http.StatusNotFound, nil)
var GarbageResponder, _ = httpmock.NewJsonResponder(http.StatusOK, "{\"foo\": \"bar\"")

func MustMigrationGetResponder(status store.WorkItemStatus) httpmock.Responder {
	var failedErr *string
	sErr := "some error"
	if status == store.FAILED {
		failedErr = &sErr
	}
	response := store.WorkItem{
		Status:           status,
		CreatedDate:      &CreatedDate,
		UUID:             uuid.MustParse(Uuid),
		ManyHash:         ManyHash,
		ManifestAddress:  ManifestAddress,
		ManifestHash:     nil,
		ManifestDatetime: nil,
		Error:            failedErr,
	}
	var migrationGetResponder, err = httpmock.NewJsonResponder(http.StatusOK, response)
	if err != nil {
		panic(err)
	}
	return migrationGetResponder
}

var MigrationUpdateResponder = func(r *http.Request) (*http.Response, error) {
	if r.Method != "PUT" {
		return httpmock.NewStringResponse(http.StatusMethodNotAllowed, ""), nil
	}

	// Decode the request body
	var item store.WorkItem

	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		return nil, fmt.Errorf("error decoding request body: %v", err)
	}
	defer r.Body.Close()

	response := store.WorkItemUpdateResponse{
		Status: item.Status,
	}

	// Check the status of the work item
	switch item.Status {
	case store.CLAIMED:
		return httpmock.NewJsonResponse(200, response)
	case store.MIGRATING:
		return httpmock.NewJsonResponse(200, response)
	case store.COMPLETED:
		if item.ManifestHash == nil {
			return nil, fmt.Errorf("manifestHash is nil")
		}
		if item.ManifestDatetime == nil {
			return nil, fmt.Errorf("manifestDatetime is empty")
		}
		response.ManifestHash = item.ManifestHash
		response.ManifestDatetime = item.ManifestDatetime
		return httpmock.NewJsonResponse(200, response)
	case store.FAILED:
		if item.Error == nil {
			return nil, fmt.Errorf("error is nil")
		}
		response.Error = item.Error
		return httpmock.NewJsonResponse(200, response)
	default:
		return nil, fmt.Errorf("invalid status: %v", item.Status)
	}
}
