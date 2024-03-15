package cmd_test

import (
	"context"
	"fmt"
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

	tt := []struct {
		name      string
		args      []string
		err       error
		out       string
		endpoints []testutils.HttpResponder
	}{
		{name: "no arg", args: []string{}, err: fmt.Errorf("URL cannot be empty")},
		{name: "U", args: append(slice, urlP...), err: fmt.Errorf("username is required")},
		{name: "UU", args: append(slice, usernameP...), err: fmt.Errorf("password is required")},
		// The default neighborhood value is 0
		{name: "UUP", args: append(slice, passwordP...), endpoints: []testutils.HttpResponder{
			{Method: "POST", Url: testutils.LoginUrl, Responder: testutils.AuthResponder},
			{Method: "GET", Url: testutils.DefaultMigrationsUrl, Responder: testutils.MustAllMigrationsGetResponder(1, store.CREATED)},
			{Method: "GET", Url: "=~^" + testutils.DefaultMigrationUrl, Responder: testutils.MustMigrationGetResponder(store.CREATED)},
			{Method: "PUT", Url: "=~^" + testutils.DefaultMigrationUrl, Responder: testutils.MigrationUpdateResponder},
		}},
		{name: "UUPN", args: append(slice, neighborhoodP...), endpoints: []testutils.HttpResponder{
			{Method: "POST", Url: testutils.LoginUrl, Responder: testutils.AuthResponder},
			{Method: "GET", Url: testutils.MigrationsUrl, Responder: testutils.MustAllMigrationsGetResponder(1, store.CREATED)},
			{Method: "GET", Url: "=~^" + testutils.MigrationUrl, Responder: testutils.MustMigrationGetResponder(store.CREATED)},
			{Method: "PUT", Url: "=~^" + testutils.MigrationUrl, Responder: testutils.MigrationUpdateResponder},
		}},
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

			_, err := testutils.Execute(t, command, tc.args...)

			require.Equal(t, tc.err, err)

			if tc.err == nil {
				require.FileExists(t, workItemPath)
			}
		})
	}
}
