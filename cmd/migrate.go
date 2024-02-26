package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strconv"

	"github.com/go-resty/resty/v2"
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
		urlStr := viper.GetString("url")
		uuidStr := viper.GetString("migrate-uuid")
		username := viper.GetString("username")
		password := viper.GetString("password")
		neighborhood := viper.GetUint64("neighborhood")

		slog.Debug("args", "url", urlStr, "uuid", uuidStr, "username", username, "neighborhood", neighborhood)

		if username == "" || password == "" {
			slog.Error("username and password are required")
			return errors.New("username and password are required")
		}

		// Parse the URL
		url, err := url.Parse(urlStr)
		if err != nil {
			slog.Error("could not parse URL", "error", err)
			return err
		}

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
		slog.Info("migrating", "uuid", item.UUID)

		slog.Debug("setting migration status to 'migrating'")

		r := resty.New().SetBaseURL(url.String()).SetPathParam("neighborhood", strconv.FormatUint(neighborhood, 10))
		s := store.NewWithClient(r)

		// Login to the remote database
		token, err := s.Login(username, password)
		if err != nil {
			slog.Error("could not login", "error", err)
			return err
		}
		if token == "" {
			slog.Error("no token returned")
			return err
		}

		// Create a new authenticated HTTP client and set it on the store
		r.SetAuthToken(token)

		// Set the work item status to 'migrating'
		response, err := s.UpdateWorkItem(*item, store.MIGRATING)
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
		err = localstate.SaveState(item)
		if err != nil {
			slog.Error("could not save state", "error", err)
			return err
		}

		// 3. Execute the migration
		// 4. Verify the migration was successful
		// 5. POST the 'talib/complete-work/' endpoint to complete the work item
		//   5.1. If the work item is completed, the `*.uuid` file should be removed
		//        Note: Completed involves both successful and failed migrations.
		//              Failed migrations should have a reason for failure persisted to the database.

		return nil
	},
}

func init() {
	migrateCmd.Flags().String("uuid", "", "UUID of the work item to claim")
	err := viper.BindPFlag("migrate-uuid", migrateCmd.Flags().Lookup("uuid"))
	if err != nil {
		slog.Error("unable to bind flag", "error", err)
	}

	rootCmd.AddCommand(migrateCmd)
}
