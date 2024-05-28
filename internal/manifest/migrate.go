package manifest

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math/big"
	"os/exec"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/liftedinit/mfx-migrator/internal/config"
	"github.com/pkg/errors"

	"github.com/liftedinit/mfx-migrator/internal/store"
)

const defaultGasLimit uint64 = 200000

// registerInterfaces registers the necessary interfaces and concrete types on the provided InterfaceRegistry.
//func registerInterfaces(registry codectypes.InterfaceRegistry) {
//	cryptocodec.RegisterInterfaces(registry)
//	authtypes.RegisterInterfaces(registry)
//	sdk.RegisterInterfaces(registry)
//	banktypes.RegisterInterfaces(registry)
//	stakingtypes.RegisterInterfaces(registry)
//}

// newClientContext creates and returns a new Cosmos SDK client context.
//func newClientContext(chainID, nodeAddress, keyringBackend, chainHomeDir string, inBuf *bufio.Reader) (client.Context, error) {
//	registry := codectypes.NewInterfaceRegistry()
//	registerInterfaces(registry)
//	cdc := codec.NewProtoCodec(registry)
//
//	kr, err := keyring.New(sdk.KeyringServiceName(), keyringBackend, chainHomeDir, inBuf, cdc)
//	if err != nil {
//		return client.Context{}, errors.WithMessage(err, "failed to create keyring")
//	}
//
//	rClient, err := http.New(nodeAddress, "/websocket")
//	if err != nil {
//		return client.Context{}, errors.WithMessage(err, "failed to create RPC client")
//	}
//
//	return client.Context{}.
//		WithChainID(chainID).
//		WithInterfaceRegistry(registry).
//		WithCodec(cdc).
//		WithKeyring(kr).
//		WithTxConfig(authtx.NewTxConfig(cdc, authtx.DefaultSignModes)).
//		WithBroadcastMode("sync").
//		WithClient(rClient).
//		WithAccountRetriever(authtypes.AccountRetriever{}).
//		WithSkipConfirmation(true), nil
//}

//func runCommand(cmds ...string) {
//	exec.Command("/bin/sh", "-c", cmds...)
//
//}

type CosmosTx struct {
	TxHash string `json:"txhash"`
	Code   int    `json:"code"`
	RawLog string `json:"raw_log"`
}

// Migrate migrates the given amount of tokens to the specified address.
func Migrate(item *store.WorkItem, migrateConfig config.MigrateConfig, denom string, amount *big.Int) (*sdk.TxResponse, *time.Time, error) {
	// Run the command to send the tokens
	cmd := exec.Command("/bin/sh", "-c", migrateConfig.Binary, "tx", "bank", "send",
		migrateConfig.BankAddress,
		item.ManifestAddress, amount.String()+denom,
		"--chain-id", migrateConfig.ChainID,
		"--node", migrateConfig.NodeAddress,
		"--keyring-backend", migrateConfig.KeyringBackend,
		"--home", migrateConfig.ChainHome,
		"--from", migrateConfig.BankAddress,
		"--gas", "auto",
		"--gas-adjustment", "1.3",
		"--output", "json",
		"--yes")
	o, err := cmd.Output()
	if err != nil {
		return nil, nil, errors.WithMessage(err, "failed to execute send command")
	}

	// Unmarshal the transaction response
	var tx CosmosTx
	if err := json.Unmarshal(o, &tx); err != nil {
		return nil, nil, errors.WithMessage(err, "failed to unmarshal transaction")
	}

	// Check if the transaction was successful
	if tx.Code != 0 {
		return nil, nil, errors.Errorf("failed to execute transaction: %s", tx.RawLog)
	}

	// Wait for the transaction to be included in a block
	cmd = exec.Command("/bin/sh", "-c", migrateConfig.Binary, "q", "wait-tx",
		tx.TxHash,
		"--node", migrateConfig.NodeAddress,
		"--home", migrateConfig.ChainHome,
		"--timeout", fmt.Sprintf("%ds", migrateConfig.WaitTxTimeout),
		"--output", "json")
	o, err = cmd.Output()
	if err != nil {
		return nil, nil, errors.WithMessage(err, "failed to wait for transaction")
	}
	slog.Info("Transaction included in block", "output", string(o))

	//
	//c := sdk.GetConfig()
	//c.SetBech32PrefixForAccount(migrateConfig.AddressPrefix, migrateConfig.AddressPrefix+"pub")
	//
	//inBuf := bufio.NewReader(os.Stdin)
	//clientCtx, err := newClientContext(migrateConfig.ChainID, migrateConfig.NodeAddress, migrateConfig.KeyringBackend, migrateConfig.ChainHome, inBuf)
	//if err != nil {
	//	return nil, nil, errors.WithMessage(err, "failed to set up client context")
	//}
	//
	//addr, info, err := getAccountInfo(clientCtx, migrateConfig.BankAddress)
	//if err != nil {
	//	return nil, nil, errors.WithMessage(err, "failed to get account info")
	//}
	//
	//manifestAddr, err := sdk.AccAddressFromBech32(item.ManifestAddress)
	//if err != nil {
	//	return nil, nil, errors.WithMessage(err, "failed to parse manifest address")
	//}
	//
	//msg := banktypes.NewMsgSend(addr, manifestAddr, sdk.NewCoins(sdk.NewCoin(denom, math.NewIntFromBigInt(amount))))
	//slog.Debug("Send message", "message", msg)
	//
	//txBuilder, err := prepareTx(clientCtx, msg, item.UUID.String(), denom)
	//if err != nil {
	//	return nil, nil, errors.WithMessage(err, "failed to prepare transaction")
	//}
	//
	//res, blockTime, err := signAndBroadcast(clientCtx, txBuilder, migrateConfig.BankAddress, info, migrateConfig.WaitTxTimeout, migrateConfig.WaitBlockTimeout)
	//if err != nil {
	//	return nil, nil, errors.WithMessage(err, "failed to sign and broadcast transaction")
	//}
	//
	//return res, blockTime, nil

	return nil, nil, nil
}

