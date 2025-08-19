package cmd_test

import (
	"context"
	"os"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/manifest-network/mfx-migrator/cmd"
	"github.com/manifest-network/mfx-migrator/testutils"
)

func TestMigrateCmd(t *testing.T) {
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}

	testutils.SetupWorkItem(t)

	var slice []string
	uuidArg := append(slice, []string{"--uuid", testutils.DummyUUIDStr}...)
	urlArg := append(uuidArg, []string{"--url", testutils.RootUrl}...)
	chainHomeArg := append(urlArg, []string{"--chain-home", "/tmp"}...)
	feeGrantArg := append(chainHomeArg, []string{"--fee-granter", "feegranter"}...)
	usernameArg := append(feeGrantArg, []string{"--username", "user"}...)
	passwordArg := append(usernameArg, []string{"--password", "pass"}...)

	pp := make([]string, len(passwordArg))
	copy(pp, passwordArg)
	chainIdArg := append(pp, []string{"--chain-id", ""}...)
	addressPrefixArg := append(pp, []string{"--address-prefix", ""}...)
	nodeAddressArg := append(pp, []string{"--node-address", ""}...)
	keyringBackendArg := append(pp, []string{"--keyring-backend", ""}...)
	bankAddressArg := append(pp, []string{"--bank-address", ""}...)

	tt := []struct {
		name string
		args []string
		err  string
		out  string
	}{
		{name: "no argument", args: []string{}, err: "required flag(s) \"uuid\" not set"},
		{name: "url missing", args: uuidArg, err: ""},
		{name: "chain home missing", args: urlArg, err: "chain home is required"},
		{name: "feegrant missing", args: chainHomeArg, err: "fee granter is required"},
		{name: "username missing", args: feeGrantArg, err: "username is required"},
		{name: "password missing", args: usernameArg, err: "password is required"},

		// Set defaults to empty strings
		{name: "chain id missing", args: chainIdArg, err: "chain ID is required"},
		{name: "address prefix missing", args: addressPrefixArg, err: "address prefix is required"},
		{name: "node address missing", args: nodeAddressArg, err: "node address is required"},
		{name: "keyring backend missing", args: keyringBackendArg, err: "keyring backend is required"},
		{name: "bank address missing", args: bankAddressArg, err: "bank address is required"},
	}

	for _, tc := range tt {
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

		t.Run(tc.name, func(t *testing.T) {
			_, err := testutils.Execute(t, command, tc.args...)
			require.ErrorContains(t, err, tc.err)
		})
	}
}
