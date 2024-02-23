package cmd

import (
	"log/slog"
	"net/url"

	"github.com/google/uuid"
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
		uuidStr := viper.GetString("uuid")
		force := viper.GetBool("force")

		// Create a new store with the default HTTP client
		url, err := url.Parse(urlStr)
		if err != nil {
			slog.Error("could not parse URL", "error", err)
			return err
		}
		store := store.New(url)
		if uuidStr != "" {
			_, err = store.ClaimWorkItemFromUUID(uuid.MustParse(uuidStr), force)
		} else {
			_, err = store.ClaimWorkItemFromQueue()
		}

		if err != nil {
			slog.Error("could not claim work item", "error", err)
			return err
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
	claimCmd.Flags().String("uuid", "", "UUID of the work item to claim")
	err = viper.BindPFlag("uuid", claimCmd.Flags().Lookup("uuid"))
	if err != nil {
		slog.Error("could not bind flag", "error", err)
	}

	rootCmd.AddCommand(claimCmd)
}
