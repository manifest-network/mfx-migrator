package cmd

import (
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
		// URL is validated in the PersistentPreRunE function
		urlStr := viper.GetString("url")
		uuidStr := viper.GetString("item-uuid")
		force := viper.GetBool("force")
		username := viper.GetString("username")
		password := viper.GetString("password")

		slog.Debug("args", "url", urlStr, "uuid", uuidStr, "force", force, "username", username)

		// Create a new store with the default HTTP client
		url, err := url.Parse(urlStr)
		if err != nil {
			slog.Error("could not parse URL", "error", err)
			return err
		}
		var s *store.Store
		s = store.New(url)
		if username != "" && password != "" {
			token, err := s.Login(username, password)
			if err != nil {
				slog.Error("could not login", "error", err)
				return err
			}
			if token == "" {
				slog.Error("no token returned")
				return err
			}
			s.SetClient(httpclient.NewWithClient(resty.New().SetAuthToken(token)))
		}

		var item *store.WorkItem
		if uuidStr != "" {
			item, err = s.ClaimWorkItemFromUUID(uuid.MustParse(uuidStr), force)
		} else {
			item, err = s.ClaimWorkItemFromQueue()
		}

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

		return nil
	},
}

func init() {
	claimCmd.Flags().BoolP("force", "f", false, "Force re-claiming of a failed work item")
	err := viper.BindPFlag("force", claimCmd.Flags().Lookup("force"))
	if err != nil {
		slog.Error("could not bind flag", "error", err)
	}

	// WARN: Naming this parameter `uuid` seems to cause a conflict with the `uuid` package
	claimCmd.Flags().String("item-uuid", "", "UUID of the work item to claim")
	err = viper.BindPFlag("item-uuid", claimCmd.Flags().Lookup("item-uuid"))
	if err != nil {
		slog.Error("could not bind flag", "error", err)
	}

	claimCmd.Flags().String("username", "", "Username for the remote database")
	err = viper.BindPFlag("username", claimCmd.Flags().Lookup("username"))
	if err != nil {
		slog.Error("could not bind flag", "error", err)
	}

	claimCmd.Flags().String("password", "", "Password for the remote database")
	err = viper.BindPFlag("password", claimCmd.Flags().Lookup("password"))
	if err != nil {
		slog.Error("could not bind flag", "error", err)
	}

	rootCmd.AddCommand(claimCmd)
}
