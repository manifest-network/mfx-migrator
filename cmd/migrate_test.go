package cmd_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/liftedinit/mfx-migrator/cmd"
	"github.com/liftedinit/mfx-migrator/testutils"
)

func TestMigrateCmd(t *testing.T) {
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}

	testutils.SetupWorkItem(t)

	urlP := []string{"--url", testutils.RootUrl}
	uuidP := []string{"--uuid", testutils.DummyUUIDStr}
	usernameP := []string{"--username", "user"}
	passwordP := []string{"--password", "pass"}
	chainIdP := []string{"--chain-id", "my-chain"}
	addressPrefixP := []string{"--address-prefix", "manifest"}
	nodeAddressP := []string{"--node-address", "http://localhost:26657"}
	keyringBackendP := []string{"--keyring-backend", "test"}
	bankAddressP := []string{"--bank-address", "alice"}

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
