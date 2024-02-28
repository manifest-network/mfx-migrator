package chain

import (
	"context"
	"log/slog"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ignite/cli/ignite/pkg/cosmosclient"
)

const (
	AddressPrefix  = "gc"                         // TODO: change this
	BankAccount    = "alice"                      // TODO: change this
	NodeAddress    = "http://localhost:26657"     // TODO: configure this
	KeyringBackend = keyring.BackendTest          // TODO: change this
	KeyringDir     = "/home/fmorency/.ghostcloud" // TODO: change this
)

// Migrate migrates the given amount of tokens to the given address.
func Migrate(to string, amount int64, denom string) (*cosmosclient.Response, error) {
	slog.Info("Sending tokens", "to", to, "amount", amount, "denom", denom)
	ctx := context.Background()

	// Create a Cosmos client instance
	client, err := cosmosclient.New(ctx,
		cosmosclient.WithAddressPrefix(AddressPrefix),
		cosmosclient.WithNodeAddress(NodeAddress),
		cosmosclient.WithKeyringDir(KeyringDir),
		cosmosclient.WithKeyringBackend(KeyringBackend))
	if err != nil {
		slog.Error("could not create client", "error", err)
		return nil, err
	}

	account, err := client.Account(BankAccount)
	if err != nil {
		slog.Error("could not get account", "error", err)
		return nil, err
	}

	txService, err := client.BankSendTx(ctx, account, to, sdk.Coins{{Denom: denom, Amount: sdk.NewInt(amount)}})
	if err != nil {
		slog.Error("could not create send transaction", "error", err)
		return nil, err
	}

	txResponse, err := txService.Broadcast(ctx)
	if err != nil {
		slog.Error("could not broadcast transaction", "error", err)
		return nil, err
	}

	// Wait for the transaction to be included in a block
	err = client.WaitForBlockHeight(ctx, txResponse.Height)
	if err != nil {
		slog.Error("could not wait for block height", "error", err)
		return nil, err
	}

	return &txResponse, nil
}

// GetBlockTime returns the time of the block at the given height.
func GetBlockTime(height int64) (*time.Time, error) {
	ctx := context.Background()

	// Create a Cosmos client instance
	client, err := cosmosclient.New(ctx,
		cosmosclient.WithAddressPrefix(AddressPrefix),
		cosmosclient.WithNodeAddress(NodeAddress),
		cosmosclient.WithKeyringDir(KeyringDir),
		cosmosclient.WithKeyringBackend(KeyringBackend))
	if err != nil {
		slog.Error("could not create client", "error", err)
		return nil, err
	}

	block, err := client.RPC.Block(ctx, &height)
	if err != nil {
		slog.Error("could not get block", "error", err)
		return nil, err
	}

	return &block.Block.Time, nil
}
