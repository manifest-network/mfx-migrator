package utils

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/liftedinit/mfx-migrator/internal/common"
)

// MakeHTTPGetRequest performs an HTTP GET request and returns the response body.
// The response body is expected to be JSON.
func MakeHTTPGetRequest(requestURL string) ([]byte, error) {
	slog.Debug("MakeHTTPGetRequest", "requestURL", requestURL)

	res, err := http.Get(requestURL)
	if err != nil {
		slog.Error("error performing http get", "error", err)
		return nil, err
	}

	contentType := res.Header.Get(common.ContentType)
	if !strings.HasPrefix(contentType, common.ContentTypeJSON) {
		slog.Error("unexpected content type", common.ContentType, contentType)
		return nil, fmt.Errorf("unexpected content type: %s", contentType)
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		slog.Error("http response error reading body", "error", err)
		return nil, err
	}

	return body, nil
}

func MakeHTTPPutRequest(requestURL string, dataBuffer *bytes.Buffer) ([]byte, error) {
	slog.Debug("MakeHTTPPutRequest", "requestURL", requestURL)

	req, err := http.NewRequest(http.MethodPut, requestURL, dataBuffer)
	if err != nil {
		slog.Error("error creating http put request", "error", err)
		return nil, err
	}

	req.Header.Set(common.ContentType, common.ContentTypeJSON)

	// Send the PUT request
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		slog.Error("error performing http put", "error", err)
		return nil, err
	}

	contentType := res.Header.Get(common.ContentType)
	if !strings.HasPrefix(contentType, common.ContentTypeJSON) {
		slog.Error("unexpected content type", common.ContentType, contentType)
		return nil, fmt.Errorf("unexpected content type: %s", contentType)
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		slog.Error("http response error reading body", "error", err)
		return nil, err
	}

	return body, nil
}
