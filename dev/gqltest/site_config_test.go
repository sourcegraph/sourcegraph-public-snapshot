package main

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestSiteConfig(t *testing.T) {
	t.Run("builtin auth provider: allowSignup", func(t *testing.T) {
		// Sign up a new user is allowed by default.
		const testUsername1 = "gqltest-auth-user-1"
		testClient1, err := gqltestutil.NewClient(*baseURL)
		require.NoError(t, err)
		require.NoError(t, testClient1.SignUp(testUsername1+"@sourcegraph.com", testUsername1, "mysecurepassword"))
		removeTestUserAfterTest(t, testClient1.AuthenticatedUserID())

		// Update site configuration to not allow sign up for builtin auth provider.
		reset, err := client.ModifySiteConfiguration(func(siteConfig *schema.SiteConfiguration) {
			siteConfig.AuthProviders = []schema.AuthProviders{
				{
					Builtin: &schema.BuiltinAuthProvider{
						AllowSignup: false,
						Type:        "builtin",
					},
				},
			}
		})
		require.NoError(t, err)
		if reset != nil {
			t.Cleanup(func() {
				require.NoError(t, reset())
			})
		}

		// Retry because the configuration update endpoint is eventually consistent
		err = gqltestutil.Retry(5*time.Second, func() error {
			// Sign up a new user should fail.
			const testUsername2 = "gqltest-auth-user-2"
			testClient2, err := gqltestutil.NewClient(*baseURL)
			require.NoError(t, err)
			if err := testClient2.SignUp(testUsername2+"@sourcegraph.com", testUsername2, "mysecurepassword"); err != nil {
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
