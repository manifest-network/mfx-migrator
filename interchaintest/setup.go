package interchaintest

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/liftedinit/manifest-ledger/interchaintest/helpers"
	tokenfactorytypes "github.com/reecepbcups/tokenfactory/x/tokenfactory/types"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/testreporter"
	poatypes "github.com/strangelove-ventures/poa"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"

	sdktestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
)

// AppEncoding returns the encoding for the test
func AppEncoding() *sdktestutil.TestEncodingConfig {
	enc := cosmos.DefaultEncoding()

	tokenfactorytypes.RegisterInterfaces(enc.InterfaceRegistry)
	poatypes.RegisterInterfaces(enc.InterfaceRegistry)

	return &enc
}

// setupConfig sets up the chain configuration for the test
func setupConfig() ibc.ChainConfig {
	cfgA := LocalChainConfig
	cfgA.Env = []string{
		fmt.Sprintf("POA_ADMIN_ADDRESS=%s", accAddr),
	}

	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(cfgA.Bech32Prefix, cfgA.Bech32Prefix+"pub")

	return cfgA
}

// setupChain sets up the chain for the test
func setupChain(t *testing.T, config ibc.ChainConfig) ([]ibc.Chain, error) {
	logger := zaptest.NewLogger(t, zaptest.Level(zapcore.DebugLevel))
	cf := interchaintest.NewBuiltinChainFactory(logger, []*interchaintest.ChainSpec{
		{
			Name:          config.Name,
			Version:       config.Images[0].Version,
			ChainName:     config.ChainID,
			NumValidators: &vals,
			NumFullNodes:  &fullNodes,
			ChainConfig:   config,
		},
	})

	chains, err := cf.Chains(t.Name())
	if err != nil {
		return nil, err
	}

	return chains, nil
}

// setupInterchain sets up the interchain for the test
// We only need one chain for this test but we could link multiple chains together
func setupInterchain(t *testing.T, ctx context.Context, manifestA *cosmos.CosmosChain) {
	// Relayer Factory
	client, network := interchaintest.DockerSetup(t)

	ic := interchaintest.NewInterchain().AddChain(manifestA)

	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)

	// Build interchain
	require.NoError(t, ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:         t.Name(),
		Client:           client,
		NetworkID:        network,
		SkipPathCreation: false,
	}))

	t.Cleanup(func() {
		_ = ic.Close()
	})
}

// setupUser sets up a user for the test
func setupUser(ctx context.Context, manifestA *cosmos.CosmosChain) (ibc.Wallet, error) {
	user1, err := interchaintest.GetAndFundTestUserWithMnemonic(ctx, "default", userMnemonic, DefaultGenesisAmt, manifestA)
	if err != nil {
		return nil, err
	}
	return user1, nil
}

// checkGenesisDefault checks the default genesis parameters on chain
func checkGenesisDefault(t *testing.T, node *cosmos.ChainNode) {
	p, err := helpers.ManifestQueryParams(context.Background(), node)
	require.NoError(t, err)
	fmt.Println(p)
	require.True(t, p.Inflation.AutomaticEnabled)
	require.EqualValues(t, p.Inflation.MintDenom, Denom)
}

// SetupChain sets up an isolated chain for the test
func SetupChain(t *testing.T, ctx context.Context) (*cosmos.CosmosChain, ibc.Wallet) {
	cfgA := setupConfig()
	chains, err := setupChain(t, cfgA)
	require.NoError(t, err)

	manifestA := chains[0].(*cosmos.CosmosChain)
	setupInterchain(t, ctx, manifestA)

	user1, err := setupUser(ctx, manifestA)
	require.NoError(t, err)

	node := manifestA.GetNode()

	checkGenesisDefault(t, node)

	return manifestA, user1
}

// SetupKeyring sets up the keyring for the test with the given users
func SetupKeyring(tmpdir string, users []ibc.Wallet) error {
	cdc := AppEncoding()
	inBuf := bufio.NewReader(os.Stdin)

	kr, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendTest, tmpdir, inBuf, cdc.Codec)
	if err != nil {
		return err
	}

	// Set the coin type as per interchain
	coinType, err := strconv.ParseUint("118", 10, 32)

	for _, user := range users {
		_, err = kr.NewAccount(user.KeyName(), userMnemonic, "", hd.CreateHDPath(uint32(coinType), 0, 0).String(), hd.Secp256k1)
		if err != nil {
			return err
		}
	}

	return nil
}
