package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/liftedinit/mfx-migrator/internal/config"
	"github.com/liftedinit/mfx-migrator/internal/utils"

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
	return client.
		SetBaseURL(url).
		SetPathParam("neighborhood", strconv.FormatUint(neighborhood, 10)).
		SetRetryCount(3).                      // Retry the request process 3 times. Retry uses an exponential backoff algorithm.
		SetRetryWaitTime(5 * time.Second).     // With a 5 seconds wait time between retries
		SetRetryMaxWaitTime(60 * time.Second). // And a maximum wait time of 60 seconds for the whole process
		SetTimeout(10 * time.Second)           // Set a timeout of 10 seconds for the request
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

// LoadConfigFromCLI loads the Config from the CLI flags
//
// `uuidKey` is the name of the viper key that holds the UUID
// This is necessary because the UUID is used in multiple commands
func LoadConfigFromCLI(uuidKey string) config.Config {
	return config.Config{
		UUID:         viper.GetString(uuidKey),
		Url:          viper.GetString("url"),
		Neighborhood: viper.GetUint64("neighborhood"),
	}
}

func LoadAuthConfigFromCLI() config.AuthConfig {
	return config.AuthConfig{
		Username: viper.GetString("username"),
		Password: viper.GetString("password"),
	}
}

func LoadClaimConfigFromCLI() config.ClaimConfig {
	return config.ClaimConfig{
		Force: viper.GetBool("force"),
	}
}

func LoadMigrationConfigFromCLI() config.MigrateConfig {
	var tokenMap map[string]utils.TokenInfo
	if err := viper.UnmarshalKey("token-map", &tokenMap); err != nil {
		panic(err)
	}
	return config.MigrateConfig{
		ChainID:          viper.GetString("chain-id"),
		AddressPrefix:    viper.GetString("address-prefix"),
		NodeAddress:      viper.GetString("node-address"),
		KeyringBackend:   viper.GetString("keyring-backend"),
		BankAddress:      viper.GetString("bank-address"),
		ChainHome:        viper.GetString("chain-home"),
		TokenMap:         tokenMap,
		WaitTxTimeout:    viper.GetUint("wait-for-tx-timeout"),
		WaitBlockTimeout: viper.GetUint("wait-for-block-timeout"),
		Binary:           viper.GetString("binary"),
		GasAdjustment:    viper.GetFloat64("gas-adjustment"),
		GasPrice:         viper.GetFloat64("gas-price"),
		GasDenom:         viper.GetString("gas-denom"),
		FeeGranter:       viper.GetString("fee-granter"),
	}
}
