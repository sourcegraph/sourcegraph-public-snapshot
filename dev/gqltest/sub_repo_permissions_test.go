package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestSubRepoPermissionsPerforce(t *testing.T) {
	checkPerforceEnvironment(t)
	enableSubRepoPermissions(t)
	createPerforceExternalService(t)
	userClient, repoName := createTestUserAndWaitForRepo(t)

	// Test cases

	t.Run("can read README.md", func(t *testing.T) {
		blob, err := userClient.GitBlob(repoName, "master", "README.md")
		if err != nil {
			t.Fatal(err)
		}
		wantBlob := `This depot is used to test user and group permissions.
`
		if diff := cmp.Diff(wantBlob, blob); diff != "" {
			t.Fatalf("Blob mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("cannot read hack.sh", func(t *testing.T) {
		// Should not be able to read hack.sh
		blob, err := userClient.GitBlob(repoName, "master", "Security/hack.sh")
		if err != nil {
			t.Fatal(err)
		}

		// This is the desired behaviour at the moment, see where we check for
		// os.IsNotExist error in GitCommitResolver.Blob
		wantBlob := ``

		if diff := cmp.Diff(wantBlob, blob); diff != "" {
			t.Fatalf("Blob mismatch (-want +got):\n%s", diff)
		}
	})
}

func createTestUserAndWaitForRepo(t *testing.T) (*gqltestutil.Client, string) {
	t.Helper()

	// We need to creat the `alice` user with a specific e-mail address. This user is
	// configured on our dogfood perforce instance with limited access to the
	// test-perms depot.
	alicePassword := "alicessupersecurepassword"
	t.Log("Creating Alice")
	userClient, err := gqltestutil.SignUp(*baseURL, "alice@perforce.sgdev.org", "alice", alicePassword)
	if err != nil {
		t.Fatal(err)
	}

	aliceID := userClient.AuthenticatedUserID()
	t.Cleanup(func() {
		if err := client.DeleteUser(aliceID, true); err != nil {
			t.Fatal(err)
		}
	})

	if err := client.SetUserEmailVerified(aliceID, "alice@perforce.sgdev.org", true); err != nil {
		t.Fatal(err)
	}

	const repoName = "perforce/test-perms"
	t.Logf("Waiting for %q as Alice", repoName)
	err = userClient.WaitForReposToBeCloned(repoName)
	if err != nil {
		t.Fatal(err)
	}
	return userClient, repoName
}

func enableSubRepoPermissions(t *testing.T) {
	t.Helper()

	siteConfig, err := client.SiteConfiguration()
	if err != nil {
		t.Fatal(err)
	}
	oldSiteConfig := new(schema.SiteConfiguration)
	*oldSiteConfig = *siteConfig
	t.Cleanup(func() {
		err = client.UpdateSiteConfiguration(oldSiteConfig)
		if err != nil {
			t.Fatal(err)
		}
	})

	siteConfig.ExperimentalFeatures = &schema.ExperimentalFeatures{
		Perforce: "enabled",
		SubRepoPermissions: &schema.SubRepoPermissions{
			Enabled: true,
		},
	}
	err = client.UpdateSiteConfiguration(siteConfig)
	if err != nil {
		t.Fatal(err)
	}
}
