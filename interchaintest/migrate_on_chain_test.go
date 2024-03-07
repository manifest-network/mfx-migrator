package interchaintest

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/liftedinit/manifest-ledger/interchaintest/helpers"
	"github.com/spf13/cobra"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/testreporter"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"

	"github.com/liftedinit/mfx-migrator/cmd"
	"github.com/liftedinit/mfx-migrator/testutils"
)

func TestMigrateOnChain(t *testing.T) {
	tmpdir := t.TempDir()
	if err := os.Chdir(tmpdir); err != nil {
		t.Fatal(err)
	}

	testutils.SetupWorkItem(t)

	ctx := context.Background()
	cfgA := LocalChainConfig
	cfgA.Env = []string{
		fmt.Sprintf("POA_ADMIN_ADDRESS=%s", accAddr),
	}

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t, zaptest.Level(zapcore.DebugLevel)), []*interchaintest.ChainSpec{
		{
			Name:          "manifest-ledger",
			Version:       "v0.0.1-alpha",
			ChainName:     cfgA.ChainID,
			NumValidators: &vals,
			NumFullNodes:  &fullNodes,
			ChainConfig:   cfgA,
		},
	})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)
	manifestA := chains[0].(*cosmos.CosmosChain)

	// Relayer Factory
	client, network := interchaintest.DockerSetup(t)

	ic := interchaintest.NewInterchain().
		AddChain(manifestA)

	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)

	// Build interchain
	require.NoError(t, ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:         t.Name(),
		Client:           client,
		NetworkID:        network,
		SkipPathCreation: false,
	}))

	// Chains
	appChain := chains[0].(*cosmos.CosmosChain)

	//poaAdmin, err := interchaintest.GetAndFundTestUserWithMnemonic(ctx, "acc0", accMnemonic, DefaultGenesisAmt, appChain)
	//if err != nil {
	//	t.Fatal(err)
	//}

	//users := interchaintest.GetAndFundTestUsers(t, ctx, "default", DefaultGenesisAmt, appChain, appChain, appChain)
	//user1 := users[0]
	//user1, user2 := users[0], users[1]
	//uaddr, addr2 := user1.FormattedAddress(), user2.FormattedAddress()

	user1, err := interchaintest.GetAndFundTestUserWithMnemonic(ctx, "default", userMnemonic, DefaultGenesisAmt, appChain)
	t.Log("USER 1 ADDR", user1.FormattedAddress())
	require.NoError(t, err)

	node := appChain.GetNode()

	// Base Query Check of genesis defaults
	p, err := helpers.ManifestQueryParams(ctx, node)
	require.NoError(t, err)
	fmt.Println(p)
	require.True(t, p.Inflation.AutomaticEnabled)
	require.EqualValues(t, p.Inflation.MintDenom, Denom)
	//inflationAddr := p.StakeHolders[0].Address

	command := &cobra.Command{Use: "migrate", PersistentPreRunE: cmd.RootCmdPersistentPreRunE, RunE: cmd.MigrateCmdRunE}

	// Create a new resty client and inject it into the command context
	rClient := resty.New()
	cCtx := context.WithValue(context.Background(), cmd.RestyClientKey, rClient)
	command.SetContext(cCtx)

	// Enable http mocking on the resty client
	httpmock.ActivateNonDefault(rClient.GetClient())
	defer httpmock.DeactivateAndReset()

	cmd.SetupRootCmdFlags(command)
	cmd.SetupMigrateCmdFlags(command)

	registry := codectypes.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(registry)
	authtypes.RegisterInterfaces(registry)
	sdk.RegisterInterfaces(registry)
	banktypes.RegisterInterfaces(registry)
	stakingtypes.RegisterInterfaces(registry)
	cdc := codec.NewProtoCodec(registry)
	inBuf := bufio.NewReader(os.Stdin)

	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("manifest", "manifest"+"pub")

	kr, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendTest, tmpdir, inBuf, cdc)
	require.NoError(t, err)

	coinType, err := strconv.ParseUint("118", 10, 32)
	_, err = kr.NewAccount(user1.KeyName(), userMnemonic, "", hd.CreateHDPath(uint32(coinType), 0, 0).String(), hd.Secp256k1)
	require.NoError(t, err)

	urlP := []string{"--url", testutils.RootUrl}
	uuidP := []string{"--uuid", "5aa19d2a-4bdf-4687-a850-1804756b3f1f"}
	usernameP := []string{"--username", "user"}
	passwordP := []string{"--password", "pass"}
	chainIdP := []string{"--chain-id", "manifest-2"}
	addressPrefixP := []string{"--address-prefix", "manifest"}
	nodeAddressP := []string{"--node-address", appChain.GetHostRPCAddress()}
	keyringBackendP := []string{"--keyring-backend", "test"}
	bankAddressP := []string{"--bank-address", user1.KeyName()}
	chainHomeP := []string{"--chain-home", tmpdir}
	logLevelP := []string{"-l", "info"}

	var slice []string
	slice = append(slice, urlP...)
	slice = append(slice, uuidP...)
	slice = append(slice, usernameP...)
	slice = append(slice, passwordP...)
	slice = append(slice, chainIdP...)
	slice = append(slice, addressPrefixP...)
	slice = append(slice, nodeAddressP...)
	slice = append(slice, keyringBackendP...)
	slice = append(slice, bankAddressP...)
	slice = append(slice, logLevelP...)

	tt := []struct {
		name      string
		args      []string
		err       error
		out       string
		endpoints []testutils.Endpoint
	}{
		// TODO: Modify how responders are registered. I need to different responses for the same URL in this test.
		{name: "UUUPCANBKH", args: append(slice, chainHomeP...), endpoints: []testutils.Endpoint{
			{Method: "POST", Url: testutils.LoginUrl, Data: "testdata/auth-token.json", Code: http.StatusOK},
			{Method: "GET", Url: "=~^" + testutils.DefaultMigrationUrl, Data: "testdata/claimed-work-item.json", Code: http.StatusOK},
			{Method: "GET", Url: "=~^" + testutils.DefaultTransactionUrl, Data: "testdata/many-tx.json", Code: http.StatusOK},
			{Method: "PUT", Url: "=~^" + testutils.DefaultMigrationUrl, Data: "testdata/work-item-update-migrating.json", Code: http.StatusOK},
		}, out: "Migration complete"},
	}

	for _, tc := range tt {
		slice = append(slice, tc.args...)
		t.Run(tc.name, func(t *testing.T) {
			for _, endpoint := range tc.endpoints {
				testutils.SetupMockResponder(t, endpoint.Method, endpoint.Url, endpoint.Data, endpoint.Code)
			}

			out, err := testutils.Execute(t, command, tc.args...)

			if tc.err == nil {
				require.Contains(t, out, tc.out)
			} else {
				require.ErrorContains(t, err, tc.err.Error())
			}

		})
	}

	t.Cleanup(func() {
		_ = ic.Close()
	})
}
