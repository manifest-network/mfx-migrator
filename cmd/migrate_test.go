package cmd_test

import (
	"context"
	"fmt"
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
		err  error
		out  string
	}{
		{name: "no arg", args: []string{}, err: fmt.Errorf("URL cannot be empty")},
		{name: "U", args: append(slice, urlP...), err: fmt.Errorf("required flag(s) \"uuid\" not set")},
		{name: "UU", args: append(slice, uuidP...), err: fmt.Errorf("username is required")},
		{name: "UUU", args: append(slice, usernameP...), err: fmt.Errorf("password is required")},
		{name: "UUUP", args: append(slice, passwordP...), err: fmt.Errorf("chain ID is required")},
		{name: "UUUPC", args: append(slice, chainIdP...), err: fmt.Errorf("address prefix is required")},
		{name: "UUUPCA", args: append(slice, addressPrefixP...), err: fmt.Errorf("node address is required")},
		{name: "UUUPCAN", args: append(slice, nodeAddressP...), err: fmt.Errorf("keyring backend is required")},
		{name: "UUUPCANB", args: append(slice, keyringBackendP...), err: fmt.Errorf("bank address is required")},
		{name: "UUUPCANBK", args: append(slice, bankAddressP...), err: fmt.Errorf("chain home is required")},
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
			out, err := testutils.Execute(t, command, tc.args...)

			require.ErrorContains(t, err, tc.err.Error())

			if tc.err == nil {
				require.Equal(t, tc.out, out)
			}
		})
	}
}
