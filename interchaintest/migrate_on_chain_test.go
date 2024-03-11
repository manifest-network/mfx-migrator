package interchaintest

import (
	"context"
	"os"
	"testing"

	"cosmossdk.io/math"
	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/spf13/cobra"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/stretchr/testify/require"

	"github.com/liftedinit/mfx-migrator/cmd"
	"github.com/liftedinit/mfx-migrator/testutils"
)

func TestMigrateOnChain(t *testing.T) {
	ctx := context.Background()
	tmpdir := interchaintest.TempDir(t)
	if err := os.Chdir(tmpdir); err != nil {
		t.Fatal(err)
	}

	testutils.SetupWorkItem(t)

	// Set up the chain and keyring
	appChain, user1 := SetupChain(t, ctx)
	chainConfig := appChain.Config()
	err := SetupKeyring(tmpdir, []ibc.Wallet{user1})
	require.NoError(t, err)

	command := &cobra.Command{Use: "migrate", PersistentPreRunE: cmd.RootCmdPersistentPreRunE, RunE: cmd.MigrateCmdRunE}
	cmd.SetupRootCmdFlags(command)
	cmd.SetupMigrateCmdFlags(command)

	// Create a new resty client and inject it into the command context
	rClient := resty.New()
	cCtx := context.WithValue(context.Background(), cmd.RestyClientKey, rClient)
	command.SetContext(cCtx)

	// Enable http mocking on the resty client
	httpmock.ActivateNonDefault(rClient.GetClient())
	defer httpmock.DeactivateAndReset()

	slice := []string{
		"--url", testutils.RootUrl,
		"--uuid", testutils.Uuid,
		"--username", "user",
		"--password", "pass",
		"--chain-id", chainConfig.ChainID,
		"--address-prefix", chainConfig.Bech32Prefix,
		"--node-address", appChain.GetHostRPCAddress(),
		"--keyring-backend", "test",
		"--bank-address", user1.KeyName(),
		"--chain-home", tmpdir,
	}

	tt := []struct {
		name      string
		args      []string
		err       error
		endpoints []testutils.HttpResponder
	}{
		// Perform a 1000:1 migration (1000 tokens -> 1 umfx)
		{name: "UUUPCANBKH", args: slice, endpoints: []testutils.HttpResponder{
			{Method: "POST", Url: testutils.LoginUrl, Responder: testutils.AuthResponder},
			{Method: "GET", Url: "=~^" + testutils.DefaultMigrationUrl, Responder: testutils.ClaimedWorkItemResponder},
			{Method: "GET", Url: "=~^" + testutils.DefaultTransactionUrl, Responder: testutils.TransactionResponseResponder},
			{Method: "PUT", Url: "=~^" + testutils.DefaultMigrationUrl, Responder: testutils.MigrationUpdateResponder},
		}},
		// TODO: Add more test cases
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// Register the http mock responders
			for _, endpoint := range tc.endpoints {
				httpmock.RegisterResponder(endpoint.Method, endpoint.Url, endpoint.Responder)
			}

			// Check the balance of the bank account pre-migration
			balance1, err := appChain.BankQueryBalance(ctx, user1.FormattedAddress(), Denom)
			require.NoError(t, err)
			require.Equal(t, balance1, DefaultGenesisAmt)

			// Execute the migration
			_, err = testutils.Execute(t, command, tc.args...)
			if tc.err != nil {
				require.ErrorContains(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
			}

			// Check the balance of the bank account post-migration
			balance1, err = appChain.BankQueryBalance(ctx, user1.FormattedAddress(), Denom)
			require.NoError(t, err)
			require.Equal(t, balance1, DefaultGenesisAmt.Sub(math.NewInt(1)))

			// Check the balance of the manifest destination account post-migration
			balance2, err := appChain.BankQueryBalance(ctx, testutils.ManifestAddress, Denom)
			require.NoError(t, err)
			require.Equal(t, balance2, math.NewInt(1))
		})
	}
}
