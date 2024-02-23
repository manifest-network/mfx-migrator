package cmd

import (
	"errors"
	"log/slog"

	"github.com/liftedinit/mfx-migrator/internal/localstate"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// verifyCmd represents the verify command
var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify the status of a migration of MFX tokens to the Manifest Ledger",
	RunE: func(cmd *cobra.Command, args []string) error {
		urlStr := viper.GetString("url")
		uuidStr := viper.GetString("verify-uuid")
		if uuidStr == "" {
			slog.Error("uuid is required")
			return errors.New("uuid is required")
		}

		s, err := localstate.LoadState(uuidStr)
		if err != nil {
			slog.Warn("unable to load local state, continuing", "error", err)
		}

		if s != nil {
			slog.Debug("local state loaded", "state", s)
		}

		// Verify the work item on the remote database
		slog.Debug("verifying remote state", "url", urlStr, "uuid", uuidStr)

		return nil
	},
}

func init() {
	verifyCmd.Flags().String("uuid", "", "UUID of the work item to verify")
	err := viper.BindPFlag("verify-uuid", verifyCmd.Flags().Lookup("uuid"))
	if err != nil {
		slog.Error("unable to bind flag", "error", err)
	}

	rootCmd.AddCommand(verifyCmd)
}
