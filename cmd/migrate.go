package cmd

import (
	"fmt"
	"log/slog"
	"math/big"
	"os"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/liftedinit/mfx-migrator/internal/config"

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
	c := LoadConfigFromCLI("migrate-uuid")
	slog.Debug("args", "c", c)
	if err := c.Validate(); err != nil {
		return err
	}

	migrateConfig := LoadMigrationConfigFromCLI()
	slog.Debug("args", "migrate-c", migrateConfig)
	if err := migrateConfig.Validate(); err != nil {
		return err
	}

	authConfig := LoadAuthConfigFromCLI()
	slog.Debug("args", "auth-c", authConfig)
	if err := authConfig.Validate(); err != nil {
		return err
	}

	slog.Info("Loading state...", "uuid", c.UUID)
	item, err := store.LoadState(c.UUID)
	if err != nil {
		return errors.WithMessage(err, "unable to load state")
	}

	if err := verifyItemStatus(item); err != nil {
		return err
	}

	r := CreateRestClient(cmd.Context(), c.Url, c.Neighborhood)
	if err := AuthenticateRestClient(r, authConfig.Username, authConfig.Password); err != nil {
		return err
	}

	err = migrate(r, item, migrateConfig)

	// The migration failed for some reason, update the work item status and save the state
	if err != nil {
		errStr := err.Error()
		sErr := setAsFailed(r, *item, &errStr)
		if sErr != nil {
			return errors.WithMessage(err, sErr.Error())
		}
	}
	return err

}

func init() {
	SetupMigrateCmdFlags(migrateCmd)
	rootCmd.AddCommand(migrateCmd)
}

// verifyItemStatus verifies the status of the work item is valid for migration.
func verifyItemStatus(item *store.WorkItem) error {
	if !(item.Status == store.CLAIMED || item.Status == store.MIGRATING) {
		return fmt.Errorf("work item status not valid for migration: %s, %s", item.UUID, item.Status)
	}
	return nil
}

// compareItems compares the local and remote work items to ensure they match.
func compareItems(item *store.WorkItem, remoteItem *store.WorkItem) error {
	if !item.Equal(*remoteItem) {
		return fmt.Errorf("local and remote work items do not match: %s, %s", item.UUID, remoteItem.UUID)
	}
	return nil
}

