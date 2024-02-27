package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/liftedinit/mfx-migrator/internal/chain"
	"github.com/liftedinit/mfx-migrator/internal/localstate"
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

		// Load the local state from the *.uuid file
		item, err := localstate.LoadState(config.UUID)
		if err != nil {
			slog.Error("unable to load state", "error", err)
			return err
		}

		// Verify the work item status is valid for migration
		if !(item.Status == store.CLAIMED || item.Status == store.MIGRATING) {
			slog.Error("local work item status not valid for migration", "uuid", config.UUID, "status", item.Status)
			return fmt.Errorf("local work item status not valid for migration: %s, %s", config.UUID, item.Status)
		}

		r := CreateRestClient(config.Url, config.Neighborhood)
		if err := AuthenticateRestClient(r, config.Username, config.Password); err != nil {
			return err
		}

		// Get the work item from the remote database
		slog.Debug("getting work item from remote database", "uuid", item.UUID)
		remoteItem, err := store.GetWorkItem(r, item.UUID)
		if err != nil {
			slog.Error("could not get work item from remote", "error", err)
			return err
		}

		// Verify the work item status is valid for migration
		if !(remoteItem.Status == store.CLAIMED || remoteItem.Status == store.MIGRATING) {
			slog.Error("remote work item status not valid for migration", "uuid", config.UUID, "status", item.Status)
			return fmt.Errorf("remote work item status not valid for migration: %s, %s", config.UUID, item.Status)
		}

		// Compare the local and remote work items
		slog.Debug("comparing local and remote work items")
		if !item.Equal(*remoteItem) {
			slog.Error("local and remote work items do not match", "local", item, "remote", remoteItem)
			return fmt.Errorf("local and remote work items do not match: %s, %s", item.UUID, remoteItem.UUID)
		}
		slog.Debug("local and remote work items match")

		// Set the work item status to 'migrating' if it is not already
		if item.Status != store.MIGRATING {
			slog.Debug("setting work item status to migrating", "uuid", item.UUID)
			// Set the work item status to 'migrating'
			response, err := store.UpdateWorkItem(r, *item, store.MIGRATING)
			if err != nil {
				slog.Error("could not update work item", "error", err)
				return err
			}

			// An error occurred setting the work item status to 'migrating'
			if response.Status != store.MIGRATING {
				slog.Error("work item not migrating", "uuid", item.UUID)
				return fmt.Errorf("work item not migrating: %s", item.UUID)
			}

			item.Status = store.MIGRATING
		}

		// Save the local state
		slog.Debug("saving state", "item", item)
		err = localstate.SaveState(item)
		if err != nil {
			slog.Error("could not save state", "error", err)
			return err
		}

		// Execute the migration
		txResponse, blockTime, err := chain.Migrate(item.ManifestAddress, 10, "token")
		//txResponse, blockTime, err := chain.Migrate("gc13ar86s8yqpne8gyqez9jvs9uhaa6j0yjqcx02r", 10, "token")
		if err != nil {
			slog.Error("could not migrate", "error", err)

			// Try to set the work item status to 'failed'
			slog.Debug("setting work item status to failed", "uuid", item.UUID)
			errStr := err.Error()
			item.Error = &errStr
			response, err := store.UpdateWorkItem(r, *item, store.FAILED)
			if err != nil {
				slog.Error("could not update work item", "error", err)
				return err
			}

			if response.Status != store.FAILED {
				slog.Error("work item not failed", "uuid", item.UUID)
				return fmt.Errorf("work item not failed: %s", item.UUID)
			}

			item.Status = store.FAILED
			slog.Debug("saving state", "item", item)
			err = localstate.SaveState(item)
			if err != nil {
				slog.Error("could not save state", "error", err)
				return err
			}

			return err
		}

		// Verify the migration was successful
		slog.Debug("verifying migration", "response", txResponse)
		if txResponse.Code != 0 {
			slog.Error("migration failed", "code", txResponse.Code, "log", txResponse.RawLog)
			return fmt.Errorf("migration failed: %s", txResponse.RawLog)
		}

		if blockTime == nil {
			slog.Error("block time is nil")
			return errors.New("block time is nil")
		}

		// Show the transaction hash
		slog.Info("migration succeeded", "tx_hash", txResponse.TxHash, "timestamp", txResponse.Timestamp)

		item.ManifestDatetime = blockTime
		item.ManifestHash = &txResponse.TxHash

		slog.Debug("setting work item status to completed", "uuid", item.UUID)
		response, err := store.UpdateWorkItem(r, *item, store.COMPLETED)
		if err != nil {
			slog.Error("could not update work item", "error", err)
			return err
		}

		// An error occurred setting the work item status to 'completed'
		if response.Status != store.COMPLETED {
			slog.Error("work item not completed", "uuid", item.UUID)
			slog.Debug("saving state", "item", item)
			item.Status = store.COMPLETED
			err = localstate.SaveState(item)
			if err != nil {
				slog.Error("could not save state", "error", err)
				return err
			}
			return fmt.Errorf("work item not completed: %s", item.UUID)
		}

		// At this point, the work item status is 'completed' and we can delete the local state
		slog.Debug("deleting state", "uuid", item.UUID)
		err = os.Remove(fmt.Sprintf("%s.json", item.UUID))
		if err != nil {
			slog.Error("could not delete state", "error", err)
			return err
		}

		return nil
	},
}

func init() {
	migrateCmd.Flags().String("uuid", "", "UUID of the work item to claim")
	err := migrateCmd.MarkFlagRequired("uuid")
	if err != nil {
		slog.Error("could not mark flag required", "error", err)
	}
	err = viper.BindPFlag("migrate-uuid", migrateCmd.Flags().Lookup("uuid"))
	if err != nil {
		slog.Error("unable to bind flag", "error", err)
	}

	rootCmd.AddCommand(migrateCmd)
}
