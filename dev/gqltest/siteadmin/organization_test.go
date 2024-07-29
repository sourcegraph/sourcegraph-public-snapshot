package main

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/dev/gqltest"
	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestOrganization(t *testing.T) {
	const testOrgName = "gqltest-org"
	orgID, err := gqltest.Client.CreateOrganization(testOrgName, testOrgName)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := gqltest.Client.DeleteOrganization(orgID)
		if err != nil {
			t.Fatal(err)
		}
	}()

	t.Run("settings cascade", func(t *testing.T) {
		err := gqltest.Client.OverwriteSettings(orgID, `{"quicklinks":[{"name":"Test quicklink","url":"http://test-quicklink.local"}]}`)
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			err := gqltest.Client.OverwriteSettings(orgID, `{}`)
			if err != nil {
				t.Fatal(err)
			}
		}()

		{
			contents, err := gqltest.Client.ViewerSettings()
			if err != nil {
				t.Fatal(err)
			}

			var got struct {
				QuickLinks []schema.QuickLink `json:"quicklinks"`
			}
			err = jsoniter.UnmarshalFromString(contents, &got)
			if err != nil {
				t.Fatal(err)
			}

			wantQuickLinks := []schema.QuickLink{
				{
					Name: "Test quicklink",
					Url:  "http://test-quicklink.local",
				},
			}
			if diff := cmp.Diff(wantQuickLinks, got.QuickLinks); diff != "" {
				t.Fatalf("QuickLinks mismatch (-want +got):\n%s", diff)
			}
		}
		// Removing all members from an organization is not allowed - add a new user to the organization to verify cascading settings below
		testUserID, err := createOrganizationUser(orgID)
		if err != nil {
			t.Fatal(err)
		}
		gqltest.RemoveTestUserAfterTest(t, testUserID)
		// Remove authenticate user (gqltest-admin) from organization (gqltest-org) should
		// no longer get cascaded settings from this organization.
		err = gqltest.Client.RemoveUserFromOrganization(gqltest.Client.AuthenticatedUserID(), orgID)
		if err != nil {
			t.Fatal(err)
		}

		{
			contents, err := gqltest.Client.ViewerSettings()
			if err != nil {
				t.Fatal(err)
			}

			var got struct {
				QuickLinks []schema.QuickLink `json:"quicklinks"`
			}
			err = jsoniter.UnmarshalFromString(contents, &got)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff([]schema.QuickLink(nil), got.QuickLinks); diff != "" {
				t.Fatalf("QuickLinks mismatch (-want +got):\n%s", diff)
			}
		}
	})

	// Docs: https://sourcegraph.com/docs/user/organizations
	t.Run("auth.userOrgMap", func(t *testing.T) {
		// Create a test user (gqltest-org-user-1) without settings "auth.userOrgMap",
		// the user should not be added to the organization (gqltest-org) automatically.
		const testUsername1 = "gqltest-org-user-1"
		testUserID1, err := gqltest.Client.CreateUser(testUsername1, testUsername1+"@sourcegraph.com")
		if err != nil {
			t.Fatal(err)
		}
		gqltest.RemoveTestUserAfterTest(t, testUserID1)

		orgs, err := gqltest.Client.UserOrganizations(testUsername1)
		if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff([]string{}, orgs); diff != "" {
			t.Fatalf("Organizations mismatch (-want +got):\n%s", diff)
		}

		// Update site configuration to set "auth.userOrgMap" which makes the new user join
		// the organization (gqltest-org) automatically.
		reset, err := gqltest.Client.ModifySiteConfiguration(func(siteConfig *schema.SiteConfiguration) {
			siteConfig.AuthUserOrgMap = map[string][]string{"*": {testOrgName}}
		})
		require.NoError(t, err)
		if reset != nil {
			t.Cleanup(func() {
				require.NoError(t, reset())
			})
		}

		var lastOrgs []string
		// Retry because the configuration update endpoint is eventually consistent
		err = gqltestutil.Retry(5*time.Second, func() error {
			// Create another test user (gqltest-org-user-2) and the user should be added to
			// the organization (gqltest-org) automatically.
			const testUsername2 = "gqltest-org-user-2"
			testUserID2, err := gqltest.Client.CreateUser(testUsername2, testUsername2+"@sourcegraph.com")
			if err != nil {
				t.Fatal(err)
			}
			gqltest.RemoveTestUserAfterTest(t, testUserID2)

			orgs, err = gqltest.Client.UserOrganizations(testUsername2)
			if err != nil {
				t.Fatal(err)
			}
			lastOrgs = orgs

			wantOrgs := []string{testOrgName}
			if cmp.Diff(wantOrgs, orgs) != "" {
				return gqltestutil.ErrContinueRetry
			}
			return nil
		})
		if err != nil {
			t.Fatal(err, "lastOrgs:", lastOrgs)
		}
	})
}

func createOrganizationUser(orgID string) (string, error) {
	const tmpUserName = "gqltest-org-user-tmp"
	tmpUserID, err := gqltest.Client.CreateUser(tmpUserName, tmpUserName+"@sourcegraph.com")
	if err != nil {
		return tmpUserID, err
	}

	err = gqltest.Client.AddUserToOrganization(orgID, tmpUserName)
	return tmpUserID, err
}
