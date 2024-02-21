package utils

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/liftedinit/mfx-migrator/internal/common"
)

// MakeHTTPRequest performs an HTTP GET request and returns the response body.
// The response body is expected to be JSON.
func MakeHTTPRequest(requestURL string) ([]byte, error) {
	slog.Debug("MakeHTTPRequest", "requestURL", requestURL)

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
