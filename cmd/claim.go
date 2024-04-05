package cmd

import (
	"log/slog"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/liftedinit/mfx-migrator/internal/config"

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
	c := LoadConfigFromCLI("claim-uuid")
	slog.Debug("args", "c", c)
	if err := c.Validate(); err != nil {
		return err
	}

	claimConfig := LoadClaimConfigFromCLI()
	slog.Debug("args", "claim-c", claimConfig)

	authConfig := LoadAuthConfigFromCLI()
	slog.Debug("args", "auth-c", authConfig)
	if err := authConfig.Validate(); err != nil {
		return err
	}

	r := CreateRestClient(cmd.Context(), c.Url, c.Neighborhood)
	if err := AuthenticateRestClient(r, authConfig.Username, authConfig.Password); err != nil {
		return err
	}

	items, err := claimWorkItem(r, c.UUID, claimConfig)
	if err != nil {
		return err
	}

	if len(items) == 0 {
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
func claimWorkItem(r *resty.Client, uuidStr string, config config.ClaimConfig) ([]*store.WorkItem, error) {
	slog.Info("Claiming work item...")
	var err error
	var items []*store.WorkItem
	if uuidStr != "" {
		var item *store.WorkItem
		item, err = store.ClaimWorkItemFromUUID(r, uuid.MustParse(uuidStr), config.Force)
		if err != nil {
			return nil, errors.WithMessage(err, "could not claim work item")
		}
		items = append(items, item)
	} else {
		items, err = store.ClaimWorkItemFromQueue(r)
		if err != nil {
			return nil, errors.WithMessage(err, "could not claim work item")
		}
	}

	for _, item := range items {
		slog.Info("Work item claimed", "uuid", item.UUID)
	}

	return items, nil
}
