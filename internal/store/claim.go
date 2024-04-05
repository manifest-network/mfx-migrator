package store

import (
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// ClaimWorkItemFromQueue retrieves a work item from the remote database work queue.
func ClaimWorkItemFromQueue(r *resty.Client) ([]*WorkItem, error) {
	// 1. Claim work items
	items, err := claimWorkItems(r)
	if err != nil {
		return nil, errors.WithMessage(err, "error claiming work items")
	}

	// 2. Save the work item states
	for _, item := range items {
		if err := SaveState(item); err != nil {
			return nil, err
		}
	}

	return items, nil
}

func ClaimWorkItemFromUUID(r *resty.Client, uuid uuid.UUID, force bool) (*WorkItem, error) {
	item, err := claimWorkItem(r, uuid, force)
	if err != nil {
		return nil, errors.WithMessage(err, "error claiming work item")
	}

	if err := SaveState(item); err != nil {
		return nil, err
	}

	return item, nil
}

func claimWorkItems(r *resty.Client) ([]*WorkItem, error) {
	req := r.R().SetResult(&[]*WorkItem{})
	response, err := req.Put("neighborhoods/{neighborhood}/migrations/claim/")
	if err != nil {
		return nil, errors.WithMessage(err, "error claiming work items")
	}

	if response == nil {
		return nil, fmt.Errorf("no response returned when claiming work items")
	}

	statusCode := response.StatusCode()
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("response status code: %d", statusCode)
	}

	claimResponse := response.Result().(*[]*WorkItem)
	if claimResponse == nil {
		return nil, fmt.Errorf("error unmarshalling claim response")
	}

	return *claimResponse, nil
}

func claimWorkItem(r *resty.Client, itemUUID uuid.UUID, force bool) (*WorkItem, error) {
	req := r.R().SetResult(&WorkItem{}).
		SetPathParam("uuid", itemUUID.String()).
		SetQueryParam("force", fmt.Sprintf("%t", force))
	response, err := req.Put("neighborhoods/{neighborhood}/migrations/claim/{uuid}")
	if err != nil {
		return nil, errors.WithMessage(err, "error claiming work item")
	}

	if response == nil {
		return nil, fmt.Errorf("no response returned when claiming work item: %s", itemUUID)
	}

	statusCode := response.StatusCode()
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("response status code: %d", statusCode)
	}

	item := response.Result().(*WorkItem)
	if item == nil {
		return nil, fmt.Errorf("error unmarshalling claim response")
	}

	return item, nil
}
