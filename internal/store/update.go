package store

import (
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"

	"github.com/liftedinit/mfx-migrator/internal/utils"
)

// UpdateWorkItemAndSaveState updates a work item in the remote database and saves the state locally.
func UpdateWorkItemAndSaveState(r *resty.Client, item WorkItem) error {
	// 1. Update the work item
	if err := updateWorkItem(r, item); err != nil {
		return errors.WithMessage(err, "error updating remote work item")
	}

	// 2. Save the work item state
	if err := SaveState(&item); err != nil {
		return err
	}

	return nil
}

// updateWorkItem updates a work item in the remote database.
func updateWorkItem(r *resty.Client, item WorkItem) error {
	// 1. Create an update request
	updateRequest := WorkItemUpdateRequest{
		Status:           item.Status,
		ManifestDatetime: item.ManifestDatetime,
		ManifestHash:     item.ManifestHash,
		Error:            item.Error,
	}

	// 2. Send the update request
	req := r.R().
		SetPathParam("uuid", item.UUID.String()).
		SetBody(&updateRequest).
		SetResult(&WorkItemUpdateResponse{})
	response, err := req.Put("neighborhoods/{neighborhood}/migrations/{uuid}")
	if err != nil {
		return errors.WithMessage(err, "error updating work item")
	}

	if response == nil {
		return fmt.Errorf("no response returned when claiming work item: %s", item.UUID)
	}

	statusCode := response.StatusCode()
	if statusCode != http.StatusOK {
		return fmt.Errorf("response status code: %d", statusCode)
	}

	// 3. Unmarshal the update response
	updateResponse := response.Result().(*WorkItemUpdateResponse)
	if updateResponse == nil {
		return fmt.Errorf("error unmarshalling update response")
	}

	// 4. Validate the work item was updated
	if !(updateResponse.Status == item.Status &&
		utils.EqualTimePtr(updateResponse.ManifestDatetime, item.ManifestDatetime) &&
		utils.EqualStringPtr(updateResponse.ManifestHash, item.ManifestHash) &&
		utils.EqualStringPtr(updateResponse.Error, item.Error)) {
		return fmt.Errorf("work item not updated: %v", item)
	}

	return nil
}
