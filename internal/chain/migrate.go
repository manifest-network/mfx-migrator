package chain

import (
	"bufio"
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
	"time"

	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
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
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/liftedinit/mfx-migrator/internal/store"
)

const (
	ChainId        = "ghostcloud"                 // TODO: change this
	AddressPrefix  = "gc"                         // TODO: change this
	BankAccount    = "alice"                      // TODO: change this
	NodeAddress    = "http://localhost:26657"     // TODO: configure this
	KeyringBackend = keyring.BackendTest          // TODO: change this
	KeyringDir     = "/home/fmorency/.ghostcloud" // TODO: change this
)

// Migrate migrates the given amount of tokens to the given address.
func Migrate(to string, amount int64, denom string, item *store.WorkItem) (*sdk.TxResponse, time.Time, error) {
	// Set up configuration
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(AddressPrefix, AddressPrefix+"pub")

	// Define the transaction's parameters
	coin := sdk.NewCoin(denom, sdk.NewInt(amount))
	memo := item.UUID.String()
	fee := sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(1)))
	gasLimit := uint64(200000)

	// Create a keyring
	inBuf := bufio.NewReader(os.Stdin)
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(interfaceRegistry)
	cdc := codec.NewProtoCodec(interfaceRegistry)
	kr, err := keyring.New(sdk.KeyringServiceName(), KeyringBackend, KeyringDir, inBuf, cdc)
	if err != nil {
		log.Fatalf("Failed to create keyring: %v", err)
	}

	authtypes.RegisterInterfaces(interfaceRegistry)
	cryptocodec.RegisterInterfaces(interfaceRegistry)
	sdk.RegisterInterfaces(interfaceRegistry)
	banktypes.RegisterInterfaces(interfaceRegistry)
	staking.RegisterInterfaces(interfaceRegistry)

	// Fetch account details
	info, err := kr.Key(BankAccount)
	if err != nil {
		log.Fatalf("Failed to get key info: %v", err)
	}

	rClient, err := rpchttp.New(NodeAddress, "/websocket")
	if err != nil {
		log.Fatalf("Failed to create RPC client: %v", err)
	}
	accountRetriever := authtypes.AccountRetriever{}

	// Create a Cosmos SDK client context
	clientCtx := client.Context{}.
		WithChainID(ChainId).
		WithInterfaceRegistry(interfaceRegistry).
		WithCodec(cdc).
		WithKeyring(kr).
		WithTxConfig(authtx.NewTxConfig(cdc, authtx.DefaultSignModes)).
		WithBroadcastMode("sync").
		WithClient(rClient).
		WithAccountRetriever(accountRetriever).
		WithSkipConfirmation(true)

	addr, err := info.GetAddress()
	if err != nil {
		log.Fatalf("Failed to get address: %v", err)
	}

	if err := accountRetriever.EnsureExists(clientCtx, addr); err != nil {
		log.Fatalf("Failed to ensure account exists: %v", err)
	}

	// Create the message
	msg := banktypes.NewMsgSend(addr, sdk.MustAccAddressFromBech32(to), sdk.NewCoins(coin))

	// Create the transaction builder
	txBuilder := clientCtx.TxConfig.NewTxBuilder()
	txBuilder.SetMemo(memo)
	txBuilder.SetFeeAmount(fee)
	txBuilder.SetGasLimit(gasLimit)
	if err := txBuilder.SetMsgs(msg); err != nil {
		log.Fatalf("Failed to set message: %v", err)
	}

	txFactory := tx.Factory{}.
		WithChainID(clientCtx.ChainID).
		WithKeybase(clientCtx.Keyring).
		WithGas(300000).
		WithGasAdjustment(1.0).
		WithSignMode(signing.SignMode_SIGN_MODE_UNSPECIFIED).
		WithAccountRetriever(clientCtx.AccountRetriever).
		WithTxConfig(clientCtx.TxConfig)

	initNum, initSeq := txFactory.AccountNumber(), txFactory.Sequence()
	if initNum == 0 || initSeq == 0 {
		accNum, seqNum, err := accountRetriever.GetAccountNumberSequence(clientCtx, addr)
		if err != nil {
			log.Fatalf("Failed to get account number and sequence: %v", err)
		}

		if initNum == 0 {
			txFactory = txFactory.WithAccountNumber(accNum)
		}

		if initSeq == 0 {
			txFactory = txFactory.WithSequence(seqNum)
		}
	}

	// Sign the transaction
	if err := tx.Sign(txFactory, BankAccount, txBuilder, true); err != nil {
		log.Fatalf("Failed to sign transaction: %v", err)
	}

	// Encode the transaction
	txBytes, err := clientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		log.Fatalf("Failed to encode transaction: %v", err)
	}

	// Broadcast the transaction
	res, err := clientCtx.BroadcastTx(txBytes)
	if err != nil {
		log.Fatalf("Failed to broadcast transaction: %v", err)
	}

	slog.Info("Transaction broadcasted", "hash", res.TxHash)

	// Wait for the transaction to be included in a block
	txResult, err := waitForTx(rClient, res.TxHash)
	if err != nil {
		log.Fatalf("Failed to wait for transaction: %v", err)
	}

	slog.Info("Transaction included in block", "height", txResult.Height)

	txBlock, err := rClient.Block(context.Background(), &txResult.Height)
	if err != nil {
		log.Fatalf("Failed to get block: %v", err)
	}

	blockTime := txBlock.Block.Time.UTC().Truncate(time.Millisecond)

	return res, blockTime, nil
	//slog.Info("Sending tokens", "to", to, "amount", amount, "denom", denom)
	//ctx := context.Background()
	//
	//// Create a Cosmos client instance
	//client, err := cosmosclient.New(ctx,
	//	cosmosclient.WithAddressPrefix(AddressPrefix),
	//	cosmosclient.WithNodeAddress(NodeAddress),
	//	cosmosclient.WithKeyringDir(KeyringDir),
	//	cosmosclient.WithKeyringBackend(KeyringBackend))
	//if err != nil {
	//	slog.Error("could not create client", "error", err)
	//	return nil, err
	//}
	//
	//account, err := client.Account(BankAccount)
	//if err != nil {
	//	slog.Error("could not get account", "error", err)
	//	return nil, err
	//}
	//
	//// TODO: Add memo with UUID
	//txService, err := client.BankSendTx(ctx, account, to, sdk.Coins{{Denom: denom, Amount: sdk.NewInt(amount)}})
	//if err != nil {
	//	slog.Error("could not create send transaction", "error", err)
	//	return nil, err
	//}
	//
	//txResponse, err := txService.Broadcast(ctx)
	//if err != nil {
	//	slog.Error("could not broadcast transaction", "error", err)
	//	return nil, err
	//}
	//
	//// Wait for the transaction to be included in a block
	//err = client.WaitForBlockHeight(ctx, txResponse.Height)
	//if err != nil {
	//	slog.Error("could not wait for block height", "error", err)
	//	return nil, err
	//}
	//
	//return &txResponse, nil
}

func waitForTx(rClient client.TendermintRPC, hash string) (*coretypes.ResultTx, error) {
	bHash, err := hex.DecodeString(hash)
	if err != nil {
		log.Fatalf("Failed to decode transaction hash: %v", err)
	}
	for {
		r, err := rClient.Tx(context.Background(), bHash, false)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				err := waitForNextBlock(rClient)
				if err != nil {
					return nil, err
				}
				continue
			}
			return nil, err
		}
		return r, nil
	}
}

func getLatestBlockHeight(client client.TendermintRPC) (int64, error) {
	status, err := client.Status(context.Background())
	if err != nil {
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
				return err
			}
			if latestHeight >= height {
				return nil
			}
		case <-time.After(30 * time.Second):
			return fmt.Errorf("timeout exceeded waiting for block")
		}
	}
}

func waitForNextBlock(client client.TendermintRPC) error {
	latestHeight, err := getLatestBlockHeight(client)
	if err != nil {
		return err
	}
	return waitForBlockHeight(client, latestHeight+1)
}
