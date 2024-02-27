package cmd

import (
	"log/slog"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/liftedinit/mfx-migrator/internal/localstate"
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
		config := LoadConfigFromCLI("claim-uuid")
		slog.Debug("args", "config", config)

		if err := config.Validate(); err != nil {
			return err
		}

		r := CreateRestClient(config.Url, config.Neighborhood)
		if err := AuthenticateRestClient(r, config.Username, config.Password); err != nil {
			return err
		}

		item, err := claimWorkItem(r, config.UUID, config.Force)
		if err != nil {
			return err
		}

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
	setupFlags()
	rootCmd.AddCommand(claimCmd)
}

func setupFlags() {
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
}

// claimWorkItem claims a work item from the database
func claimWorkItem(r *resty.Client, uuidStr string, force bool) (*store.WorkItem, error) {
	slog.Info("Claiming work item...")
	var err error
	var item *store.WorkItem
	if uuidStr != "" {
		item, err = store.ClaimWorkItemFromUUID(r, uuid.MustParse(uuidStr), force)
	} else {
		item, err = store.ClaimWorkItemFromQueue(r)
	}

	// An error occurred during the claim
	if err != nil {
		slog.Error("could not claim work item", "error", err)
		return nil, err
	}

	slog.Info("work item claimed", "uuid", item.UUID)

	return item, nil
}
