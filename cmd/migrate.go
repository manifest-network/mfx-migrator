package cmd

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/liftedinit/mfx-migrator/internal/localstate"
	"github.com/liftedinit/mfx-migrator/internal/store"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Execute the MFX token migration associated with the given UUID.",
	RunE: func(cmd *cobra.Command, args []string) error {
		url := viper.GetString("url")
		uuidStr := viper.GetString("uuid")

		if uuidStr == "" {
			slog.Error("uuid is required")
			return errors.New("uuid is required")
		}

		// Load the local state from the *.uuid file
		item, err := localstate.LoadState(uuidStr)
		if err != nil {
			slog.Error("unable to load state", "error", err)
			return err
		}

		// Verify the work item is claimed
		if item.Status != store.CLAIMED {
			slog.Error("work item not claimed", "uuid", uuidStr)
			return fmt.Errorf("work item not claimed: %s", uuidStr)
		}

		// Execute the migration
		migrate(url, item)

		return nil
	},
}

func migrate(url string, item *store.WorkItem) {
	slog.Debug("migrating", "url", url, "item", item)
	// 3. Execute the migration
	// 4. Verify the migration was successful
	// 5. POST the 'talib/complete-work/' endpoint to complete the work item
	//   5.1. If the work item is completed, the `*.uuid` file should be removed
	//        Note: Completed involves both successful and failed migrations.
	//              Failed migrations should have a reason for failure persisted to the database.
}

func init() {
	migrateCmd.Flags().String("uuid", "", "UUID of the work item to claim")
	err := viper.BindPFlag("uuid", migrateCmd.Flags().Lookup("uuid"))
	if err != nil {
		slog.Error("unable to bind flag", "error", err)
	}

	rootCmd.AddCommand(migrateCmd)
}
