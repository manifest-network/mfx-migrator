package cmd

import (
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/liftedinit/mfx-migrator/internal/store"
)

// CreateRestClient creates a new resty client with the parsed URL and the claim config
func CreateRestClient(url string, neighborhood uint64) *resty.Client {
	slog.Info("Creating REST client...")
	// Retry the claim process 3 times with a 5 seconds wait time between retries and a maximum wait time of 60 seconds.
	// Retry uses an exponential backoff algorithm.
	return resty.New().
		SetBaseURL(url).
		SetPathParam("neighborhood", strconv.FormatUint(neighborhood, 10)).
		SetRetryCount(3).
		SetRetryWaitTime(5 * time.Second).SetRetryMaxWaitTime(60 * time.Second)
}

// AuthenticateRestClient logs in to the remote database
func AuthenticateRestClient(r *resty.Client, username, password string) error {
	slog.Info("Authenticating...")
	response, err := r.R().
		SetBody(map[string]interface{}{"username": username, "password": password}).
		SetResult(&store.Token{}).
		Post("/auth/login")
	if err != nil {
		slog.Error("could not login", "error", err)
		return err
	}

	token := response.Result().(*store.Token)
	if token == nil {
		slog.Error("no token returned")
		return errors.New("no token returned")
	}

	if token.AccessToken == "" {
		slog.Error("empty token returned")
		return errors.New("empty token returned")
	}

	slog.Debug("setting auth token", "token", token.AccessToken)
	r.SetAuthToken(token.AccessToken)

	return nil
}
