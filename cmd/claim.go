package cmd

import (
	"log/slog"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/liftedinit/mfx-migrator/internal/store"
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
		url := viper.GetString("url")
		uuidStr := viper.GetString("uuid")
		force := viper.GetBool("force")

		store := store.NewRemoteStore(&store.RouterImpl{})

		var claimed bool
		var err error
		if uuidStr != "" {
			// The user has specified a UUID to claim.
			workItemUUID := uuid.MustParse(uuidStr)
			claimed, err = store.ClaimWorkItemFromUUID(url, workItemUUID, force)
		} else {
			// The user has not specified a UUID to claim.
			// Claim the first available work item.
			claimed, err = store.ClaimWorkItemFromQueue(url)
		}

		if err != nil {
			slog.Error("could not claim work item", "error", err)
			return err
		}

		if !claimed {
			slog.Info("no work items available")
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
