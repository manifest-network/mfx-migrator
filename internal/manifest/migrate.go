package manifest

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math/big"
	"os/exec"
	"time"

	"github.com/pkg/errors"

	"github.com/liftedinit/mfx-migrator/internal/config"

	"github.com/liftedinit/mfx-migrator/internal/store"
)

const (
	OutputFormat = "json"
)

var (
	gas = []string{"--gas", "auto"}
	yes = []string{"--yes"}
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

// executeCommand executes the provided command and returns the output.
func executeCommand(name string, arg ...string) ([]byte, error) {
	cmd := exec.Command(name, arg...)
	slog.Debug("Executing command", "command", cmd.String())
	output, err := cmd.Output()
	slog.Debug("Command output", "output", string(output))
	if err != nil {
		var exitErr *exec.ExitError
		var resErr error
		if errors.As(err, &exitErr) {
			resErr = errors.WithMessage(err, fmt.Sprintf("failed to execute command: %s", string(exitErr.Stderr)))
		} else {
			resErr = errors.WithMessage(err, "failed to execute command")
		}
		return nil, errors.WithMessage(resErr, "failed to execute command")
	}
	return output, nil
}

// unmarshalOutput unmarshals the provided JSON output into the provided destination.
func unmarshalOutput(output []byte, dest interface{}) error {
	if err := json.Unmarshal(output, dest); err != nil {
		return errors.WithMessage(err, "failed to unmarshal output")
	}
	return nil
}

// Migrate migrates the given amount of tokens to the specified address.
func Migrate(item *store.WorkItem, migrateConfig config.MigrateConfig, denom string, amount *big.Int) (*CosmosTx, *time.Time, error) {
	node := []string{"--node", migrateConfig.NodeAddress}
	chainId := []string{"--chain-id", migrateConfig.ChainID}
	keyringBackend := []string{"--keyring-backend", migrateConfig.KeyringBackend}
	home := []string{"--home", migrateConfig.ChainHome}
	from := []string{"--from", migrateConfig.BankAddress}
	gasAdjustment := []string{"--gas-adjustment", fmt.Sprintf("%f", migrateConfig.GasAdjustment)}
	gasPrice := []string{"--gas-prices", fmt.Sprintf("%f", migrateConfig.GasPrice) + denom}
	output := []string{"--output", OutputFormat}

	// Send the tokens to the manifest address
	txSend := []string{"tx", "bank", "send", migrateConfig.BankAddress, item.ManifestAddress, amount.String() + denom}
	txSend = append(txSend, node...)
	txSend = append(txSend, chainId...)
	txSend = append(txSend, keyringBackend...)
	txSend = append(txSend, home...)
	txSend = append(txSend, from...)
	txSend = append(txSend, gas...)
	txSend = append(txSend, gasAdjustment...)
	txSend = append(txSend, gasPrice...)
	txSend = append(txSend, output...)
	txSend = append(txSend, yes...)
	o, err := executeCommand(migrateConfig.Binary, txSend...)
	if err != nil {
		return nil, nil, err
	}

	// Unmarshal the transaction response
	var tx CosmosTx
	if err = unmarshalOutput(o, &tx); err != nil {
		return nil, nil, err
	}
	if tx.Code != 0 {
		return nil, nil, errors.Errorf("failed to execute transaction: %s", tx.RawLog)
	}

	// Wait for the transaction to be included in a block
	qWaitTx := []string{"q", "event-query-tx-for", tx.TxHash}
	qWaitTx = append(qWaitTx, node...)
	qWaitTx = append(qWaitTx, home...)
	qWaitTx = append(qWaitTx, output...)
	o, err = executeCommand(migrateConfig.Binary, qWaitTx...)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "failed to wait for transaction")
	}

	var res EventQueryTxFor
	if err = unmarshalOutput(o, &res); err != nil {
		return nil, nil, err
	}

	// Fetch the block header for the transaction to get the block time
	qBlock := []string{"q", "block", "--type", "height", res.Height}
	qBlock = append(qBlock, node...)
	qBlock = append(qBlock, home...)
	qBlock = append(qBlock, output...)
	o, err = executeCommand(migrateConfig.Binary, qBlock...)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "failed to fetch block")
	}

	var block BlockHeader
	if err = unmarshalOutput(o, &block); err != nil {
		return nil, nil, err
	}

	blockTime := block.Header.Time.UTC().Truncate(time.Millisecond)
	return &tx, &blockTime, nil
}