// getAccountInfo retrieves account information from the keyring.
//func getAccountInfo(ctx client.Context, bankAccount string) (sdk.AccAddress, *keyring.Record, error) {
//	info, err := ctx.Keyring.Key(bankAccount)
//	if err != nil {
//		return nil, nil, errors.WithMessage(err, "failed to fetch bank account details")
//	}
//
//	addr, err := info.GetAddress()
//	if err != nil {
//		return nil, nil, errors.WithMessage(err, "failed to get bank address from key")
//	}
//
//	if err := ctx.AccountRetriever.EnsureExists(ctx, addr); err != nil {
//		return nil, nil, errors.WithMessage(err, "failed to ensure bank account exists")
//	}
//
//	return addr, info, nil
//}

// prepareTx prepares a transaction builder with the given message.
//func prepareTx(ctx client.Context, msg sdk.Msg, memo, denom string) (client.TxBuilder, error) {
//	txBuilder := ctx.TxConfig.NewTxBuilder()
//	if err := txBuilder.SetMsgs(msg); err != nil {
//		return nil, errors.WithMessage(err, "failed to set transaction message")
//	}
//
//	txBuilder.SetMemo(memo)
//	txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin(denom, math.NewInt(0))))
//	txBuilder.SetGasLimit(defaultGasLimit)
//
//	return txBuilder, nil
//}

