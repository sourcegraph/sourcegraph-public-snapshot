package authtest

import (
	"bytes"
	"io"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestCodeIntelEndpoints(t *testing.T) {

	// Create a test user (authtest-user-code-intel) which is not a site admin, the
	// user should receive access denied for LSIF endpoints of repositories the user
	// does not have access to.
	const testUsername = "authtest-user-code-intel"
	userClient, err := gqltestutil.SignUp(*baseURL, testUsername+"@sourcegraph.com", testUsername, "mysecurepassword")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := client.DeleteUser(userClient.AuthenticatedUserID(), true)
		if err != nil {
			t.Fatal(err)
		}
	}()

	// Set up external service
	esID, err := client.AddExternalService(
		gqltestutil.AddExternalServiceInput{
			Kind:        extsvc.KindGitHub,
			DisplayName: "authtest-github-code-intel-repository",
			Config: mustMarshalJSONString(
				&schema.GitHubConnection{
					Authorization: &schema.GitHubAuthorization{},
					Repos: []string{
						"sgtest/go-diff",
						"sgtest/private", // Private
					},
					RepositoryPathPattern: "github.com/{nameWithOwner}",
					Token:                 *githubToken,
					Url:                   "https://ghe.sgdev.org/",
				},
			),
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := client.DeleteExternalService(esID)
		if err != nil {
			t.Fatal(err)
		}
	}()

	const privateRepo = "github.com/sgtest/private"
	err = client.WaitForReposToBeCloned(
		"github.com/sgtest/go-diff",
		privateRepo,
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("LSIF upload", func(t *testing.T) {
		// Update site configuration to enable "lsifEnforceAuth".
		siteConfig, err := client.SiteConfiguration()
		if err != nil {
			t.Fatal(err)
		}

		oldSiteConfig := new(schema.SiteConfiguration)
		*oldSiteConfig = *siteConfig
		defer func() {
			err = client.UpdateSiteConfiguration(oldSiteConfig)
			if err != nil {
				t.Fatal(err)
			}
		}()

		siteConfig.LsifEnforceAuth = true
		err = client.UpdateSiteConfiguration(siteConfig)
		if err != nil {
			t.Fatal(err)
		}

		// Retry because the configuration update endpoint is eventually consistent
		var lastBody string
		err = gqltestutil.Retry(10*time.Second, func() error {
			resp, err := userClient.Post(*baseURL+"/.api/lsif/upload?commit=6ffc6072f5ed13d8e8782490705d9689cd2c546a&repository=github.com/sgtest/private", nil)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = resp.Body.Close() }()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}

			if bytes.Contains(body, []byte("must provide github_token")) {
				return nil
			}

			lastBody = string(body)
			return gqltestutil.ErrContinueRetry
		})
		if err != nil {
			t.Fatal(err, "lastBody:", lastBody)
		}
	})

	t.Run("executor endpoints (no access token configured)", func(t *testing.T) {
		// Update site configuration to remove any executor access token.
		defer setExecutorAccessToken(t, "")()

		resp, err := userClient.Get(*baseURL + "/.executors/")
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode/100 != 5 {
			t.Fatalf(`Want status code 5xx error but got %d`, resp.StatusCode)
		}
	})

	t.Run("executor endpoints (no access token supplied)", func(t *testing.T) {
		// Update site configuration to set the executor access token.
		defer setExecutorAccessToken(t, "hunter2hunter2hunter2")()

		resp, err := userClient.Get(*baseURL + "/.executors/")
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode/100 != 4 {
			t.Fatalf(`Want status code 4xx error but got %d`, resp.StatusCode)
		}
	})
}

func setExecutorAccessToken(t *testing.T, token string) func() {
	siteConfig, err := client.SiteConfiguration()
	if err != nil {
		t.Fatal(err)
	}

	oldSiteConfig := new(schema.SiteConfiguration)
	*oldSiteConfig = *siteConfig
	siteConfig.ExecutorsAccessToken = token

	if err := client.UpdateSiteConfiguration(siteConfig); err != nil {
		t.Fatal(err)
	}
	return func() {
		if err := client.UpdateSiteConfiguration(oldSiteConfig); err != nil {
			t.Fatal(err)
		}
	}
}
