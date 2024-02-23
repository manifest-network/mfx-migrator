package cmd

import (
	"errors"
	"log/slog"
	"net/url"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/liftedinit/mfx-migrator/internal/httpclient"
	"github.com/liftedinit/mfx-migrator/internal/localstate"
	"github.com/liftedinit/mfx-migrator/internal/store"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// claimCmd represents the claim command
var claimCmd = &cobra.Command{
	Use:   "claim",
	Short: "Claim a work item from the database.",
	Long: `The claim command should be used to claim a work item from the database.

If no work items are available, the command should exit.
Claimed work items should be marked as 'claimed'' in the database.

Trying to claim a work item that is already claimed should return an error.
Trying to claim a work item that is already completed should return an error.
Trying to claim a work item that is already failed should return an error, unless the '-f' flag is set.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		urlStr := viper.GetString("url")
		uuidStr := viper.GetString("claim-uuid")
		force := viper.GetBool("force")
		username := viper.GetString("username")
		password := viper.GetString("password")

		slog.Debug("args", "url", urlStr, "uuid", uuidStr, "force", force, "username", username)

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

		// Create a new store with the default HTTP client
		s := store.New(url)

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
		s.SetClient(httpclient.NewWithClient(resty.New().SetAuthToken(token)))

		// Try claiming a work item
		var item *store.WorkItem
		if uuidStr != "" {
			item, err = s.ClaimWorkItemFromUUID(uuid.MustParse(uuidStr), force)
		} else {
			item, err = s.ClaimWorkItemFromQueue()
		}

		// An error occurred during the claim
		if err != nil {
			slog.Error("could not claim work item", "error", err)
			return err
		}

		// If we have a work item, save it to the local state
		if item != nil {
			err = localstate.SaveState(item)
			if err != nil {
				slog.Error("could not save state", "error", err)
				return err
			}
		}

		// At this point, no work items are available
		return nil
	},
}

func init() {
	claimCmd.Flags().BoolP("force", "f", false, "Force re-claiming of a failed work item")
	err := viper.BindPFlag("force", claimCmd.Flags().Lookup("force"))
	if err != nil {
		slog.Error("could not bind flag", "error", err)
	}

	claimCmd.Flags().String("uuid", "", "UUID of the work item to claim")
	err = viper.BindPFlag("claim-uuid", claimCmd.Flags().Lookup("uuid"))
	if err != nil {
		slog.Error("could not bind flag", "error", err)
	}

	rootCmd.AddCommand(claimCmd)
}
