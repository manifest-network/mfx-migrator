package manifest

import (
	"bufio"
	"context"
	"encoding/hex"
	"fmt"
	"log/slog"
	"math/big"
	"os"
	"strings"
	"time"

	"cosmossdk.io/math"
	"github.com/cometbft/cometbft/rpc/client/http"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/pkg/errors"

	"github.com/liftedinit/mfx-migrator/internal/utils"

	"github.com/liftedinit/mfx-migrator/internal/store"
)

// TODO: Refactor & Cleanup

type MigrationConfig struct {
	ChainID        string
	NodeAddress    string
	KeyringBackend string
	ChainHome      string
	AddressPrefix  string
	BankAddress    string
	TokenMap       map[string]utils.TokenInfo
}

const defaultGasLimit uint64 = 200000

// registerInterfaces registers the necessary interfaces and concrete types on the provided InterfaceRegistry.
func registerInterfaces(registry codectypes.InterfaceRegistry) {
	cryptocodec.RegisterInterfaces(registry)
	authtypes.RegisterInterfaces(registry)
	sdk.RegisterInterfaces(registry)
	banktypes.RegisterInterfaces(registry)
	stakingtypes.RegisterInterfaces(registry)
}

// newClientContext creates and returns a new Cosmos SDK client context.
func newClientContext(chainID, nodeAddress, keyringBackend, chainHomeDir string, inBuf *bufio.Reader) (client.Context, error) {
	registry := codectypes.NewInterfaceRegistry()
	registerInterfaces(registry)
	cdc := codec.NewProtoCodec(registry)

	kr, err := keyring.New(sdk.KeyringServiceName(), keyringBackend, chainHomeDir, inBuf, cdc)
	if err != nil {
		return client.Context{}, errors.WithMessage(err, "failed to create keyring")
	}

	rClient, err := http.New(nodeAddress, "/websocket")
	if err != nil {
		return client.Context{}, errors.WithMessage(err, "failed to create RPC client")
	}

	return client.Context{}.
		WithChainID(chainID).
		WithInterfaceRegistry(registry).
		WithCodec(cdc).
		WithKeyring(kr).
		WithTxConfig(authtx.NewTxConfig(cdc, authtx.DefaultSignModes)).
		WithBroadcastMode("sync").
		WithClient(rClient).
		WithAccountRetriever(authtypes.AccountRetriever{}).
		WithSkipConfirmation(true), nil
}

// Migrate migrates the given amount of tokens to the specified address.
func Migrate(item *store.WorkItem, migrateConfig MigrationConfig, denom string, amount *big.Int) (*sdk.TxResponse, *time.Time, error) {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(migrateConfig.AddressPrefix, migrateConfig.AddressPrefix+"pub")

	inBuf := bufio.NewReader(os.Stdin)
	clientCtx, err := newClientContext(migrateConfig.ChainID, migrateConfig.NodeAddress, migrateConfig.KeyringBackend, migrateConfig.ChainHome, inBuf)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "failed to set up client context")
	}

	addr, info, err := getAccountInfo(clientCtx, migrateConfig.BankAddress)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "failed to get account info")
	}

	manifestAddr, err := sdk.AccAddressFromBech32(item.ManifestAddress)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "failed to parse manifest address")
	}

	msg := banktypes.NewMsgSend(addr, manifestAddr, sdk.NewCoins(sdk.NewCoin(denom, math.NewIntFromBigInt(amount))))
	slog.Debug("Send message", "message", msg)

	txBuilder, err := prepareTx(clientCtx, msg, item.UUID.String(), denom)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "failed to prepare transaction")
	}

	res, blockTime, err := signAndBroadcast(clientCtx, txBuilder, migrateConfig.BankAddress, info)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "failed to sign and broadcast transaction")
	}

	return res, blockTime, nil
}

// getAccountInfo retrieves account information from the keyring.
func getAccountInfo(ctx client.Context, bankAccount string) (sdk.AccAddress, *keyring.Record, error) {
	info, err := ctx.Keyring.Key(bankAccount)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "failed to fetch bank account details")
	}

	addr, err := info.GetAddress()
	if err != nil {
		return nil, nil, errors.WithMessage(err, "failed to get bank address from key")
	}

	if err := ctx.AccountRetriever.EnsureExists(ctx, addr); err != nil {
		return nil, nil, errors.WithMessage(err, "failed to ensure bank account exists")
	}

	return addr, info, nil
}

