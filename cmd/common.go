package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"

	"github.com/liftedinit/mfx-migrator/internal/store"
)

const RestyClientKey store.ContextKey = "restyClient"

// CreateRestClient creates a new resty client with the parsed URL and the claim config
func CreateRestClient(ctx context.Context, url string, neighborhood uint64) *resty.Client {
	slog.Info("Creating REST client...")

	// If a resty client is already in the context, use it. Otherwise, create a new one.
	// This allows the resty client to be injected for testing purposes.
	var client *resty.Client
	if ctxClient := ctx.Value(RestyClientKey); ctxClient != nil {
		client = ctxClient.(*resty.Client)
	} else {
		client = resty.New()
	}
	// Retry the claim process 3 times with a 5 seconds wait time between retries and a maximum wait time of 60 seconds.
	// Retry uses an exponential backoff algorithm.
	return client.
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
		return errors.WithMessage(err, "could not login")
	}

	if response == nil {
		return fmt.Errorf("no response returned when logging in")
	}

	statusCode := response.StatusCode()
	if statusCode != 200 {
		return fmt.Errorf("response status code: %d", statusCode)
	}

	token := response.Result().(*store.Token)
	if token == nil {
		return fmt.Errorf("no token returned")
	}

	if token.AccessToken == "" {
		return fmt.Errorf("empty token returned")
	}

	slog.Debug("setting auth token", "token", token.AccessToken)
	r.SetAuthToken(token.AccessToken)

	return nil
}
