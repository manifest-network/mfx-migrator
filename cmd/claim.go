package cmd

import (
	"log/slog"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/pkg/errors"
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
	RunE: ClaimCmdRunE,
}

func ClaimCmdRunE(cmd *cobra.Command, args []string) error {
	config := LoadConfigFromCLI("claim-uuid")
	slog.Debug("args", "config", config)
	if err := config.Validate(); err != nil {
		return err
	}

	claimConfig := LoadClaimConfigFromCLI()
	slog.Debug("args", "claim-config", claimConfig)

	authConfig := LoadAuthConfigFromCLI()
	slog.Debug("args", "auth-config", authConfig)
	if err := authConfig.Validate(); err != nil {
		return err
	}

	r := CreateRestClient(cmd.Context(), config.Url, config.Neighborhood)
	if err := AuthenticateRestClient(r, authConfig.Username, authConfig.Password); err != nil {
		return err
	}

	item, err := claimWorkItem(r, config.UUID, claimConfig.Force)
	if err != nil {
		return err
	}

	if item == nil {
		slog.Info("No work items available")
	}

	return nil
}

func init() {
	SetupClaimCmdFlags(claimCmd)
	rootCmd.AddCommand(claimCmd)
}

func SetupClaimCmdFlags(command *cobra.Command) {
	command.Flags().BoolP("force", "f", false, "Force re-claiming of a failed work item")
	if err := viper.BindPFlag("force", command.Flags().Lookup("force")); err != nil {
		slog.Error(ErrorBindingFlag, "error", err)
	}

	command.Flags().String("uuid", "", "UUID of the work item to claim")
	if err := viper.BindPFlag("claim-uuid", command.Flags().Lookup("uuid")); err != nil {
		slog.Error(ErrorBindingFlag, "error", err)
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
		return nil, errors.WithMessage(err, "could not claim work item")
	}

	if item != nil {
		slog.Info("Work item claimed", "uuid", item.UUID)
	}

	return item, nil
}
