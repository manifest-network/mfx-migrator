package manifest

import (
	"bufio"
	"context"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

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

	"github.com/liftedinit/mfx-migrator/internal/utils"

	"github.com/liftedinit/mfx-migrator/internal/store"
)

// TODO: Tests
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
		slog.Error("Failed to create keyring", "error", err)
		return client.Context{}, fmt.Errorf("failed to create keyring: %w", err)
	}

	rClient, err := http.New(nodeAddress, "/websocket")
	if err != nil {
		slog.Error("Failed to create RPC client", "error", err)
		return client.Context{}, fmt.Errorf("failed to create RPC client: %w", err)
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
func Migrate(item *store.WorkItem, migrateConfig MigrationConfig, denom string, amount int64) (*sdk.TxResponse, *time.Time, error) {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(migrateConfig.AddressPrefix, migrateConfig.AddressPrefix+"pub")

	inBuf := bufio.NewReader(os.Stdin)
	clientCtx, err := newClientContext(migrateConfig.ChainID, migrateConfig.NodeAddress, migrateConfig.KeyringBackend, migrateConfig.ChainHome, inBuf)
	if err != nil {
		slog.Error("Failed to set up client context", "error", err)
		return nil, nil, fmt.Errorf("failed to set up client context: %w", err)
	}

	addr, info, err := getAccountInfo(clientCtx, migrateConfig.BankAddress)
	if err != nil {
		slog.Error("Failed to get account info", "error", err)
		return nil, nil, err
	}

	manifestAddr, err := sdk.AccAddressFromBech32(item.ManifestAddress)
	if err != nil {
		slog.Error("Failed to parse manifest address", "error", err)
		return nil, nil, fmt.Errorf("failed to parse manifest address: %w", err)
	}

	msg := banktypes.NewMsgSend(addr, manifestAddr, sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(amount))))
	txBuilder, err := prepareTx(clientCtx, msg, item.UUID.String(), denom)
	if err != nil {
		slog.Error("Failed to prepare transaction", "error", err)
		return nil, nil, err
	}

	res, blockTime, err := signAndBroadcast(clientCtx, txBuilder, migrateConfig.BankAddress, info)
	if err != nil {
		slog.Error("Failed to sign and broadcast transaction", "error", err)
		return nil, nil, err
	}

	return res, blockTime, nil
}

// getAccountInfo retrieves account information from the keyring.
func getAccountInfo(ctx client.Context, bankAccount string) (sdk.AccAddress, *keyring.Record, error) {
	info, err := ctx.Keyring.Key(bankAccount)
	if err != nil {
		slog.Error("Failed to fetch bank account details", "error", err)
		return nil, nil, fmt.Errorf("failed to fetch account details: %w", err)
	}

	addr, err := info.GetAddress()
	if err != nil {
		slog.Error("Failed to get bank address from key", "error", err)
		return nil, nil, fmt.Errorf("failed to get address from key: %w", err)
	}

	if err := ctx.AccountRetriever.EnsureExists(ctx, addr); err != nil {
		slog.Error("Failed to ensure bank account exists", "error", err)
		return nil, nil, fmt.Errorf("failed to ensure account exists: %w", err)
	}

	return addr, info, nil
}

// prepareTx prepares a transaction builder with the given message.
func prepareTx(ctx client.Context, msg sdk.Msg, memo, denom string) (client.TxBuilder, error) {
	txBuilder := ctx.TxConfig.NewTxBuilder()
	if err := txBuilder.SetMsgs(msg); err != nil {
		slog.Error("Failed to set message", "error", err)
		return nil, fmt.Errorf("failed to set message: %w", err)
	}

	txBuilder.SetMemo(memo)
	txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(1))))
	txBuilder.SetGasLimit(defaultGasLimit)

	return txBuilder, nil
}

