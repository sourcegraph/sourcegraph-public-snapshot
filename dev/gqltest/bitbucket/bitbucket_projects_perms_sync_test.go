package main

import (
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/dev/gqltest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const (
	projectKey      = "SOURCEGRAPH"
	clonedRepoName1 = "bbs/SOURCEGRAPH/jsonrpc2"
	clonedRepoName2 = "bbs/SOURCEGRAPH/empty-repo-1"
)

func TestBitbucketProjectsPermsSync_SetUnrestrictedPermissions(t *testing.T) {
	if len(*gqltest.BbsURL) == 0 || len(*gqltest.BbsToken) == 0 || len(*gqltest.BbsUsername) == 0 {
		t.Skip("Environment variable BITBUCKET_SERVER_URL, BITBUCKET_SERVER_TOKEN, or BITBUCKET_SERVER_USERNAME is not set")
	}

	// External service setup
	esID, err := setUpExternalService(t)
	if err != nil {
		t.Fatal(err)
	}

	gqltest.RemoveExternalServiceAfterTest(t, esID)

	// Triggering the sync job
	unrestricted := true
	err = gqltest.Client.SetRepositoryPermissionsForBitbucketProject(gqltestutil.BitbucketProjectPermsSyncArgs{
		ProjectKey:      projectKey,
		CodeHost:        esID,
		UserPermissions: make([]types.UserPermission, 0),
		Unrestricted:    &unrestricted,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Wait up to 30 seconds for worker to finish the permissions sync.
	err = waitForSyncJobToFinish()
	if err != nil {
		t.Fatal("Waiting for repository permissions to be synced:", err)
	}

	// Perform the checks
	err = checkRepoPermissions(clonedRepoName1, true)
	if err != nil {
		t.Fatal(err)
	}

	err = checkRepoPermissions(clonedRepoName2, true)
	if err != nil {
		t.Fatal(err)
	}
}

func TestBitbucketProjectsPermsSync_FromRestrictedToUnrestrictedPermissions(t *testing.T) {
	if len(*gqltest.BbsURL) == 0 || len(*gqltest.BbsToken) == 0 || len(*gqltest.BbsUsername) == 0 {
		t.Skip("Environment variable BITBUCKET_SERVER_URL, BITBUCKET_SERVER_TOKEN, or BITBUCKET_SERVER_USERNAME is not set")
	}

	// External service setup
	esID, err := setUpExternalService(t)
	if err != nil {
		t.Fatal(err)
	}

	gqltest.RemoveExternalServiceAfterTest(t, esID)

	// Triggering the sync job to set permissions for existing user
	unrestricted := false
	err = gqltest.Client.SetRepositoryPermissionsForBitbucketProject(gqltestutil.BitbucketProjectPermsSyncArgs{
		ProjectKey:      projectKey,
		CodeHost:        esID,
		UserPermissions: []types.UserPermission{{BindID: "gqltest@sourcegraph.com", Permission: "READ"}},
		Unrestricted:    &unrestricted,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Wait up to 30 seconds for worker to finish the permissions sync.
	err = waitForSyncJobToFinish()
	if err != nil {
		t.Fatal("Waiting for repository permissions to be synced:", err)
	}

	// Checking repo permissions
	err = checkRepoPermissions(clonedRepoName1, false)
	if err != nil {
		t.Fatal(err)
	}

	// Triggering the sync job to set unrestricted permissions
	unrestricted = true
	err = gqltest.Client.SetRepositoryPermissionsForBitbucketProject(gqltestutil.BitbucketProjectPermsSyncArgs{
		ProjectKey:      projectKey,
		CodeHost:        esID,
		UserPermissions: make([]types.UserPermission, 0),
		Unrestricted:    &unrestricted,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Wait up to 30 seconds for worker to finish the permissions sync.
	err = waitForSyncJobToFinish()
	if err != nil {
		t.Fatal("Waiting for repository permissions to be synced:", err)
	}

	// Checking repo permissions
	err = checkRepoPermissions(clonedRepoName1, true)
	if err != nil {
		t.Fatal(err)
	}
}

func TestBitbucketProjectsPermsSync_SetPendingPermissions_NonExistentUsersOnly(t *testing.T) {
	if len(*gqltest.BbsURL) == 0 || len(*gqltest.BbsToken) == 0 || len(*gqltest.BbsUsername) == 0 {
		t.Skip("Environment variable BITBUCKET_SERVER_URL, BITBUCKET_SERVER_TOKEN, or BITBUCKET_SERVER_USERNAME is not set")
	}

	// External service setup
	esID, err := setUpExternalService(t)
	if err != nil {
		t.Fatal(err)
	}

	gqltest.RemoveExternalServiceAfterTest(t, esID)

	// Triggering the sync job
	unrestricted := false
	err = gqltest.Client.SetRepositoryPermissionsForBitbucketProject(gqltestutil.BitbucketProjectPermsSyncArgs{
		ProjectKey: projectKey,
		CodeHost:   esID,
		UserPermissions: []types.UserPermission{
			{
				BindID:     "some-user-1@domain.com",
				Permission: "READ",
			},
			{
				BindID:     "some-user-2@domain.com",
				Permission: "READ",
			},
			{
				BindID:     "some-user-3@domain.com",
				Permission: "READ",
			},
		},
		Unrestricted: &unrestricted,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Wait up to 30 seconds for worker to finish the permissions sync.
	err = waitForSyncJobToFinish()
	if err != nil {
		t.Fatal("Waiting for repository permissions to be synced:", err)
	}

	// Perform the checks
	pendingPerms, err := gqltest.Client.UsersWithPendingPermissions()
	if err != nil {
		t.Fatal(err)
	}

	if len(pendingPerms) != 3 {
		t.Fatalf("Expected 3 pending permissions entries, got: %d", len(pendingPerms))
	}

	want := []string{
		"some-user-1@domain.com",
		"some-user-2@domain.com",
		"some-user-3@domain.com",
	}

	if diff := cmp.Diff(want, pendingPerms); diff != "" {
		t.Fatalf("Pending permissions mismatch (-want +got):\n%s", diff)
	}
}

func TestBitbucketProjectsPermsSync_SetPendingPermissions_ExistentAndNonExistentUsers(t *testing.T) {
	if len(*gqltest.BbsURL) == 0 || len(*gqltest.BbsToken) == 0 || len(*gqltest.BbsUsername) == 0 {
		t.Skip("Environment variable BITBUCKET_SERVER_URL, BITBUCKET_SERVER_TOKEN, or BITBUCKET_SERVER_USERNAME is not set")
	}

	// External service setup
	esID, err := setUpExternalService(t)
	if err != nil {
		t.Fatal(err)
	}

	gqltest.RemoveExternalServiceAfterTest(t, esID)

	// Triggering the sync job
	unrestricted := false
	err = gqltest.Client.SetRepositoryPermissionsForBitbucketProject(gqltestutil.BitbucketProjectPermsSyncArgs{
		ProjectKey: projectKey,
		CodeHost:   esID,
		UserPermissions: []types.UserPermission{
			{
				BindID:     "gqltest@sourcegraph.com", // existing user
				Permission: "READ",
			},
			{
				BindID:     "some-user-2@domain.com",
				Permission: "READ",
			},
			{
				BindID:     "some-user-3@domain.com",
				Permission: "READ",
			},
		},
		Unrestricted: &unrestricted,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Wait up to 30 seconds for worker to finish the permissions sync.
	err = waitForSyncJobToFinish()
	if err != nil {
		t.Fatal("Waiting for repository permissions to be synced:", err)
	}

	// Perform the checks
	// First we check pending permissions
	pendingPerms, err := gqltest.Client.UsersWithPendingPermissions()
	if err != nil {
		t.Fatal(err)
	}

	want := []string{
		"some-user-2@domain.com",
		"some-user-3@domain.com",
	}

	if diff := cmp.Diff(want, pendingPerms); diff != "" {
		t.Fatalf("Pending permissions mismatch (-want +got):\n%s", diff)
	}

	// Then we check existing user permissions
	permissionsInfo, err := gqltest.Client.UserPermissions(*gqltest.Username)
	if err != nil {
		t.Fatal(err)
	}

	if len(permissionsInfo) == 0 {
		t.Fatalf("User '%s' has no expected permissions", *gqltest.Username)
	}

	if permissionsInfo[0] != "READ" {
		t.Fatalf("READ permission hasn't been set for user '%s'", *gqltest.Username)
	}
}

func waitForSyncJobToFinish() error {
	return gqltestutil.Retry(30*time.Second, func() error {
		status, failureMessage, err := gqltest.Client.GetLastBitbucketProjectPermissionJob(projectKey)
		if err != nil || status == "" {
			return errors.New("Error during getting the status of a Bitbucket Permissions job")
		}

		if status == "errored" || status == "failed" {
			return errors.Newf("Bitbucket Permissions job failed with status: '%s' and failure message: '%s'", status, failureMessage)
		}

		if status == "completed" {
			return nil
		}
		return gqltestutil.ErrContinueRetry
	})
}

func setUpExternalService(t *testing.T) (esID string, err error) {
	t.Helper()
	// Set up external service.
	// It is configured to clone only "SOURCEGRAPH/jsonrpc2" repo, but this project
	// also has another repo "empty-repo-1"
	esID, err = gqltest.Client.AddExternalService(gqltestutil.AddExternalServiceInput{
		Kind:        extsvc.KindBitbucketServer,
		DisplayName: "gqltest-bitbucket-server",
		Config: gqltest.MustMarshalJSONString(struct {
			URL                   string   `json:"url"`
			Token                 string   `json:"token"`
			Username              string   `json:"username"`
			Repos                 []string `json:"repos"`
			RepositoryPathPattern string   `json:"repositoryPathPattern"`
		}{
			URL:                   *gqltest.BbsURL,
			Token:                 *gqltest.BbsToken,
			Username:              *gqltest.BbsUsername,
			Repos:                 []string{"SOURCEGRAPH/jsonrpc2", "SOURCEGRAPH/empty-repo-1"},
			RepositoryPathPattern: "bbs/{projectKey}/{repositorySlug}",
		}),
	})

	// The repo-updater might not be up yet, but it will eventually catch up for the
	// external service we just added, thus it is OK to ignore this transient error.
	if err != nil && !strings.Contains(err.Error(), "/sync-external-service") {
		return "", err
	}

	err = gqltest.Client.WaitForReposToBeCloned(clonedRepoName1)
	if err != nil {
		return "", err
	}
	return
}

func checkRepoPermissions(repoName string, wantUnrestricted bool) error {
	permissionsInfo, err := gqltest.Client.RepositoryPermissionsInfo(repoName)
	if err != nil {
		return err
	}

	if permissionsInfo.Permissions[0] != "READ" {
		return errors.New("READ permission hasn't been set")
	}

	if wantUnrestricted != permissionsInfo.Unrestricted {
		return errors.Newf("unrestricted permissions mismatch. Want: '%v', get: '%v'", wantUnrestricted, permissionsInfo.Unrestricted)
	}
	return nil
}
