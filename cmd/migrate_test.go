package cmd_test

import (
	"context"
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
		name string
		args []string
		err  string
		out  string
	}{
		{name: "no argument", args: []string{}, err: "URL cannot be empty"},
		{name: "uuid missing", args: append(slice, urlP...), err: "required flag(s) \"uuid\" not set"},
		{name: "username missing", args: append(slice, uuidP...), err: "username is required"},
		{name: "password missing", args: append(slice, usernameP...), err: "password is required"},
		{name: "chain id missing", args: append(slice, passwordP...), err: "chain ID is required"},
		{name: "address prefix missing", args: append(slice, chainIdP...), err: "address prefix is required"},
		{name: "node address missing", args: append(slice, addressPrefixP...), err: "node address is required"},
		{name: "keyring backend missing", args: append(slice, nodeAddressP...), err: "keyring backend is required"},
		{name: "bank address missing", args: append(slice, keyringBackendP...), err: "bank address is required"},
		{name: "chain home missing", args: append(slice, bankAddressP...), err: "chain home is required"},
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
			_, err := testutils.Execute(t, command, tc.args...)
			require.ErrorContains(t, err, tc.err)
		})
	}
}