// signAndBroadcast signs and broadcasts the transaction, returning the transaction response and block time.
func signAndBroadcast(ctx client.Context, txBuilder client.TxBuilder, bankAccount string, info *keyring.Record) (*sdk.TxResponse, *time.Time, error) {
	txFactory := tx.Factory{}.
		WithChainID(ctx.ChainID).
		WithKeybase(ctx.Keyring).
		WithGas(300000).
		WithGasAdjustment(1.0).
		WithSignMode(signing.SignMode_SIGN_MODE_UNSPECIFIED).
		WithAccountRetriever(ctx.AccountRetriever).
		WithTxConfig(ctx.TxConfig)

	addr, err := info.GetAddress()
	if err != nil {
		slog.Error("Failed to get address from key", "error", err)
		return nil, nil, fmt.Errorf("failed to get address: %w", err)
	}
	initNum, initSeq := txFactory.AccountNumber(), txFactory.Sequence()
	if initNum == 0 || initSeq == 0 {
		accNum, seqNum, aErr := ctx.AccountRetriever.GetAccountNumberSequence(ctx, addr)
		if aErr != nil {
			slog.Error("Error getting account number sequence", "error", aErr)
			return nil, nil, aErr
		}

		if initNum == 0 {
			txFactory = txFactory.WithAccountNumber(accNum)
		}

		if initSeq == 0 {
			txFactory = txFactory.WithSequence(seqNum)
		}
	}

	// Sign the transaction
	if tErr := tx.Sign(txFactory, bankAccount, txBuilder, true); tErr != nil {
		slog.Error("Failed to sign transaction", "error", tErr)
		return nil, nil, fmt.Errorf("failed to sign transaction: %w", tErr)
	}

	// Broadcast the transaction
	txBytes, err := ctx.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		slog.Error("Failed to encode transaction", "error", err)
		return nil, nil, fmt.Errorf("failed to encode transaction: %w", err)
	}

	res, err := ctx.BroadcastTx(txBytes)
	if err != nil {
		slog.Error("Failed to broadcast transaction", "error", err)
		return nil, nil, fmt.Errorf("failed to broadcast transaction: %w", err)
	}

	slog.Info("Transaction broadcasted", "hash", res.TxHash)

	// Wait for the transaction to be included in a block
	txResult, err := waitForTx(ctx.Client, res.TxHash)
	if err != nil {
		slog.Error("Failed to wait for transaction", "error", err)
		return nil, nil, err
	}

	slog.Info("Transaction included in block", "height", txResult.Height)

	txBlock, err := ctx.Client.Block(context.Background(), &txResult.Height)
	if err != nil {
		slog.Error("Failed to fetch block", "error", err)
		return nil, nil, fmt.Errorf("failed to fetch block: %w", err)
	}

	blockTime := txBlock.Block.Time.UTC().Truncate(time.Millisecond)

	return res, &blockTime, nil
}

// waitForTx waits for a transaction to be included in a block.
func waitForTx(rClient client.TendermintRPC, hash string) (*coretypes.ResultTx, error) {
	bHash, err := hex.DecodeString(hash)
	if err != nil {
		slog.Error("Failed to decode hash", "error", err)
		return nil, fmt.Errorf("failed to decode hash: %w", err)
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
						slog.Error("error waiting for next block", "error", cErr)
						return nil, cErr
					}
					continue
				}
				slog.Error("Failed to fetch transaction", "error", tErr)
				return nil, fmt.Errorf("error fetching transaction: %w", tErr)
			}
			return r, nil
		}
	}
}

func getLatestBlockHeight(client client.TendermintRPC) (int64, error) {
	status, err := client.Status(context.Background())
	if err != nil {
		slog.Error("Failed to get blockchain status", "error", err)
		return 0, err
	}
	return status.SyncInfo.LatestBlockHeight, nil
}

func waitForBlockHeight(client client.TendermintRPC, height int64) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			latestHeight, err := getLatestBlockHeight(client)
			if err != nil {
				slog.Error("Failed to get latest block height", "error", err)
				return err
			}
			if latestHeight >= height {
				return nil
			}
		case <-time.After(30 * time.Second):
			// TODO: Configure timeout
			slog.Error("Timeout exceeded waiting for block")
			return fmt.Errorf("timeout exceeded waiting for block")
		}
	}
}

func waitForNextBlock(client client.TendermintRPC) error {
	latestHeight, err := getLatestBlockHeight(client)
	if err != nil {
		slog.Error("Failed to get latest block height", "error", err)
		return err
	}
	return waitForBlockHeight(client, latestHeight+1)
}
