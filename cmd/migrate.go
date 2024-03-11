package cmd

import (
	"fmt"
	"log/slog"
	"math"
	"os"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/liftedinit/mfx-migrator/internal/many"
	"github.com/liftedinit/mfx-migrator/internal/utils"

	"github.com/liftedinit/mfx-migrator/internal/manifest"
	"github.com/liftedinit/mfx-migrator/internal/store"
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Execute the MFX token migration associated with the given UUID.",
	RunE:  MigrateCmdRunE,
}

func MigrateCmdRunE(cmd *cobra.Command, args []string) error {
	config := LoadConfigFromCLI("migrate-uuid")
	slog.Debug("args", "config", config)
	if err := config.Validate(); err != nil {
		return err
	}

	migrateConfig := LoadMigrationConfigFromCLI()
	slog.Debug("args", "migrate-config", migrateConfig)
	if err := migrateConfig.Validate(); err != nil {
		return err
	}

	slog.Info("Loading state...", "uuid", config.UUID)
	item, err := store.LoadState(config.UUID)
	if err != nil {
		slog.Error("unable to load state", "error", err)
		return err
	}

	if err := verifyItemStatus(item); err != nil {
		return err
	}

	r := CreateRestClient(cmd.Context(), config.Url, config.Neighborhood)
	if err := AuthenticateRestClient(r, config.Username, config.Password); err != nil {
		return err
	}

	return migrate(r, item, migrateConfig)

}

func init() {
	SetupMigrateCmdFlags(migrateCmd)
	rootCmd.AddCommand(migrateCmd)
}

// verifyItemStatus verifies the status of the work item is valid for migration.
func verifyItemStatus(item *store.WorkItem) error {
	if !(item.Status == store.CLAIMED || item.Status == store.MIGRATING) {
		slog.Error("work item status not valid for migration", "uuid", item.UUID, "status", item.Status)
		return fmt.Errorf("work item status not valid for migration: %s, %s", item.UUID, item.Status)
	}
	return nil
}

// compareItems compares the local and remote work items to ensure they match.
func compareItems(item *store.WorkItem, remoteItem *store.WorkItem) error {
	if !item.Equal(*remoteItem) {
		slog.Error("local and remote work items do not match", "local", item, "remote", remoteItem)
		return fmt.Errorf("local and remote work items do not match: %s, %s", item.UUID, remoteItem.UUID)
	}
	return nil
}

func SetupMigrateCmdFlags(command *cobra.Command) {
	args := []struct {
		name     string
		key      string
		usage    string
		required bool
	}{
		{"chain-id", "chain-id", "Chain ID of the blockchain to migrate to", false},
		{"address-prefix", "address-prefix", "Address prefix of the blockchain to migrate to", false},
		{"node-address", "node-address", "Node address of the blockchain to migrate to", false},
		{"keyring-backend", "keyring-backend", "Keyring backend to use", false},
		{"bank-address", "bank-address", "Bank address to send tokens from", false},
		{"chain-home", "chain-home", "Root directory of the chain configuration", false},
		{"uuid", "migrate-uuid", "UUID of the work item to claim", true},
	}

	for _, arg := range args {
		command.Flags().String(arg.name, "", arg.usage)
		if err := viper.BindPFlag(arg.key, command.Flags().Lookup(arg.name)); err != nil {
			slog.Error(ErrorBindingFlag, "error", err)
		}
		if arg.required {
			if err := command.MarkFlagRequired(arg.name); err != nil {
				slog.Error(ErrorMarkingFlagRequired, "error", err)
			}
		}
	}
}

func mapToken(symbol string, tokenMap map[string]utils.TokenInfo) (*utils.TokenInfo, error) {
	if _, ok := tokenMap[symbol]; !ok {
		slog.Error("token not found in token map", "symbol", symbol, "tokenMap", tokenMap)
		return nil, fmt.Errorf("token %s not found in token map", symbol)
	}
	info := tokenMap[symbol]
	return &info, nil
}

// convertPrecision adjusts the precision of an integer number.
// TODO: Harden this function
func convertPrecision(n int64, currentPrecision int64, targetPrecision int64) (int64, error) {
	if currentPrecision == targetPrecision {
		return 0, fmt.Errorf("current precision is equal to target precision: %d", currentPrecision)
	}

	// Calculate the difference in precision
	precisionDiff := targetPrecision - currentPrecision

	if precisionDiff > 0 {
		// Increase precision by multiplying
		return n * int64(math.Pow(10, float64(precisionDiff))), nil
	} else {
		// Decrease precision by dividing
		return n / int64(math.Pow(10, float64(-precisionDiff))), nil
	}
}

