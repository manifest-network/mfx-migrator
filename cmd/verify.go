package cmd

import (
	"fmt"
	"log/slog"

	"github.com/liftedinit/mfx-migrator/internal/state"
	"github.com/liftedinit/mfx-migrator/internal/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// verifyCmd represents the verify command
var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify the status of a migration of MFX tokens to the Manifest Ledger",
	RunE: func(cmd *cobra.Command, args []string) error {
		url := viper.GetString("url")
		uuidStr := viper.GetString("uuid")

		if uuidStr != "" && !utils.IsValidUUID(uuidStr) {
			slog.Error("invalid uuid", "uuid", uuidStr)
			return fmt.Errorf("invalid uuid: %s", uuidStr)
		}

		s, err := state.LoadState(uuidStr)
		if err != nil {
			slog.Warn("unable to load local state, continuing", "error", err)
		}

		if s != nil {
			slog.Debug("local state loaded", "state", s)
		}

		// Verify the work item on the remote database
		slog.Debug("verifying remote state", "url", url, "uuid", uuidStr)

		return nil
	},
}

func init() {
	verifyCmd.Flags().StringP("uuid", "u", "", "UUID of the MFX migration")
	viper.BindPFlag("uuid", verifyCmd.Flags().Lookup("uuid"))

	rootCmd.AddCommand(verifyCmd)
}
