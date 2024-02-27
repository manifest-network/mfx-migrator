package cmd

import (
	"errors"
	"log/slog"
	"net/url"
	"strconv"
	"time"

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
		urlStr := viper.GetString("url")
		uuidStr := viper.GetString("claim-uuid")
		force := viper.GetBool("force")
		username := viper.GetString("username")
		password := viper.GetString("password")
		neighborhood := viper.GetUint64("neighborhood")

		slog.Debug("args", "url", urlStr, "uuid", uuidStr, "force", force, "username", username, "neighborhood", neighborhood)

		if username == "" || password == "" {
			slog.Error("username and password are required")
			return errors.New("username and password are required")
		}

		// Parse the URL
		url, err := url.Parse(urlStr)
		if err != nil {
			slog.Error("could not parse URL", "error", err)
			return err
		}

		// Retry the claim process 3 times with a 5 seconds wait time between retries and a maximum wait time of 60 seconds.
		// Retry uses an exponential backoff algorithm.
		r := resty.New().
			SetBaseURL(url.String()).
			SetPathParam("neighborhood", strconv.FormatUint(neighborhood, 10)).
			SetRetryCount(3).
			SetRetryWaitTime(5 * time.Second).SetRetryMaxWaitTime(60 * time.Second)
		s := store.NewWithClient(r)

		// Login to the remote database
		slog.Debug("logging in", "username", username, "password", "[REDACTED]")
		response, err := r.R().SetBody(map[string]interface{}{"username": username, "password": password}).SetResult(&store.Token{}).Post("/auth/login")
		if err != nil {
			slog.Error("could not login", "error", err)
			return err
		}

		token := response.Result().(*store.Token)
		if token == nil {
			slog.Error("no token returned")
			return errors.New("no token returned")
		}

		if token.AccessToken == "" {
			slog.Error("empty token returned")
			return errors.New("empty token returned")
		}

		slog.Debug("setting auth token", "token", token.AccessToken)
		// Set the auth token
		r.SetAuthToken(token.AccessToken)

		// Try claiming a work item
		var item *store.WorkItem
		if uuidStr != "" {
			item, err = s.ClaimWorkItemFromUUID(uuid.MustParse(uuidStr), force)
		} else {
			item, err = s.ClaimWorkItemFromQueue()
		}

		// An error occurred during the claim
		if err != nil {
			slog.Error("could not claim work item", "error", err)
			return err
		}

		// If we have a work item, save it to the local state
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

	rootCmd.AddCommand(claimCmd)
}
