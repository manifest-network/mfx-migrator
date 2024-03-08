package interchaintest

import (
	"bufio"
	"context"
	"encoding/json"
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
	"github.com/liftedinit/mfx-migrator/cmd"
	"github.com/liftedinit/mfx-migrator/testutils"
	"github.com/spf13/cobra"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/testreporter"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
)

func TestMigrateOnChain(t *testing.T) {
	tmpdir := interchaintest.TempDir(t)
	if err := os.Chdir(tmpdir); err != nil {
		t.Fatal(err)
	}

	testutils.SetupWorkItem(t)

	ctx := context.Background()
	cfgA := LocalChainConfig
	cfgA.Env = []string{
		fmt.Sprintf("POA_ADMIN_ADDRESS=%s", accAddr),
	}

	logger := zaptest.NewLogger(t, zaptest.Level(zapcore.DebugLevel))
	cf := interchaintest.NewBuiltinChainFactory(logger, []*interchaintest.ChainSpec{
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
		endpoints []HttpResponder
	}{
		{name: "UUUPCANBKH", args: append(slice, chainHomeP...), endpoints: []HttpResponder{
			{Method: "POST", Url: testutils.LoginUrl, Responder: AuthResponder},
			{Method: "GET", Url: "=~^" + testutils.DefaultMigrationUrl, Responder: ClaimedWorkItemResponder},
			{Method: "GET", Url: "=~^" + testutils.DefaultTransactionUrl, Responder: TransactionResponseResponder},
			{Method: "PUT", Url: "=~^" + testutils.DefaultMigrationUrl, Responder: MigrationUpdateResponder},
		}, out: "Migration complete"},
	}

	for _, tc := range tt {
		slice = append(slice, tc.args...)
		t.Run(tc.name, func(t *testing.T) {
			for _, endpoint := range tc.endpoints {
				httpmock.RegisterResponder(endpoint.Method, endpoint.Url, endpoint.Responder)
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

var AuthResponder, _ = httpmock.NewJsonResponder(http.StatusOK, map[string]string{"access_token": "ya29.Gl0UBZ3"})
var ClaimedWorkItemResponder, _ = httpmock.NewJsonResponder(http.StatusOK, map[string]any{
	"status":           2,
	"createdDate":      "2024-03-01T16:54:02.651Z",
	"uuid":             "5aa19d2a-4bdf-4687-a850-1804756b3f1f",
	"manyHash":         "d1e60bf3bbbe497448498f942d340b872a89046854827dc43dd703ccbf7a8c78",
	"manifestAddress":  "manifest1jjzy5en2000728mzs3wn86a6u6jpygzajj2fg2",
	"manifestHash":     nil,
	"manifestDatetime": nil,
	"error":            nil,
})

var TransactionResponseResponder, _ = httpmock.NewJsonResponder(http.StatusOK, map[string]any{
	"argument": map[string]any{
		"from":   "foobar",
		"to":     "maiyg",
		"amount": 101,
		"symbol": "dummy",
		"memo":   []string{"5aa19d2a-4bdf-4687-a850-1804756b3f1f", "manifest1jjzy5en2000728mzs3wn86a6u6jpygzajj2fg2"},
	},
})

var callCount = 0
var MigrationUpdateResponder = func(r *http.Request) (*http.Response, error) {
	callCount++
	if callCount == 1 {
		// Return the first response
		return httpmock.NewJsonResponse(200, map[string]interface{}{
			"status":           3,
			"createdDate":      "2024-03-01T16:54:02.651Z",
			"uuid":             "5aa19d2a-4bdf-4687-a850-1804756b3f1f",
			"manyHash":         "d1e60bf3bbbe497448498f942d340b872a89046854827dc43dd703ccbf7a8c78",
			"manifestAddress":  "manifest1jjzy5en2000728mzs3wn86a6u6jpygzajj2fg2",
			"manifestHash":     nil,
			"manifestDatetime": nil,
			"error":            nil,
		})
	} else if callCount == 2 {
		var item map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&item)
		if err != nil {
			return httpmock.NewJsonResponse(http.StatusNotFound, nil)
		}
		defer r.Body.Close()

		// Return the second response
		return httpmock.NewJsonResponse(200, map[string]interface{}{
			"status":           4,
			"createdDate":      "2024-03-01T16:54:02.651Z",
			"uuid":             "5aa19d2a-4bdf-4687-a850-1804756b3f1f",
			"manyHash":         "d1e60bf3bbbe497448498f942d340b872a89046854827dc43dd703ccbf7a8c78",
			"manifestAddress":  "manifest1jjzy5en2000728mzs3wn86a6u6jpygzajj2fg2",
			"manifestHash":     item["manifestHash"],
			"manifestDatetime": item["manifestDatetime"],
			"error":            nil,
		})
	} else {
		// Default response
		return httpmock.NewJsonResponse(http.StatusNotFound, nil)
	}
}

type HttpResponder struct {
	Method    string
	Url       string
	Responder func(r *http.Request) (*http.Response, error)
}
