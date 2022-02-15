package main

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestSubRepoPermissionsPerforce(t *testing.T) {
	checkPerforceEnvironment(t)
	enableSubRepoPermissions(t)
	createPerforceExternalService(t)

	// We need to creat the `alice` user with a specific e-mail address. This user is
	// configured on our dogfood perforce instance with limited access to the
	// test-perms depot.
	alicePassword := "alicessupersecurepassword"
	t.Log("Creating Alice")
	aliceClient, err := gqltestutil.SignUp(*baseURL, "alice@perforce.sgdev.org", "alice", alicePassword)
	if err != nil {
		t.Fatal(err)
	}

	aliceID := aliceClient.AuthenticatedUserID()
	t.Cleanup(func() {
		if err := client.DeleteUser(aliceID, true); err != nil {
			t.Fatal(err)
		}
	})

	if err := client.SetUserEmailVerified(aliceID, "alice@perforce.sgdev.org", true); err != nil {
		t.Fatal(err)
	}

	const repoName = "perforce/test-perms"

	// We have not enabled EnforceAuthzForSiteAdmin so they can see this repo
	t.Logf("Waiting for %q as Admin", repoName)
	err = client.WaitForReposToBeCloned(repoName)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Waiting for %q as Alice", repoName)
	err = aliceClient.WaitForReposToBeCloned(repoName)
	if err != nil {
		t.Fatal(err)
	}

	blob, err := aliceClient.GitBlob(repoName, "master", "README.md")
	if err != nil {
		t.Fatal(err)
	}

	wantBlob := `This depot is used to test user and group permissions.
`
	if diff := cmp.Diff(wantBlob, blob); diff != "" {
		t.Fatalf("Blob mismatch (-want +got):\n%s", diff)
	}

	// Should not be able to read hack.sh
	blob, err = aliceClient.GitBlob(repoName, "master", "Security/hack.sh")
	if err != nil {
		t.Fatal(err)
	}

	// TODO: We probably want an error, not an empty string here
	wantBlob = ``

	if diff := cmp.Diff(wantBlob, blob); diff != "" {
		t.Fatalf("Blob mismatch (-want +got):\n%s", diff)
	}
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

func logCurrentSiteConfig(t *testing.T) {
	t.Helper()

	c, err := client.SiteConfiguration()
	if err != nil {
		t.Fatal(err)
	}

	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf(string(data))
}
