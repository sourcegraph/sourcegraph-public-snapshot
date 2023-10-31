package main

import (
	"strings"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestSiteConfig(t *testing.T) {
	t.Skip("for now")
	t.Run("builtin auth provider: allowSignup", func(t *testing.T) {
		// Sign up a new user is allowed by default.
		const testUsername1 = "gqltest-auth-user-1"
		testClient1, err := gqltestutil.SignUp(*baseURL, testUsername1+"@sourcegraph.com", testUsername1, "mysecurepassword")
		if err != nil {
			t.Fatal(err)
		}
		removeTestUserAfterTest(t, testClient1.AuthenticatedUserID())

		// Update site configuration to not allow sign up for builtin auth provider.
		siteConfig, lastID, err := client.SiteConfiguration()
		if err != nil {
			t.Fatal(err)
		}
		oldSiteConfig := new(schema.SiteConfiguration)
		*oldSiteConfig = *siteConfig
		defer func() {
			_, lastID, err := client.SiteConfiguration()
			if err != nil {
				t.Fatal(err)
			}
			err = client.UpdateSiteConfiguration(oldSiteConfig, lastID)
			if err != nil {
				t.Fatal(err)
			}
		}()

		siteConfig.AuthProviders = []schema.AuthProviders{
			{
				Builtin: &schema.BuiltinAuthProvider{
					AllowSignup: false,
					Type:        "builtin",
				},
			},
		}
		err = client.UpdateSiteConfiguration(siteConfig, lastID)
		if err != nil {
			t.Fatal(err)
		}

		// Retry because the configuration update endpoint is eventually consistent
		err = gqltestutil.Retry(5*time.Second, func() error {
			// Sign up a new user should fail.
			const testUsername2 = "gqltest-auth-user-2"
			testClient2, err := gqltestutil.SignUp(*baseURL, testUsername2+"@sourcegraph.com", testUsername2, "mysecurepassword")
			if err != nil {
				if strings.Contains(err.Error(), "Signup is not enabled") {
					return nil
				}
				t.Fatal(err)
			}
			removeTestUserAfterTest(t, testClient2.AuthenticatedUserID())
			return gqltestutil.ErrContinueRetry
		})
		if err != nil {
			t.Fatal(err)
		}
	})
}

func removeTestUserAfterTest(t *testing.T, userID string) {
	t.Helper()
	t.Cleanup(func() {
		if err := client.DeleteUser(userID, true); err != nil {
			t.Fatal(err)
		}
	})
}