// signAndBroadcast signs and broadcasts the transaction, returning the transaction response and block time.
//func signAndBroadcast(ctx client.Context, txBuilder client.TxBuilder, bankAccount string, info *keyring.Record, waitForTxTimeout, blockTimeout uint) (*sdk.TxResponse, *time.Time, error) {
//	txFactory := tx.Factory{}.
//		WithChainID(ctx.ChainID).
//		WithKeybase(ctx.Keyring).
//		WithGas(0).
//		WithGasAdjustment(1.0).
//		WithSignMode(signing.SignMode_SIGN_MODE_UNSPECIFIED).
//		WithAccountRetriever(ctx.AccountRetriever).
//		WithTxConfig(ctx.TxConfig)
//
//	addr, err := info.GetAddress()
//	if err != nil {
//		return nil, nil, errors.WithMessage(err, "failed to get address from key")
//	}
//	initNum, initSeq := txFactory.AccountNumber(), txFactory.Sequence()
//	if initNum == 0 || initSeq == 0 {
//		accNum, seqNum, aErr := ctx.AccountRetriever.GetAccountNumberSequence(ctx, addr)
//		if aErr != nil {
//			return nil, nil, errors.WithMessage(aErr, "failed to get account number and sequence")
//		}
//
//		if initNum == 0 {
//			txFactory = txFactory.WithAccountNumber(accNum)
//		}
//
//		if initSeq == 0 {
//			txFactory = txFactory.WithSequence(seqNum)
//		}
//	}
//
//	// Sign the transaction
//	if tErr := tx.Sign(context.Background(), txFactory, bankAccount, txBuilder, true); tErr != nil {
//		return nil, nil, errors.WithMessage(tErr, "failed to sign transaction")
//	}
//
//	// Broadcast the transaction
//	txBytes, err := ctx.TxConfig.TxEncoder()(txBuilder.GetTx())
//	if err != nil {
//		return nil, nil, errors.WithMessage(err, "failed to encode transaction")
//	}
//
//	res, err := ctx.BroadcastTx(txBytes)
//	if err != nil {
//		return nil, nil, errors.WithMessage(err, "failed to broadcast transaction")
//	}
//
//	slog.Info("Transaction broadcasted", "hash", res.TxHash)
//
//	// Wait for the transaction to be included in a block
//	txResult, err := waitForTx(ctx.Client, res.TxHash, waitForTxTimeout, blockTimeout)
//	if err != nil {
//		return nil, nil, errors.WithMessage(err, "failed to wait for transaction")
//	}
//
//	slog.Debug("Transaction result", "tx", txResult.TxResult)
//	if txResult.TxResult.Code != 0 {
//		return nil, nil, fmt.Errorf("transaction failed: %s", txResult.TxResult.Log)
//	}
//
//	slog.Info("Transaction included in block", "height", txResult.Height)
//
//	txBlock, err := ctx.Client.Block(context.Background(), &txResult.Height)
//	if err != nil {
//		return nil, nil, errors.WithMessage(err, "failed to fetch block")
//	}
//
//	blockTime := txBlock.Block.Time.UTC().Truncate(time.Millisecond)
//
//	return res, &blockTime, nil
//}
//
//// waitForTx waits for a transaction to be included in a block.
//func waitForTx(rClient client.CometRPC, hash string, txTimeout, blockTimeout uint) (*coretypes.ResultTx, error) {
//	bHash, err := hex.DecodeString(hash)
//	if err != nil {
//		return nil, errors.WithMessage(err, "failed to decode hash")
//	}
//
//	// Create a context that will be cancelled after the specified timeout
//	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(txTimeout)*time.Second)
//	defer cancel()
//
//	for {
//		select {
//		case <-ctx.Done():
//			// The context has been cancelled, return an error
//			return nil, ctx.Err()
//		default:
//			r, tErr := rClient.Tx(context.Background(), bHash, false)
//			if tErr != nil {
//				if strings.Contains(tErr.Error(), "not found") {
//					if cErr := waitForNextBlock(rClient, blockTimeout); cErr != nil {
//						return nil, errors.WithMessage(cErr, "failed to wait for next block")
//					}
//					continue
//				}
//				return nil, errors.WithMessage(tErr, "error fetching transaction")
//			}
//			return r, nil
//		}
//	}
//}
//
//func getLatestBlockHeight(client client.CometRPC) (int64, error) {
//	status, err := client.Status(context.Background())
//	if err != nil {
//		return 0, errors.WithMessage(err, "failed to get blockchain status")
//	}
//	return status.SyncInfo.LatestBlockHeight, nil
//}
//
//func waitForBlockHeight(client client.CometRPC, height int64, timeout uint) error {
//	ticker := time.NewTicker(time.Second)
//	defer ticker.Stop()
//
//	for {
//		select {
//		case <-ticker.C:
//			latestHeight, err := getLatestBlockHeight(client)
//			if err != nil {
//				return errors.WithMessage(err, "failed to get latest block height")
//			}
//			if latestHeight >= height {
//				return nil
//			}
//		case <-time.After(time.Duration(timeout) * time.Second):
//			return fmt.Errorf("timeout exceeded waiting for block")
//		}
//	}
//}
//
//func waitForNextBlock(client client.CometRPC, timeout uint) error {
//	latestHeight, err := getLatestBlockHeight(client)
//	if err != nil {
//		return errors.WithMessage(err, "failed to get latest block height")
//	}
//	return waitForBlockHeight(client, latestHeight+1, timeout)
//}
