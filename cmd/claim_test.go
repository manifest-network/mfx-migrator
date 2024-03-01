package cmd_test

import (
	"context"
	"errors"
	"net/http"
	"os"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/liftedinit/mfx-migrator/cmd"
	"github.com/liftedinit/mfx-migrator/testutils"
)

func TestClaimCmd(t *testing.T) {
	tmpdir := testutils.SetupTmpDir(t)
	defer os.RemoveAll(tmpdir)

	tt := []struct {
		name      string
		args      []string
		err       error
		out       string
		endpoints []testutils.Endpoint
	}{
		{name: "no arg", args: []string{}, err: errors.New("URL cannot be empty")},
		{name: "url arg only", args: []string{"--url", testutils.RootUrl}, err: errors.New("username is required")},
		{name: "url and username", args: []string{"--url", testutils.RootUrl, "--username", "user"}, err: errors.New("password is required")},
		// The default neighborhood value is 0
		{name: "url, username and password", args: []string{"--url", testutils.RootUrl, "--username", "user", "--password", "pass"}, endpoints: []testutils.Endpoint{
			{Method: "POST", Url: testutils.LoginUrl, Data: "testdata/auth-token.json", Code: http.StatusOK},
			{Method: "GET", Url: testutils.DefaultMigrationsUrl, Data: "testdata/work-items.json", Code: http.StatusOK},
			{Method: "GET", Url: "=~^" + testutils.DefaultMigrationUrl, Data: "testdata/work-item.json", Code: http.StatusOK},
			{Method: "PUT", Url: "=~^" + testutils.DefaultMigrationUrl, Data: "testdata/work-item-update-success.json", Code: http.StatusOK},
		}},
		{name: "url, username, password and neighborhood", args: []string{"--url", testutils.RootUrl, "--username", "user", "--password", "pass", "--neighborhood", "1"}, endpoints: []testutils.Endpoint{
			{Method: "POST", Url: testutils.LoginUrl, Data: "testdata/auth-token.json", Code: http.StatusOK},
			{Method: "GET", Url: testutils.MigrationsUrl, Data: "testdata/work-items.json", Code: http.StatusOK},
			{Method: "GET", Url: "=~^" + testutils.MigrationUrl, Data: "testdata/work-item.json", Code: http.StatusOK},
			{Method: "PUT", Url: "=~^" + testutils.MigrationUrl, Data: "testdata/work-item-update-success.json", Code: http.StatusOK},
		}},
	}
	command := &cobra.Command{Use: "claim", PersistentPreRunE: cmd.RootCmdPersistentPreRunE, RunE: cmd.ClaimCmdRunE}

	// Create a new resty client and inject it into the command context
	// TODO: Is there a better way to do this?
	client := resty.New()
	ctx := context.WithValue(context.Background(), cmd.RestyClientKey, client)
	command.SetContext(ctx)

	// Enable http mocking on the resty client
	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()

	cmd.SetupRootCmdFlags(command)
	cmd.SetupClaimCmdFlags(command)

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			for _, endpoint := range tc.endpoints {
				testutils.SetupMockResponder(t, endpoint.Method, endpoint.Url, endpoint.Data, endpoint.Code)
			}

			out, err := testutils.Execute(t, command, tc.args...)

			require.Equal(t, tc.err, err)

			if tc.err == nil {
				require.Equal(t, tc.out, out)
			}
		})
	}
}
