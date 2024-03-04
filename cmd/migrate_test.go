package cmd_test

import (
	"context"
	"errors"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/jarcoal/httpmock"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/liftedinit/mfx-migrator/cmd"
	"github.com/liftedinit/mfx-migrator/internal/store"
	"github.com/liftedinit/mfx-migrator/internal/utils"
	"github.com/liftedinit/mfx-migrator/testutils"
)

const (
	dummyUUIDStr      = "5aa19d2a-4bdf-4687-a850-1804756b3f1f"
	dummyHash         = "d1e60bf3bbbe497448498f942d340b872a89046854827dc43dd703ccbf7a8c78"
	dummyManifestAddr = "manifest1jjzy5en2000728mzs3wn86a6u6jpygzajj2fg2"
	dummyCreatedDate  = "2024-03-01T16:54:02.651Z"
)

func setupWorkItem(t *testing.T) {
	dummyUUID := uuid.MustParse(dummyUUIDStr)
	parsedCreatedDate, err := time.Parse(time.RFC3339, dummyCreatedDate)
	if err != nil {
		t.Fatal(err)
	}

	viper.Set("token-map", map[string]utils.TokenInfo{
		"dummy": {Denom: "dummy", Precision: 6},
	})

	// Some item
	item := store.WorkItem{
		Status:           2,
		CreatedDate:      &parsedCreatedDate,
		UUID:             dummyUUID,
		ManyHash:         dummyHash,
		ManifestAddress:  dummyManifestAddr,
		ManifestHash:     nil,
		ManifestDatetime: nil,
		Error:            nil,
	}

	if err := store.SaveState(&item); err != nil {
		t.Fatal(err)
	}
}

func TestMigrateCmd(t *testing.T) {
	tmpdir := testutils.SetupTmpDir(t)
	defer os.RemoveAll(tmpdir)

	setupWorkItem(t)
	err := testutils.CopyDirFromEmbedFS(testutils.MockKeyring, "keyring-test", tmpdir)
	require.NoError(t, err)

	urlP := []string{"--url", testutils.RootUrl}
	uuidP := []string{"--uuid", dummyUUIDStr}
	usernameP := []string{"--username", "user"}
	passwordP := []string{"--password", "pass"}
	chainIdP := []string{"--chain-id", "my-chain"}
	addressPrefixP := []string{"--address-prefix", "manifest"}
	nodeAddressP := []string{"--node-address", "http://localhost:26657"}
	keyringBackendP := []string{"--keyring-backend", "test"}
	bankAddressP := []string{"--bank-address", "alice"}
	chainHomeP := []string{"--chain-home", tmpdir}

	var slice []string

	tt := []struct {
		name      string
		args      []string
		err       error
		out       string
		endpoints []testutils.Endpoint
	}{
		{name: "no arg", args: []string{}, err: errors.New("URL cannot be empty")},
		{name: "U", args: append(slice, urlP...), err: errors.New("required flag(s) \"uuid\" not set")},
		{name: "UU", args: append(slice, uuidP...), err: errors.New("username is required")},
		{name: "UUU", args: append(slice, usernameP...), err: errors.New("password is required")},
		{name: "UUUP", args: append(slice, passwordP...), err: errors.New("chain ID is required")},
		{name: "UUUPC", args: append(slice, chainIdP...), err: errors.New("address prefix is required")},
		{name: "UUUPCA", args: append(slice, addressPrefixP...), err: errors.New("node address is required")},
		{name: "UUUPCAN", args: append(slice, nodeAddressP...), err: errors.New("keyring backend is required")},
		{name: "UUUPCANB", args: append(slice, keyringBackendP...), err: errors.New("bank address is required")},
		{name: "UUUPCANBK", args: append(slice, bankAddressP...), err: errors.New("chain home is required")},
		// TODO: Fix the following test case
		// The test case fails because I need to mock the blockchain
		{name: "UUUPCANBKH", args: append(slice, chainHomeP...), endpoints: []testutils.Endpoint{
			{Method: "POST", Url: testutils.LoginUrl, Data: "testdata/auth-token.json", Code: http.StatusOK},
			{Method: "GET", Url: "=~^" + testutils.DefaultMigrationUrl, Data: "testdata/claimed-work-item.json", Code: http.StatusOK},
			{Method: "GET", Url: "=~^" + testutils.DefaultTransactionUrl, Data: "testdata/many-tx.json", Code: http.StatusOK},
			{Method: "PUT", Url: "=~^" + testutils.DefaultMigrationUrl, Data: "testdata/work-item-update-migrating.json", Code: http.StatusOK},
		}, err: errors.New("work item not updated")},
	}

	command := &cobra.Command{Use: "migrate", PersistentPreRunE: cmd.RootCmdPersistentPreRunE, RunE: cmd.MigrateCmdRunE}

	// Create a new resty client and inject it into the command context
	client := resty.New()
	ctx := context.WithValue(context.Background(), cmd.RestyClientKey, client)
	command.SetContext(ctx)

	// Enable http mocking on the resty client
	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()

	cmd.SetupRootCmdFlags(command)
	cmd.SetupMigrateCmdFlags(command)

	for _, tc := range tt {
		slice = append(slice, tc.args...)
		t.Run(tc.name, func(t *testing.T) {
			for _, endpoint := range tc.endpoints {
				testutils.SetupMockResponder(t, endpoint.Method, endpoint.Url, endpoint.Data, endpoint.Code)
			}

			out, err := testutils.Execute(t, command, tc.args...)

			require.ErrorContains(t, err, tc.err.Error())

			if tc.err == nil {
				require.Equal(t, tc.out, out)
			}
		})
	}
}
