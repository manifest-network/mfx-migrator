package cmd

import (
	"log/slog"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/liftedinit/mfx-migrator/internal/store"
)

// verifyCmd represents the verify command
var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify the status of a migration of MFX tokens to the Manifest Ledger",
	RunE: func(cmd *cobra.Command, args []string) error {
		config := LoadConfigFromCLI("verify-uuid")
		slog.Debug("args", "config", config)
		if err := config.Validate(); err != nil {
			return err
		}

		s, err := store.LoadState(config.UUID)
		if err != nil {
			slog.Warn("unable to load local state, continuing", "warning", err)
		}

		if s != nil {
			slog.Debug("local state loaded", "state", s)
		}

		// Verify the work item on the remote database
		slog.Debug("verifying remote state", "url", config.Url, "uuid", config.UUID)

		r := CreateRestClient(cmd.Context(), config.Url, config.Neighborhood)

		item, err := store.GetWorkItem(r, uuid.MustParse(config.UUID))
		if err != nil {
			return errors.WithMessage(err, "unable to get work item")
		}

		if item == nil {
			return errors.WithMessage(err, "work item not found")
		}

		slog.Info("work item", "item", item)

		return nil
	},
}

func init() {
	verifyCmd.Flags().String("uuid", "", "UUID of the work item to verify")
	if err := verifyCmd.MarkFlagRequired("uuid"); err != nil {
		slog.Error(ErrorMarkingFlagRequired, "error", err)
	}
	err := viper.BindPFlag("verify-uuid", verifyCmd.Flags().Lookup("uuid"))
	if err != nil {
		slog.Error("unable to bind flag", "error", err)
	}

	rootCmd.AddCommand(verifyCmd)
}
