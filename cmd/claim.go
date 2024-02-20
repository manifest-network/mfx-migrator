package cmd

import (
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/liftedinit/mfx-migrator/internal/state"
	"github.com/liftedinit/mfx-migrator/internal/store"
	"github.com/liftedinit/mfx-migrator/internal/utils"
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
		// 1. Claim a work item from the database.
		//    1.1. Claimed work items should be marked as `claimed` in the database.
		// 	  1.2. Trying to claim a work item that is already claimed should return an error.
		//    1.3. Trying to claim a work item that is already completed should return an error.
		//    1.4. Trying to claim a work item that is already failed should return an error, unless a `force` flag is set.
		// 2. If no work items are available, exit
		// 3. If a work item is claimed, create a `*.uuidStr` file containing state information
		// 4. Exit

		// URL is validated in the PersistentPreRunE function
		url := viper.GetString("url")
		uuidStr := viper.GetString("uuid")
		force := viper.GetBool("force")

		// Validate UUID
		if uuidStr != "" && !utils.IsValidUUID(uuidStr) {
			slog.Error("invalid uuid", "uuid", uuidStr)
			return fmt.Errorf("invalid uuid: %s", uuidStr)
		}

		var claimed bool
		var err error

		if uuidStr != "" {
			// Try to claim the work item with the given UUID
			// If the UUID is not found, return an error
			// If the UUID is already claimed, return an error
			// If the UUID is already completed, return an error
			// If the UUID is already failed, return an error, unless the `-f` flag is set
			claimed, err = claimRemoteWorkItemFromUUID(url, uuidStr, force)
		} else {
			// Try to claim a work item from the database
			claimed, err = claimWorkItem(url)
		}

		if err != nil {
			slog.Error("could not claim work item", "error", err)
			return err
		}

		if claimed == false {
			slog.Info("no work items available")
		}

		return nil
	},
}

func init() {
	claimCmd.Flags().BoolP("force", "f", false, "Force re-claiming of a failed work item")
	viper.BindPFlag("force", claimCmd.Flags().Lookup("force"))
	claimCmd.Flags().String("uuid", "", "UUID of the work item to claim")
	viper.BindPFlag("uuid", claimCmd.Flags().Lookup("uuid"))

	rootCmd.AddCommand(claimCmd)
}

func claimRemoteWorkItem() bool {
	return rand.Intn(2) == 1
}

func claimRemoteWorkItemFromUUID(url string, uuid string, force bool) (bool, error) {
	return rand.Intn(2) == 1, nil
}

func claimWorkItem(url string) (bool, error) {
	workItems, err := store.GetAllWorkItems(url)
	if err != nil {
		slog.Error("could not get work", "error", err)
		return false, err
	}

	for _, workItem := range workItems.Items {
		// If the work item is not in the correct state, skip it
		if workItem.Status != state.CREATED {
			slog.Warn("work item is not in the correct state", "uuid", workItem.UUID, "status", workItem.Status)
		}

		s := state.NewState(workItem.UUID, state.CLAIMING, time.Now())
		if err := s.Save(); err != nil {
			slog.Error("could not save state", "error", err)
			return false, err
		}

		// TODO: Try to claim the work item on the real DB
		isClaimedRemotely := claimRemoteWorkItem()

		if isClaimedRemotely {
			// If the work item was claimed successfully on the DB, mark it as claimed in the local state
			if err = s.Update(state.CLAIMED, time.Now()); err != nil {
				slog.Error("could not update state", "error", err)
				return false, err
			}

			slog.Info("work item claimed", "uuid", workItem.UUID)
			return true, nil
		}

		if err = s.Delete(); err != nil {
			slog.Error("could not delete state", "error", err)
			return false, err
		}
	}

	return false, nil
}