func setupStringCmdFlags(command *cobra.Command) {
	args := []struct {
		name     string
		key      string
		value    string
		usage    string
		required bool
	}{
		{"chain-id", "chain-id", "manifest-1", "Chain ID of the blockchain to migrate to", false},
		{"address-prefix", "address-prefix", "manifest", "Address prefix of the blockchain to migrate to", false},
		{"node-address", "node-address", "http://localhost:26657", "Node address of the blockchain to migrate to", false},
		{"keyring-backend", "keyring-backend", "test", "Keyring backend to use", false},
		{"bank-address", "bank-address", "bank", "Bank address to send tokens from", false},
		{"chain-home", "chain-home", "", "Root directory of the chain configuration", false},
		{"uuid", "migrate-uuid", "", "UUID of the work item to migrate", true},
	}

	for _, arg := range args {
		command.Flags().String(arg.name, arg.value, arg.usage)
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

func setupUIntCmdFlags(command *cobra.Command) {
	args := []struct {
		name  string
		key   string
		value uint
		usage string
	}{
		{"wait-for-tx-timeout", "wait-for-tx-timeout", 15, "Number of seconds spent waiting for the transaction to be included in a block"},
		{"wait-for-block-timeout", "wait-for-block-timeout", 30, "Number of seconds spent waiting for the block to be committed"},
	}

	for _, arg := range args {
		command.Flags().Uint(arg.name, arg.value, arg.usage)
		if err := viper.BindPFlag(arg.key, command.Flags().Lookup(arg.name)); err != nil {
			slog.Error(ErrorBindingFlag, "error", err)
		}
	}

}

func SetupMigrateCmdFlags(command *cobra.Command) {
	setupStringCmdFlags(command)
	setupUIntCmdFlags(command)
}

func mapToken(symbol string, tokenMap map[string]utils.TokenInfo) (*utils.TokenInfo, error) {
	if _, ok := tokenMap[symbol]; !ok {
		return nil, fmt.Errorf("token %s not found in token map", symbol)
	}
	info := tokenMap[symbol]
	return &info, nil
}

// migrate migrates a work item to the Manifest Ledger.
func migrate(r *resty.Client, item *store.WorkItem, config config.MigrateConfig) error {
	slog.Info("Migrating work item...", "uuid", item.UUID)

	remoteItem, err := store.GetWorkItem(r, item.UUID)
	if err != nil {
		return errors.WithMessage(err, "error getting remote work item")
	}

	// Verify the item is ready for migration
	if err = verifyItemStatus(remoteItem); err != nil {
		return errors.WithMessage(err, "error verifying item status")
	}

	// Verify the local and remote items match
	if err = compareItems(item, remoteItem); err != nil {
		return errors.WithMessage(err, "error comparing items")
	}

	txInfo, err := many.GetTxInfo(r, item.ManyHash)
	if err != nil {
		return errors.WithMessage(err, "error getting MANY tx info")
	}

	// Check the MANY transaction info
	if err = many.CheckTxInfo(txInfo, item.UUID, item.ManifestAddress); err != nil {
		return errors.WithMessage(err, "error checking MANY tx info")
	}

	// Map the MANY token symbol to the destination chain token
	tokenInfo, err := mapToken(txInfo.Arguments.Symbol, config.TokenMap)
	if err != nil {
		return errors.WithMessage(err, "error mapping token")
	}

	slog.Debug("Amount before conversion", "amount", txInfo.Arguments.Amount)

	// Convert the amount to the destination chain precision
	// NOTE: currentPrecision is hardcoded to 9 for now as all tokens on the MANY network have 9 digits places
	amount, err := utils.ConvertPrecision(txInfo.Arguments.Amount, 9, tokenInfo.Precision)
	if err != nil {
		return errors.WithMessage(err, "error converting token to destination precision")
	}

	var newItem = *item

	// If the item status is not MIGRATING, set it to MIGRATING
	if newItem.Status != store.MIGRATING {
		if err = setAsMigrating(r, newItem); err != nil {
			return errors.WithMessage(err, "could not set status to MIGRATING")
		}
	}

	// Send the tokens
	txHash, blockTime, err := sendTokens(&newItem, config, tokenInfo.Denom, amount)
	if err != nil {
		return errors.WithMessage(err, "error sending tokens")
	}

	slog.Info("Migration succeeded on chain...", "hash", txHash, "timestamp", blockTime)
	// Set the status to COMPLETED
	if err = setAsCompleted(r, newItem, txHash, blockTime); err != nil {
		return errors.WithMessage(err, "error setting status to COMPLETED")
	}

	// Delete the state file, as the work item is now completed and the state is stored in the database
	if err = deleteState(&newItem); err != nil {
		return errors.WithMessage(err, "error deleting state")
	}

	slog.Info("Migration complete", "uuid", newItem.UUID)

	return nil
}

func deleteState(item *store.WorkItem) error {
	slog.Info("Deleting local state file...")
	if err := os.Remove(fmt.Sprintf("%s.json", item.UUID)); err != nil {
		return errors.WithMessage(err, "error deleting state")
	}
	return nil
}

// setAsMigrating sets the status of the work item to MIGRATING and updates the state.
func setAsMigrating(r *resty.Client, newItem store.WorkItem) error {
	newItem.Status = store.MIGRATING
	if err := store.UpdateWorkItemAndSaveState(r, newItem); err != nil {
		return errors.WithMessage(err, "error setting status to MIGRATING")
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
		return errors.WithMessage(err, "error setting status to COMPLETED")
	}
	return nil
}

func setAsFailed(r *resty.Client, newItem store.WorkItem, errStr *string) error {
	newItem.Status = store.FAILED
	newItem.Error = errStr
	if err := store.UpdateWorkItemAndSaveState(r, newItem); err != nil {
		return errors.WithMessage(err, "error setting status to FAILED")
	}
	return nil
}

// sendTokens sends the tokens from the bank account to the user account.
func sendTokens(item *store.WorkItem, config config.MigrateConfig, denom string, amount *big.Int) (*string, *time.Time, error) {
	txResponse, blockTime, err := manifest.Migrate(item, config, denom, amount)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "error during migration, operator intervention required")
	}

	if txResponse.Code != 0 {
		return nil, nil, errors.WithMessagef(err, "migration failed: %s", txResponse.RawLog)
	}

	return &txResponse.TxHash, blockTime, nil
}
