package main

import (
	"strings"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestBitbucketProjectsPermsSync_SetPermissionsUnrestricted(t *testing.T) {
	if len(*bbsURL) == 0 || len(*bbsToken) == 0 || len(*bbsUsername) == 0 {
		t.Skip("Environment variable BITBUCKET_SERVER_URL, BITBUCKET_SERVER_TOKEN, or BITBUCKET_SERVER_USERNAME is not set")
	}

	// Set up external service
	esID, err := client.AddExternalService(gqltestutil.AddExternalServiceInput{
		Kind:        extsvc.KindBitbucketServer,
		DisplayName: "gqltest-bitbucket-server",
		Config: mustMarshalJSONString(struct {
			URL                   string   `json:"url"`
			Token                 string   `json:"token"`
			Username              string   `json:"username"`
			Repos                 []string `json:"repos"`
			RepositoryPathPattern string   `json:"repositoryPathPattern"`
		}{
			URL:                   *bbsURL,
			Token:                 *bbsToken,
			Username:              *bbsUsername,
			Repos:                 []string{"SOURCEGRAPH/jsonrpc2"},
			RepositoryPathPattern: "bbs/{projectKey}/{repositorySlug}",
		}),
	})

	// The repo-updater might not be up yet, but it will eventually catch up for the
	// external service we just added, thus it is OK to ignore this transient error.
	if err != nil && !strings.Contains(err.Error(), "/sync-external-service") {
		t.Fatal(err)
	}
	defer func() {
		err := client.DeleteExternalService(esID, false)
		if err != nil {
			t.Fatal(err)
		}
	}()

	const repoName = "bbs/SOURCEGRAPH/jsonrpc2"
	err = client.WaitForReposToBeCloned(repoName)
	if err != nil {
		t.Fatal(err)
	}

	// Setting unrestricted permissions to all repos of SOURCEGRAPH Bitbucket
	// project SG has only bbs/SOURCEGRAPH/jsonrpc2 repo cloned -- this repo should
	// be unrestricted
	unrestricted := true
	const projectKey = "SOURCEGRAPH"
	err = client.SetRepositoryPermissionsForBitbucketProject(gqltestutil.BitbucketProjectPermsSyncArgs{
		ProjectKey:      projectKey,
		CodeHost:        esID,
		UserPermissions: make([]types.UserPermission, 0),
		Unrestricted:    &unrestricted,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Wait up to 30 seconds for worker to finish the permissions sync.
	err = gqltestutil.Retry(30*time.Second, func() error {
		status, err := client.GetLastBitbucketProjectPermissionJob(projectKey)
		if err != nil || status == "" {
			t.Fatal("Error during getting the status of a Bitbucket Permissions job")
		}

		if status == "completed" {
			return nil
		}
		return gqltestutil.ErrContinueRetry
	})
	if err != nil {
		t.Fatal("Waiting for repository permissions to be synced:", err)
	}

	// Checking that unrestricted permission is set to bbs/SOURCEGRAPH/jsonrpc2
	permissionsInfo, err := client.RepositoryPermissionsInfo(repoName)
	if err != nil {
		t.Fatal(err)
	}

	if permissionsInfo.Permissions[0] != "READ" {
		t.Fatal("READ permission hasn't been set:", err)
	}

	if !permissionsInfo.Unrestricted {
		t.Fatal("unrestricted permission hasn't been set:", err)
	}
}
