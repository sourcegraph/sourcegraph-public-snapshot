package authtest

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
)

func TestOrganization(t *testing.T) {
	const testOrgName = "authtest-organization"
	orgID, err := client.CreateOrganization(testOrgName, testOrgName)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := client.DeleteOrganization(orgID)
		if err != nil {
			t.Fatal(err)
		}
	}()

	// Create a test user (authtest-user-organization) which is not a member of
	// "authtest-organization", the user should not have access to any of the
	// organization's resources.
	const testUsername = "authtest-user-organization"
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

	t.Run("view organization", func(t *testing.T) {
		org, err := userClient.Organization(testOrgName)
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(false, org.ViewerIsMember); diff != "" {
			t.Fatalf("Mismatch (-want +got):\n%s", diff)
		}
	})

	tests := []struct {
		name    string
		run     func() error
		wantErr string
	}{
		{
			name: "settings cascade",
			run: func() error {
				_, err := userClient.SettingsCascade(orgID)
				return err
			},
			wantErr: "current user is not an org member",
		},
		{
			name: "overwrite settings",
			run: func() error {
				return userClient.OverwriteSettings(orgID, "test config")
			},
			wantErr: "current user is not an org member",
		},
		{
			name: "update organization",
			run: func() error {
				return userClient.UpdateOrganization(orgID, "new name")
			},
			wantErr: "current user is not an org member",
		},
		{
			name: "invite user to organization",
			run: func() error {
				_, err := userClient.InviteUserToOrganization(orgID, testUsername)
				return err
			},
			wantErr: "current user is not an org member",
		},
		{
			name: "remove user from organization",
			run: func() error {
				return userClient.RemoveUserFromOrganization(userClient.AuthenticatedUserID(), orgID)
			},
			wantErr: "current user is not an org member",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := fmt.Sprintf("%v", test.run())
			if !strings.Contains(got, test.wantErr) {
				t.Fatalf("Want %q but got %q", test.wantErr, got)
			}
		})
	}
}
