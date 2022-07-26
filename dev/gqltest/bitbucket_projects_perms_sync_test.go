package main

import (
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestBitbucketProjectsPermsSync(t *testing.T) {
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

	const projectKey = "SOURCEGRAPH"

	tests := []struct {
		name         string
		projectKey   string
		codeHost     string
		userPerms    []types.UserPermission
		unrestricted bool
		checkFunc    func() error
	}{
		{
			name:         "Setting unrestricted permissions",
			projectKey:   projectKey,
			codeHost:     esID,
			userPerms:    make([]types.UserPermission, 0),
			unrestricted: true,
			checkFunc: func() error {
				permissionsInfo, err := client.RepositoryPermissionsInfo(repoName)
				if err != nil {
					return err
				}

				if permissionsInfo.Permissions[0] != "READ" {
					return errors.Wrap(err, "READ permission hasn't been set")
				}

				if !permissionsInfo.Unrestricted {
					return errors.Wrap(err, "unrestricted permission hasn't been set:")
				}
				return nil
			},
		},
		{
			name:       "Setting pending permissions for non-existing users",
			projectKey: projectKey,
			codeHost:   esID,
			userPerms: []types.UserPermission{
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
			unrestricted: false,
			checkFunc: func() error {
				pendingPerms, err := client.UsersWithPendingPermissions()
				if err != nil {
					return err
				}

				if len(pendingPerms) != 3 {
					return errors.Newf("Expected 3 pending permissions entries, got: %d", len(pendingPerms))
				}

				want := []string{
					"some-user-1@domain.com",
					"some-user-2@domain.com",
					"some-user-3@domain.com",
				}

				if diff := cmp.Diff(want, pendingPerms); diff != "" {
					return errors.Newf("Pending permissions mismatch (-want +got):\n%s", diff)
				}

				return nil
			},
		},
		{
			name:       "Setting permissions for both existing and non-existing users",
			projectKey: projectKey,
			codeHost:   esID,
			userPerms: []types.UserPermission{
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
			unrestricted: false,
			checkFunc: func() error {
				// First we check pending permissions
				pendingPerms, err := client.UsersWithPendingPermissions()
				if err != nil {
					return err
				}

				want := []string{
					"some-user-2@domain.com",
					"some-user-3@domain.com",
				}

				if diff := cmp.Diff(want, pendingPerms); diff != "" {
					return errors.Newf("Pending permissions mismatch (-want +got):\n%s", diff)
				}

				// Then we check existing user permissions
				permissionsInfo, err := client.UserPermissions(*username)
				if err != nil {
					return err
				}

				if len(permissionsInfo) == 0 {
					return errors.Newf("User '%s' has no expected permissions", *username)
				}

				if permissionsInfo[0] != "READ" {
					return errors.Wrapf(err, "READ permission hasn't been set for user '%s'", *username)
				}

				return nil
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err = client.SetRepositoryPermissionsForBitbucketProject(gqltestutil.BitbucketProjectPermsSyncArgs{
				ProjectKey:      test.projectKey,
				CodeHost:        test.codeHost,
				UserPermissions: test.userPerms,
				Unrestricted:    &test.unrestricted,
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

				if status == "errored" || status == "failed" {
					t.Fatalf("Bitbucket Permissions job failed with status: '%s'", status)
				}

				if status == "completed" {
					return nil
				}
				return gqltestutil.ErrContinueRetry
			})
			if err != nil {
				t.Fatal("Waiting for repository permissions to be synced:", err)
			}

			// Perform checks described in the test case
			if err = test.checkFunc(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
