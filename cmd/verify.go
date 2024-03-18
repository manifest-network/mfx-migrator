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
		c := LoadConfigFromCLI("verify-uuid")
		slog.Debug("args", "c", c)
		if err := c.Validate(); err != nil {
			return err
		}

		s, err := store.LoadState(c.UUID)
		if err != nil {
			slog.Warn("unable to load local state, continuing", "warning", err)
		}

		if s != nil {
			slog.Info("Local state item", "item", s)
		}

		// Verify the work item on the remote database
		slog.Debug("verifying remote state", "url", c.Url, "uuid", c.UUID)

		r := CreateRestClient(cmd.Context(), c.Url, c.Neighborhood)

		item, err := store.GetWorkItem(r, uuid.MustParse(c.UUID))
		if err != nil {
			return errors.WithMessage(err, "unable to get work item")
		}

		if item == nil {
			return errors.WithMessage(err, "work item not found")
		}

		slog.Info("Remote state item", "item", item)

		if s != nil {
			slog.Debug("comparing local and remote states", "local", s, "remote", item)
			if item.Equal(*s) {
				slog.Info("Local and remote states match")
			} else {
				slog.Info("Local and remote states do not match")
			}
		}

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
