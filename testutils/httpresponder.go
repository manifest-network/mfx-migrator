package testutils

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jarcoal/httpmock"

	"github.com/liftedinit/mfx-migrator/internal/many"
	"github.com/liftedinit/mfx-migrator/internal/store"
)

const (
	CreatedDate     = "2024-03-01T16:54:02.651Z"
	Uuid            = "5aa19d2a-4bdf-4687-a850-1804756b3f1f"
	ManyHash        = "d1e60bf3bbbe497448498f942d340b872a89046854827dc43dd703ccbf7a8c78"
	ManifestAddress = "manifest1jjzy5en2000728mzs3wn86a6u6jpygzajj2fg2"
)

type HttpResponder struct {
	Method    string
	Url       string
	Responder func(r *http.Request) (*http.Response, error)
}

var AuthResponder, _ = httpmock.NewJsonResponder(http.StatusOK, map[string]string{"access_token": "ya29.Gl0UBZ3"})
var ClaimedWorkItemResponder, _ = httpmock.NewJsonResponder(http.StatusOK, map[string]any{
	"status":           store.CLAIMED,
	"createdDate":      CreatedDate,
	"uuid":             Uuid,
	"manyHash":         ManyHash,
	"manifestAddress":  ManifestAddress,
	"manifestHash":     nil,
	"manifestDatetime": nil,
	"error":            nil,
})

var TransactionResponseResponder, _ = httpmock.NewJsonResponder(http.StatusOK, map[string]any{
	"argument": map[string]any{
		"from":   "foobar",
		"to":     many.IllegalAddr,
		"amount": 1000, // 1000:1
		"symbol": "dummy",
		"memo":   []string{Uuid, ManifestAddress},
	},
})

var callCount = 0
var MigrationUpdateResponder = func(r *http.Request) (*http.Response, error) {
	callCount++
	if callCount == 1 {
		// Return the first response
		return httpmock.NewJsonResponse(200, map[string]interface{}{
			"status":           store.MIGRATING,
			"createdDate":      CreatedDate,
			"uuid":             Uuid,
			"manyHash":         ManyHash,
			"manifestAddress":  ManifestAddress,
			"manifestHash":     nil,
			"manifestDatetime": nil,
			"error":            nil,
		})
	} else if callCount == 2 {
		var item map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&item)
		if err != nil {
			return httpmock.NewJsonResponse(http.StatusNotFound, nil)
		}
		defer r.Body.Close()

		if item["manifestHash"] == nil {
			return nil, fmt.Errorf("manifestHash is nil")
		}
		if item["manifestDatetime"] == nil {
			return nil, fmt.Errorf("manifestDatetime is nil")
		}

		// Return the second response
		return httpmock.NewJsonResponse(200, map[string]interface{}{
			"status":           store.COMPLETED,
			"createdDate":      CreatedDate,
			"uuid":             Uuid,
			"manyHash":         ManyHash,
			"manifestAddress":  ManifestAddress,
			"manifestHash":     item["manifestHash"],
			"manifestDatetime": item["manifestDatetime"],
			"error":            nil,
		})
	} else {
		// Default response
		return httpmock.NewJsonResponse(http.StatusNotFound, nil)
	}
}