// prepareTx prepares a transaction builder with the given message.
func prepareTx(ctx client.Context, msg sdk.Msg, memo, denom string) (client.TxBuilder, error) {
	txBuilder := ctx.TxConfig.NewTxBuilder()
	if err := txBuilder.SetMsgs(msg); err != nil {
		return nil, errors.WithMessage(err, "failed to set transaction message")
	}

	txBuilder.SetMemo(memo)
	txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin(denom, math.NewInt(0))))
	txBuilder.SetGasLimit(defaultGasLimit)

	return txBuilder, nil
}

// signAndBroadcast signs and broadcasts the transaction, returning the transaction response and block time.
func signAndBroadcast(ctx client.Context, txBuilder client.TxBuilder, bankAccount string, info *keyring.Record) (*sdk.TxResponse, *time.Time, error) {
	txFactory := tx.Factory{}.
		WithChainID(ctx.ChainID).
		WithKeybase(ctx.Keyring).
		WithGas(0).
		WithGasAdjustment(1.0).
		WithSignMode(signing.SignMode_SIGN_MODE_UNSPECIFIED).
		WithAccountRetriever(ctx.AccountRetriever).
		WithTxConfig(ctx.TxConfig)

	addr, err := info.GetAddress()
	if err != nil {
		return nil, nil, errors.WithMessage(err, "failed to get address from key")
	}
	initNum, initSeq := txFactory.AccountNumber(), txFactory.Sequence()
	if initNum == 0 || initSeq == 0 {
		accNum, seqNum, aErr := ctx.AccountRetriever.GetAccountNumberSequence(ctx, addr)
		if aErr != nil {
			return nil, nil, errors.WithMessage(aErr, "failed to get account number and sequence")
		}

		if initNum == 0 {
			txFactory = txFactory.WithAccountNumber(accNum)
		}

		if initSeq == 0 {
			txFactory = txFactory.WithSequence(seqNum)
		}
	}

	// Sign the transaction
	if tErr := tx.Sign(context.Background(), txFactory, bankAccount, txBuilder, true); tErr != nil {
		return nil, nil, errors.WithMessage(tErr, "failed to sign transaction")
	}

	// Broadcast the transaction
	txBytes, err := ctx.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, nil, errors.WithMessage(err, "failed to encode transaction")
	}

	res, err := ctx.BroadcastTx(txBytes)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "failed to broadcast transaction")
	}

	slog.Info("Transaction broadcasted", "hash", res.TxHash)

	// Wait for the transaction to be included in a block
	txResult, err := waitForTx(ctx.Client, res.TxHash)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "failed to wait for transaction")
	}

	slog.Debug("Transaction result", "tx", txResult.TxResult)
	if txResult.TxResult.Code != 0 {
		return nil, nil, fmt.Errorf("transaction failed: %s", txResult.TxResult.Log)
	}

	slog.Info("Transaction included in block", "height", txResult.Height)

	txBlock, err := ctx.Client.Block(context.Background(), &txResult.Height)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "failed to fetch block")
	}

	blockTime := txBlock.Block.Time.UTC().Truncate(time.Millisecond)

	return res, &blockTime, nil
}

// waitForTx waits for a transaction to be included in a block.
func waitForTx(rClient client.CometRPC, hash string) (*coretypes.ResultTx, error) {
	bHash, err := hex.DecodeString(hash)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to decode hash")
	}

	// Create a context that will be cancelled after the specified timeout
	// TODO: Configure timeout
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			// The context has been cancelled, return an error
			return nil, ctx.Err()
		default:
			r, tErr := rClient.Tx(context.Background(), bHash, false)
			if tErr != nil {
				if strings.Contains(tErr.Error(), "not found") {
					if cErr := waitForNextBlock(rClient); cErr != nil {
						return nil, errors.WithMessage(cErr, "failed to wait for next block")
					}
					continue
				}
				return nil, errors.WithMessage(tErr, "error fetching transaction")
			}
			return r, nil
		}
	}
}

func getLatestBlockHeight(client client.CometRPC) (int64, error) {
	status, err := client.Status(context.Background())
	if err != nil {
		return 0, errors.WithMessage(err, "failed to get blockchain status")
	}
	return status.SyncInfo.LatestBlockHeight, nil
}

func waitForBlockHeight(client client.CometRPC, height int64) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			latestHeight, err := getLatestBlockHeight(client)
			if err != nil {
				return errors.WithMessage(err, "failed to get latest block height")
			}
			if latestHeight >= height {
				return nil
			}
		case <-time.After(30 * time.Second):
			// TODO: Configure timeout
			return fmt.Errorf("timeout exceeded waiting for block")
		}
	}
}

func waitForNextBlock(client client.CometRPC) error {
	latestHeight, err := getLatestBlockHeight(client)
	if err != nil {
		return errors.WithMessage(err, "failed to get latest block height")
	}
	return waitForBlockHeight(client, latestHeight+1)
}
