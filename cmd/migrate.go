package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/liftedinit/mfx-migrator/internal/chain"
	"github.com/liftedinit/mfx-migrator/internal/store"
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Execute the MFX token migration associated with the given UUID.",
	RunE: func(cmd *cobra.Command, args []string) error {
		config := LoadConfigFromCLI("migrate-uuid")
		slog.Debug("args", "config", config)
		if err := config.Validate(); err != nil {
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

		r := CreateRestClient(config.Url, config.Neighborhood)
		if err := AuthenticateRestClient(r, config.Username, config.Password); err != nil {
			return err
		}

		return migrate(r, item)
	},
}

func init() {
	setupMigrateFlags()
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

func setupMigrateFlags() {
	migrateCmd.Flags().String("uuid", "", "UUID of the work item to claim")
	if err := migrateCmd.MarkFlagRequired("uuid"); err != nil {
		slog.Error(ErrorMarkingFlagRequired, "error", err)
	}
	if err := viper.BindPFlag("migrate-uuid", migrateCmd.Flags().Lookup("uuid")); err != nil {
		slog.Error(ErrorBindingFlag, "error", err)
	}
}

// migrate migrates a work item to the Manifest Ledger.
func migrate(r *resty.Client, item *store.WorkItem) error {
	slog.Info("Migrating work item...", "uuid", item.UUID)

	remoteItem, err := store.GetWorkItem(r, item.UUID)
	if err != nil {
		return err
	}

	// Verify the item is ready for migration
	if err = verifyItemStatus(remoteItem); err != nil {
		return err
	}

	// Verify the local and remote items match
	if err = compareItems(item, remoteItem); err != nil {
		return err
	}

	var newItem = *item

	// If the item status is not MIGRATING, set it to MIGRATING
	if newItem.Status != store.MIGRATING {
		if err = setAsMigrating(r, newItem); err != nil {
			return err
		}
	}

	// Send the tokens
	txHash, height, err := sendTokens(r, &newItem)
	if err != nil {
		return err
	}

	slog.Info("Migration succeeded on chain...", "hash", txHash, "height", height)
	// Set the status to COMPLETED
	if err = setAsCompleted(r, newItem, txHash); err != nil {
		return err
	}

	// Set the migration datetime from the block time
	// The timestamp is not available in the response, so we need to make another request to get the block time
	// See https://github.com/ignite/cli/blob/f3ab0d709ec41e31a1c57f2fe86c8902d8a50497/ignite/pkg/cosmosclient/txservice.go#L63-L65
	if err = updateBlockTime(r, newItem, *height); err != nil {
		return err
	}

	// Delete the state file, as the work item is now completed and the state is stored in the database
	if err = deleteState(&newItem); err != nil {
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
func setAsCompleted(r *resty.Client, newItem store.WorkItem, txHash *string) error {
	newItem.Status = store.COMPLETED
	newItem.ManifestHash = txHash
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

// updateBlockTime updates the block time of the work item and updates the state.
func updateBlockTime(r *resty.Client, newItem store.WorkItem, height int64) error {
	slog.Info("Getting block time...", "height", height)
	blockTime, err := chain.GetBlockTime(height)
	if err != nil {
		slog.Error("could not get block time", "error", err)
		return err
	}

	if blockTime == nil {
		slog.Error("block time is nil")
		return errors.New("block time is nil")
	}

	newItem.ManifestDatetime = blockTime
	if err = store.UpdateWorkItemAndSaveState(r, newItem); err != nil {
		return err
	}

	return nil
}

// sendTokens sends the tokens from the bank account to the user account.
func sendTokens(r *resty.Client, item *store.WorkItem) (*string, *int64, error) {
	txResponse, err := chain.Migrate(item.ManifestAddress, 10, "token")
	//txResponse, err := chain.Migrate("gc13ar86s8yqpne8gyqez9jvs9uhaa6j0yjqcx02r", 10, "token")
	if err != nil {
		slog.Error("error during migration, operator intervention required", "error", err)
		errStr := err.Error()
		if err = setAsFailed(r, *item, &errStr); err != nil {
			return nil, nil, err
		}

		return nil, nil, err
	}

	if txResponse.Code != 0 {
		slog.Error("migration failed", "code", txResponse.Code, "log", txResponse.RawLog)
		return nil, nil, fmt.Errorf("migration failed: %s", txResponse.RawLog)
	}

	return &txResponse.TxHash, &txResponse.Height, nil
}
