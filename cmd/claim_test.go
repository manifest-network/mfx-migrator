package cmd_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/liftedinit/mfx-migrator/internal/store"

	"github.com/liftedinit/mfx-migrator/cmd"
	"github.com/liftedinit/mfx-migrator/testutils"
)

func TestClaimCmd(t *testing.T) {
	tmpdir := t.TempDir()
	if err := os.Chdir(tmpdir); err != nil {
		t.Fatal(err)
	}

	workItemPath := tmpdir + "/" + testutils.Uuid + ".json"

	var slice []string
	urlP := []string{"--url", testutils.RootUrl}
	usernameP := []string{"--username", "user"}
	passwordP := []string{"--password", "pass"}
	neighborhoodP := []string{"--neighborhood", "1"}

	slog.Default()

	// TODO: Test force claim of a failed item clears the previous error
	tt := []struct {
		name      string
		args      []string
		err       string
		expected  string
		endpoints []testutils.HttpResponder
	}{
		{name: "no argument", args: []string{}, err: "URL cannot be empty"},
		{name: "username missing", args: append(slice, urlP...), err: "username is required"},
		{name: "password missing", args: append(slice, usernameP...), err: "password is required"},
		{name: "claim from queue (default neighborhood)", args: append(slice, passwordP...), endpoints: []testutils.HttpResponder{
			{Method: "POST", Url: testutils.LoginUrl, Responder: testutils.AuthResponder},
			{Method: "GET", Url: testutils.DefaultMigrationsUrl, Responder: testutils.MustAllMigrationsGetResponder(1, store.CREATED)},
			{Method: "GET", Url: "=~^" + testutils.DefaultMigrationUrl, Responder: testutils.MustMigrationGetResponder(store.CREATED)},
			{Method: "PUT", Url: "=~^" + testutils.DefaultMigrationUrl, Responder: testutils.MigrationUpdateResponder},
		}, expected: "Work item claimed"},
		{name: "claim from queue (neighborhood == 1)", args: append(slice, neighborhoodP...), endpoints: []testutils.HttpResponder{
			{Method: "POST", Url: testutils.LoginUrl, Responder: testutils.AuthResponder},
			{Method: "GET", Url: testutils.MigrationsUrl, Responder: testutils.MustAllMigrationsGetResponder(1, store.CREATED)},
			{Method: "GET", Url: "=~^" + testutils.MigrationUrl, Responder: testutils.MustMigrationGetResponder(store.CREATED)},
			{Method: "PUT", Url: "=~^" + testutils.MigrationUrl, Responder: testutils.MigrationUpdateResponder},
		}, expected: "Work item claimed"},
		{name: "auth endpoint not found", args: slice, endpoints: []testutils.HttpResponder{
			{Method: "POST", Url: testutils.LoginUrl, Responder: testutils.NotFoundResponder},
		}, err: "response status code: 404"},
		{name: "unable to claim from queue (invalid state)", args: slice, endpoints: []testutils.HttpResponder{
			{Method: "POST", Url: testutils.LoginUrl, Responder: testutils.AuthResponder},
			{Method: "GET", Url: testutils.MigrationsUrl, Responder: testutils.MustAllMigrationsGetResponder(1, store.CLAIMED)},
			{Method: "GET", Url: "=~^" + testutils.MigrationUrl, Responder: testutils.MustMigrationGetResponder(store.CLAIMED)},
		}, expected: "invalid state"},
	}
	command := &cobra.Command{Use: "claim", PersistentPreRunE: cmd.RootCmdPersistentPreRunE, RunE: cmd.ClaimCmdRunE}

	// Create a new resty client and inject it into the command context
	client := resty.New()
	ctx := context.WithValue(context.Background(), cmd.RestyClientKey, client)
	command.SetContext(ctx)

	// Enable http mocking on the resty client
	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()

	cmd.SetupRootCmdFlags(command)
	cmd.SetupClaimCmdFlags(command)

	for _, tc := range tt {
		slice = append(slice, tc.args...)
		t.Run(tc.name, func(t *testing.T) {
			for _, endpoint := range tc.endpoints {
				httpmock.RegisterResponder(endpoint.Method, endpoint.Url, endpoint.Responder)
			}

			out, err := testutils.Execute(t, command, tc.args...)

			if tc.err == "" {
				require.Contains(t, out, tc.expected)
			} else {
				require.ErrorContains(t, err, tc.err)
			}
		})

		// Remove the work item file if it exists
		_, err := os.Stat(workItemPath)
		if !os.IsNotExist(err) {
			err = os.Remove(workItemPath)
			require.NoError(t, err)
		}
	}
}
