package main

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
	"github.com/sourcegraph/sourcegraph/schema"
)

const (
	perforceRepoName = "perforce/test-perms"
	aliceEmail       = "alice@perforce.sgdev.org"
	aliceUsername    = "alice"
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

	t.Run("file list excludes excluded files", func(t *testing.T) {
		files, err := userClient.GitListFilenames(repoName, "master")
		if err != nil {
			t.Fatal(err)
		}

		// Notice that Security/hack.sh is excluded
		wantFiles := []string{
			"Backend/main.go",
			"Frontend/app.ts",
			"README.md",
		}

		if diff := cmp.Diff(wantFiles, files); diff != "" {
			t.Fatalf("fileNames mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestSubRepoPermissionsSearch(t *testing.T) {
	// context: https://sourcegraph.slack.com/archives/C07KZF47K/p1658178309055259
	// But it seems that there is still an issue with P4 and they're currently timing out.
	// cc @mollylogue
	t.Skip("Currently broken")
	checkPerforceEnvironment(t)
	enableSubRepoPermissions(t)
	createPerforceExternalService(t)
	userClient, _ := createTestUserAndWaitForRepo(t)

	err := client.WaitForReposToBeIndexed(perforceRepoName)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name          string
		query         string
		zeroResult    bool
		minMatchCount int64
	}{
		{
			name:          "indexed search, nonzero result",
			query:         `index:only This depot is used to test`,
			minMatchCount: 1,
		},
		{
			name:          "unindexed multiline search, nonzero result",
			query:         `index:no This depot is used to test`,
			minMatchCount: 1,
		},
		{
			name:       "indexed search of restricted content",
			query:      `index:only uploading your secrets`,
			zeroResult: true,
		},
		{
			name:       "unindexed search of restricted content",
			query:      `index:no uploading your secrets`,
			zeroResult: true,
		},
		{
			name:       "structural, indexed search of restricted content",
			query:      `repo:^perforce/test-perms$ echo "..." index:only patterntype:structural`,
			zeroResult: true,
		},
		{
			name:       "structural, unindexed search of restricted content",
			query:      `repo:^perforce/test-perms$ echo "..." index:no patterntype:structural`,
			zeroResult: true,
		},
		{
			name:          "structural, indexed search, nonzero result",
			query:         `println(...) index:only patterntype:structural`,
			minMatchCount: 1,
		},
		{
			name:          "structural, unindexed search, nonzero result",
			query:         `println(...) index:no patterntype:structural`,
			minMatchCount: 1,
		},
		{
			name:          "filename search, nonzero result",
			query:         `repo:^perforce/test-perms$ type:path app`,
			minMatchCount: 1,
		},
		{
			name:       "filename search of restricted content",
			query:      `repo:^perforce/test-perms$ type:path hack`,
			zeroResult: true,
		},
		{
			name:          "content search, nonzero result",
			query:         `repo:^perforce/test-perms$ type:file let`,
			minMatchCount: 1,
		},
		{
			name:       "content search of restricted content",
			query:      `repo:^perforce/test-perms$ type:file echo`,
			zeroResult: true,
		},
		{
			name:          "diff search, nonzero result",
			query:         `repo:^perforce/test-perms$ type:diff let`,
			minMatchCount: 1,
		},
		{
			name:       "diff search of restricted content",
			query:      `repo:^perforce/test-perms$ type:diff echo`,
			zeroResult: true,
		},
		{
			name:          "symbol search, nonzero result",
			query:         `repo:^perforce/test-perms$ type:symbol main`,
			minMatchCount: 1,
		},
		{
			name:       "symbol search of restricted content",
			query:      `repo:^perforce/test-perms$ type:symbol asdf`,
			zeroResult: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			results, err := userClient.SearchFiles(test.query)
			if err != nil {
				t.Fatal(err)
			}

			if test.zeroResult {
				if len(results.Results) > 0 {
					t.Fatalf("Want zero result but got %d", len(results.Results))
				}
			} else {
				if len(results.Results) == 0 {
					t.Fatal("Want non-zero results but got 0")
				}
			}

			if results.MatchCount < test.minMatchCount {
				t.Fatalf("Want at least %d match count but got %d", test.minMatchCount, results.MatchCount)
			}
		})
	}

	t.Run("commit search", func(t *testing.T) {
		results, err := userClient.SearchCommits(`repo:^perforce/test-perms$ type:commit`)
		if err != nil {
			t.Fatal(err)
		}
		// Alice should have access to only 1 commit at the moment (other 2 commits modify hack.sh which is
		// inaccessible for Alice)
		// TODO: Alice now has access to 2 commits with recent code changes to update the filtering of commits when sub-repo perms are enabled
		commitsNumber := len(results.Results)
		if commitsNumber != 2 {
			t.Fatalf("Should have access to 2 commits but got %d", commitsNumber)
		}
	})

	commitAccessTests := []struct {
		name      string
		revision  string
		hasAccess bool
	}{
		{
			// I'm not seeing this commit at all in the commit history. Is it just a commit that doesn't exist at all?
			name:     "direct access to inaccessible commit",
			revision: "87440329a7bae580b90280aaaafdc14ee7c1f8ef",
		},
		{
			name:      "direct access to accessible commit",
			revision:  "36d7eda16b9a881ef153126a4036efc4f6afb0c1",
			hasAccess: true,
		},
		// TODO: this commit now can be accessed since we've updated our handling of commit filtering to show commits that modify _any_ file the user has access to
		// we just filter out the files the user doesn't have access to when showing the diff. The todo is to add a new commit that modifies _only_ a file that the user
		// doesn't have access to, and add that here instead.
		//{
		//	name:     "direct access to inaccessible commit-2",
		//	revision: "d9d835aa4b08e1dcb06a21a6dffe6e44f0a141d1",
		//},
	}

	for _, test := range commitAccessTests {
		t.Run(test.name, func(t *testing.T) {
			_, err := userClient.GitGetCommitMessage(perforceRepoName, test.revision)
			if err != nil {
				if test.hasAccess {
					t.Fatal(err)
				}
			} else {
				if !test.hasAccess {
					t.Fatal("No error during accessing restricted commit")
				}
			}
		})
	}

	t.Run("archive repo", func(t *testing.T) {
		url := fmt.Sprintf("%s/%s/-/raw/", *baseURL, perforceRepoName)
		response, err := userClient.GetWithHeaders(url, map[string][]string{"Accept": {"application/zip"}})
		if err != nil {
			t.Fatal(err)
		}
		if response.StatusCode == http.StatusOK {
			t.Fatalf("Should not be able to get an archive of repo with enabled sub-repo perms")
		}
	})

	t.Run("code intel search", func(t *testing.T) {
		result, err := userClient.SearchFiles("context:global \\bhack1337\\b type:file patternType:regexp count:500 case:yes file:\\.(go)$ repo:^perforce/test-perms$@8574314b8de445ec652cab87cbaa1a8dbe6ba6c4")
		if err != nil {
			t.Fatal(err)
		}
		for _, file := range result.Results {
			if strings.HasPrefix(file.File.Name, "hack") {
				t.Fatal("Should not find references for restricted files")
			}
		}
	})
}

func createTestUserAndWaitForRepo(t *testing.T) (*gqltestutil.Client, string) {
	t.Helper()

	// We need to create the `alice` user with a specific e-mail address. This user is
	// configured on our dogfood perforce instance with limited access to the
	// test-perms depot.
	// Alice has access to root, Backend and Frontend directories. (there are .md, .ts and .go files)
	// Alice doesn't have access to Security directory. (there is a .sh file)
	alicePassword := "alicessupersecurepassword"
	t.Log("Creating Alice")
	userClient, err := gqltestutil.SignUpOrSignIn(*baseURL, aliceEmail, aliceUsername, alicePassword)
	if err != nil {
		t.Fatal(err)
	}

	aliceID := userClient.AuthenticatedUserID()
	t.Cleanup(func() {
		if err := client.DeleteUser(aliceID, true); err != nil {
			t.Fatal(err)
		}
	})

	if err := client.SetUserEmailVerified(aliceID, aliceEmail, true); err != nil {
		t.Fatal(err)
	}

	err = userClient.WaitForReposToBeCloned(perforceRepoName)
	if err != nil {
		t.Fatal(err)
	}

	syncUserPerms(t, aliceID, aliceUsername)
	return userClient, perforceRepoName
}

func syncUserPerms(t *testing.T, userID, userName string) {
	t.Helper()
	err := client.ScheduleUserPermissionsSync(userID)
	if err != nil {
		t.Fatal(err)
	}

	// Wait up to 30 seconds for the user to have permissions synced
	// from the code host at least once.
	err = gqltestutil.Retry(30*time.Second, func() error {
		userPermsInfo, err := client.UserPermissionsInfo(userName)
		if err != nil {
			t.Fatal(err)
		}
		if userPermsInfo != nil && !userPermsInfo.SyncedAt.IsZero() {
			return nil
		}
		return gqltestutil.ErrContinueRetry
	})
	if err != nil {
		t.Fatal("Waiting for user permissions to be synced:", err)
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
