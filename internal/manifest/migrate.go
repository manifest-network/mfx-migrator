package manifest

import (
	"encoding/json"
	"log/slog"
	"math/big"
	"os/exec"
	"time"

	"github.com/liftedinit/mfx-migrator/internal/config"
	"github.com/pkg/errors"

	"github.com/liftedinit/mfx-migrator/internal/store"
)

type CosmosTx struct {
	TxHash string `json:"txhash"`
	Code   int    `json:"code"`
	RawLog string `json:"raw_log"`
}

type EventQueryTxFor struct {
	Height string `json:"height"`
}

type BlockHeader struct {
	Header struct {
		Time time.Time `json:"time"`
	}
}

// Migrate migrates the given amount of tokens to the specified address.
func Migrate(item *store.WorkItem, migrateConfig config.MigrateConfig, denom string, amount *big.Int) (*CosmosTx, *time.Time, error) {
	// Run the command to send the tokens
	cmd := exec.Command(migrateConfig.Binary, "tx", "bank", "send",
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
	slog.Info("Transaction command", "command", cmd.String())
	o, err := cmd.Output()
	slog.Info("Transaction output", "output", string(o))
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
	cmd = exec.Command(migrateConfig.Binary, "q", "event-query-tx-for",
		tx.TxHash,
		"--node", migrateConfig.NodeAddress,
		"--home", migrateConfig.ChainHome,
		"--output", "json")
	slog.Info("Waiting for transaction", "command", cmd.String())
	o, err = cmd.Output()
	slog.Info("Transaction included in block", "output", string(o))
	if err != nil {
		return nil, nil, errors.WithMessage(err, "failed to wait for transaction")
	}

	var res EventQueryTxFor
	if err := json.Unmarshal(o, &res); err != nil {
		return nil, nil, errors.WithMessage(err, "failed to unmarshal response")
	}

	cmd = exec.Command(migrateConfig.Binary, "q", "block",
		"--type", "height", res.Height,
		"--node", migrateConfig.NodeAddress,
		"--home", migrateConfig.ChainHome,
		"--output", "json")
	slog.Info("Fetching block", "command", cmd.String())
	o, err = cmd.Output()
	slog.Info("Block fetched", "output", string(o))
	if err != nil {
		return nil, nil, errors.WithMessage(err, "failed to fetch block")
	}

	var block BlockHeader
	if err := json.Unmarshal(o, &block); err != nil {
		return nil, nil, errors.WithMessage(err, "failed to unmarshal block")
	}

	blockTime := block.Header.Time.UTC().Truncate(time.Millisecond)
	return &tx, &blockTime, nil
}