// migrate migrates a work item to the Manifest Ledger.
func migrate(r *resty.Client, item *store.WorkItem, config MigrateConfig) error {
	slog.Info("Migrating work item...", "uuid", item.UUID)

	remoteItem, err := store.GetWorkItem(r, item.UUID)
	if err != nil {
		slog.Error("error getting remote work item", "error", err)
		return err
	}

	// Verify the item is ready for migration
	if err = verifyItemStatus(remoteItem); err != nil {
		slog.Error("error verifying item status", "error", err)
		return err
	}

	// Verify the local and remote items match
	if err = compareItems(item, remoteItem); err != nil {
		slog.Error("error comparing items", "error", err)
		return err
	}

	txInfo, err := many.GetTxInfo(r, item.ManyHash)
	if err != nil {
		slog.Error("error getting MANY tx info", "error", err)
		return err
	}

	// Check the MANY transaction info
	if err = many.CheckTxInfo(txInfo, item.UUID, item.ManifestAddress); err != nil {
		slog.Error("error checking MANY tx info", "error", err)
		return err
	}

	// Map the MANY token symbol to the destination chain token
	tokenInfo, err := mapToken(txInfo.Arguments.Symbol, config.TokenMap)
	if err != nil {
		slog.Error("error mapping token", "error", err)
		return err
	}

	slog.Debug("Amount before conversion", "amount", txInfo.Arguments.Amount)

	// Convert the amount to the destination chain precision
	// TODO: currentPrecision is hardcoded to 9 for now as all tokens on the MANY network have 9 digits places
	amount, err := convertPrecision(txInfo.Arguments.Amount, 9, tokenInfo.Precision)
	if err != nil {
		slog.Error("error converting token to destination precision", "error", err)
		return err
	}

	slog.Debug("Amount after conversion", "amount", amount)
	// Block migration if the amount is less than or equal to 0
	// This can happen if the amount is less than the precision of the destination chain
	// E.g., TokenA has 9 decimal places and TokenB has 6 decimal places, then the minimum amount of TokenA that can be migrated is 1000 resulting in 1 TokenB
	if amount <= 0 {
		slog.Error("amount after conversion is less than or equal to 0", "amount", amount)
		return fmt.Errorf("amount after conversion is less than or equal to 0: %d", amount)
	}

	var newItem = *item

	// If the item status is not MIGRATING, set it to MIGRATING
	if newItem.Status != store.MIGRATING {
		if err = setAsMigrating(r, newItem); err != nil {
			slog.Error("error setting status to MIGRATING", "error", err)
			return err
		}
	}

	// Send the tokens
	txHash, blockTime, err := sendTokens(r, &newItem, config, tokenInfo.Denom, amount)
	if err != nil {
		slog.Error("error sending tokens", "error", err)
		return err
	}

	slog.Info("Migration succeeded on chain...", "hash", txHash, "timestamp", blockTime)
	// Set the status to COMPLETED
	if err = setAsCompleted(r, newItem, txHash, blockTime); err != nil {
		slog.Error("error setting status to COMPLETED", "error", err)
		return err
	}

	// Delete the state file, as the work item is now completed and the state is stored in the database
	if err = deleteState(&newItem); err != nil {
		slog.Error("error deleting state", "error", err)
		return err
	}

	slog.Info("Migration complete", "uuid", newItem.UUID)

	return nil
}

func deleteState(item *store.WorkItem) error {
	slog.Info("Deleting local state file...")
	if err := os.Remove(fmt.Sprintf("%s.json", item.UUID)); err != nil {
		slog.Error("could not delete state", "error", err)
		return err
	}
	return nil
}

// setAsMigrating sets the status of the work item to MIGRATING and updates the state.
func setAsMigrating(r *resty.Client, newItem store.WorkItem) error {
	newItem.Status = store.MIGRATING
	if err := store.UpdateWorkItemAndSaveState(r, newItem); err != nil {
		return err
	}
	return nil
}

// setAsCompleted sets the status of the work item to COMPLETED.
// It also sets the manifest hash and updates the state.
func setAsCompleted(r *resty.Client, newItem store.WorkItem, txHash *string, blockTime *time.Time) error {
	newItem.Status = store.COMPLETED
	newItem.ManifestHash = txHash
	newItem.ManifestDatetime = blockTime
	if err := store.UpdateWorkItemAndSaveState(r, newItem); err != nil {
		return err
	}
	return nil
}

func setAsFailed(r *resty.Client, newItem store.WorkItem, errStr *string) error {
	newItem.Status = store.FAILED
	newItem.Error = errStr
	if err := store.UpdateWorkItemAndSaveState(r, newItem); err != nil {
		return err
	}
	return nil
}

// sendTokens sends the tokens from the bank account to the user account.
func sendTokens(r *resty.Client, item *store.WorkItem, config MigrateConfig, denom string, amount int64) (*string, *time.Time, error) {
	txResponse, blockTime, err := manifest.Migrate(item, manifest.MigrationConfig{
		ChainID:        config.ChainID,
		NodeAddress:    config.NodeAddress,
		KeyringBackend: config.KeyringBackend,
		ChainHome:      config.ChainHome,
		AddressPrefix:  config.AddressPrefix,
		BankAddress:    config.BankAddress,
		TokenMap:       config.TokenMap,
	}, denom, amount)
	if err != nil {
		slog.Error("error during migration, operator intervention required", "error", err)
		errStr := err.Error()
		if fErr := setAsFailed(r, *item, &errStr); fErr != nil {
			return nil, nil, fErr
		}

		return nil, nil, err
	}

	if txResponse.Code != 0 {
		slog.Error("migration failed", "code", txResponse.Code, "log", txResponse.RawLog)
		return nil, nil, fmt.Errorf("migration failed: %s", txResponse.RawLog)
	}

	return &txResponse.TxHash, blockTime, nil
}
