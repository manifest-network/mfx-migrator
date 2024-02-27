package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strconv"

	"github.com/go-resty/resty/v2"
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

		// Verify the work item status is valid for migration
		if !(item.Status == store.CLAIMED || item.Status == store.MIGRATING) {
			slog.Error("local work item status not valid for migration", "uuid", uuidStr, "status", item.Status)
			return fmt.Errorf("local work item status not valid for migration: %s, %s", uuidStr, item.Status)
		}

		// Create a new store instance
		r := resty.New().SetBaseURL(url.String()).SetPathParam("neighborhood", strconv.FormatUint(neighborhood, 10))
		s := store.NewWithClient(r)

		// Login to the remote database
		slog.Debug("logging in", "username", username, "password", "[REDACTED]")
		loginResponse, err := r.R().SetBody(map[string]interface{}{"username": username, "password": password}).SetResult(&store.Token{}).Post("/auth/login")
		if err != nil {
			slog.Error("could not login", "error", err)
			return err
		}
		token := loginResponse.Result().(*store.Token)
		if token.AccessToken == "" {
			slog.Error("no token returned")
			return err
		}

		// Set the auth token
		r.SetAuthToken(token.AccessToken)

		// Get the work item from the remote database
		slog.Debug("getting work item from remote database", "uuid", item.UUID)
		remoteItem, err := s.GetWorkItem(item.UUID)
		if err != nil {
			slog.Error("could not get work item from remote", "error", err)
			return err
		}

		// Verify the work item status is valid for migration
		if !(remoteItem.Status == store.CLAIMED || remoteItem.Status == store.MIGRATING) {
			slog.Error("remote work item status not valid for migration", "uuid", uuidStr, "status", item.Status)
			return fmt.Errorf("remote work item status not valid for migration: %s, %s", uuidStr, item.Status)
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
			response, err := s.UpdateWorkItem(*item, store.FAILED)
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
		response, err := s.UpdateWorkItem(*item, store.COMPLETED)
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
	err := viper.BindPFlag("migrate-uuid", migrateCmd.Flags().Lookup("uuid"))
	if err != nil {
		slog.Error("unable to bind flag", "error", err)
	}

	rootCmd.AddCommand(migrateCmd)
}
