package cmd_test

import (
	"context"
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
	urlArg := append(slice, []string{"--url", testutils.RootUrl}...)
	usernameArg := append(urlArg, []string{"--username", "user"}...)
	passwordArg := append(usernameArg, []string{"--password", "pass"}...)
	neighborhoodArg := append(passwordArg, []string{"--neighborhood", "1"}...)

	tt := []struct {
		name      string
		args      []string
		err       string
		expected  string
		endpoints []testutils.HttpResponder
	}{
		{name: "no argument", args: []string{}, err: "URL cannot be empty"},
		{name: "username missing", args: urlArg, err: "username is required"},
		{name: "password missing", args: usernameArg, err: "password is required"},
		{name: "claim from queue (default neighborhood)", args: passwordArg, endpoints: []testutils.HttpResponder{
			{Method: "POST", Url: testutils.LoginUrl, Responder: testutils.AuthResponder},
			{Method: "GET", Url: testutils.DefaultMigrationsUrl, Responder: testutils.MustAllMigrationsGetResponder(1, store.CREATED)},
			{Method: "GET", Url: "=~^" + testutils.DefaultMigrationUrl, Responder: testutils.MustMigrationGetResponder(store.CREATED)},
			{Method: "PUT", Url: "=~^" + testutils.DefaultMigrationUrl, Responder: testutils.MigrationUpdateResponder},
		}, expected: "Work item claimed"},
		{name: "claim from queue (neighborhood == 1)", args: neighborhoodArg, endpoints: []testutils.HttpResponder{
			{Method: "POST", Url: testutils.LoginUrl, Responder: testutils.AuthResponder},
			{Method: "GET", Url: testutils.MigrationsUrl, Responder: testutils.MustAllMigrationsGetResponder(1, store.CREATED)},
			{Method: "GET", Url: "=~^" + testutils.MigrationUrl, Responder: testutils.MustMigrationGetResponder(store.CREATED)},
			{Method: "PUT", Url: "=~^" + testutils.MigrationUrl, Responder: testutils.MigrationUpdateResponder},
		}, expected: "Work item claimed"},
		{name: "auth endpoint not found", args: neighborhoodArg, endpoints: []testutils.HttpResponder{
			{Method: "POST", Url: testutils.LoginUrl, Responder: testutils.NotFoundResponder},
		}, err: "response status code: 404"},
		{name: "unable to claim from queue (invalid state)", args: neighborhoodArg, endpoints: []testutils.HttpResponder{
			{Method: "POST", Url: testutils.LoginUrl, Responder: testutils.AuthResponder},
			{Method: "GET", Url: testutils.MigrationsUrl, Responder: testutils.MustAllMigrationsGetResponder(1, store.CLAIMED)},
			{Method: "GET", Url: "=~^" + testutils.MigrationUrl, Responder: testutils.MustMigrationGetResponder(store.CLAIMED)},
		}, expected: "invalid state"},
		{name: "unable to claim from queue (no work items available)", args: neighborhoodArg, endpoints: []testutils.HttpResponder{
			{Method: "POST", Url: testutils.LoginUrl, Responder: testutils.AuthResponder},
			{Method: "GET", Url: testutils.MigrationsUrl, Responder: testutils.MustAllMigrationsGetResponder(0, store.CREATED)},
		}, expected: "No work items available"},
		{name: "claim from uuid", args: append(neighborhoodArg, []string{"--uuid", testutils.DummyUUIDStr}...), endpoints: []testutils.HttpResponder{
			{Method: "POST", Url: testutils.LoginUrl, Responder: testutils.AuthResponder},
			{Method: "GET", Url: "=~^" + testutils.MigrationUrl, Responder: testutils.MustMigrationGetResponder(store.CREATED)},
			{Method: "PUT", Url: "=~^" + testutils.MigrationUrl, Responder: testutils.MigrationUpdateResponder},
		}, expected: "Work item claimed"},
		{name: "unable to claim from uuid (invalid state)", args: append(neighborhoodArg, []string{"--uuid", testutils.DummyUUIDStr}...), endpoints: []testutils.HttpResponder{
			{Method: "POST", Url: testutils.LoginUrl, Responder: testutils.AuthResponder},
			{Method: "GET", Url: "=~^" + testutils.MigrationUrl, Responder: testutils.MustMigrationGetResponder(store.CLAIMED)},
		}, err: "invalid state"},
		{name: "unable to claim from uuid (not found)", args: append(neighborhoodArg, []string{"--uuid", testutils.DummyUUIDStr}...), endpoints: []testutils.HttpResponder{
			{Method: "POST", Url: testutils.LoginUrl, Responder: testutils.AuthResponder},
			{Method: "GET", Url: "=~^" + testutils.MigrationUrl, Responder: testutils.NotFoundResponder},
		}, err: "response status code: 404"},
		{name: "force claim from uuid", args: append(neighborhoodArg, []string{"--uuid", testutils.DummyUUIDStr, "--force"}...), endpoints: []testutils.HttpResponder{
			{Method: "POST", Url: testutils.LoginUrl, Responder: testutils.AuthResponder},
			{Method: "GET", Url: "=~^" + testutils.MigrationUrl, Responder: testutils.MustMigrationGetResponder(store.FAILED)},
			{Method: "PUT", Url: "=~^" + testutils.MigrationUrl, Responder: testutils.MigrationUpdateResponder},
		}, expected: "forcing re-claim of work item"},
	}
	for _, tc := range tt {
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

		t.Run(tc.name, func(t *testing.T) {
			for _, endpoint := range tc.endpoints {
				httpmock.RegisterResponder(endpoint.Method, endpoint.Url, endpoint.Responder)
			}

			out, err := testutils.Execute(t, command, tc.args...)
			t.Log(out)

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
