package authtest

import (
	"bytes"
	"io"
	"net/http"
	"testing"
	"time"

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
		err = gqltestutil.Retry(5*time.Second, func() error {
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

	t.Run("executor endpoints", func(t *testing.T) {
		resp, err := userClient.Get(*baseURL + "/.executors/")
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf(`Want status code %d error but got %d`, http.StatusUnauthorized, resp.StatusCode)
		}
	})
}
