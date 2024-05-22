package authtest

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestRepository(t *testing.T) {
	if len(*githubToken) == 0 {
		t.Skip("Environment variable GITHUB_TOKEN is not set")
	}

	// Set up external service
	esID, err := client.AddExternalService(
		gqltestutil.AddExternalServiceInput{
			Kind:        extsvc.KindGitHub,
			DisplayName: "authtest-github-repository",
			Config: mustMarshalJSONString(
				&schema.GitHubConnection{
					Authorization: &schema.GitHubAuthorization{},
					Repos: []string{
						"sgtest/go-diff",
						"sgtest/private", // Private
					},
					RepositoryPathPattern: "github.com/{nameWithOwner}",
					Token:                 *githubToken,
					Url:                   "https://ghe.sgdev.org/",
				},
			),
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := client.DeleteExternalService(esID, false)
		if err != nil {
			t.Fatal(err)
		}
	}()

	const privateRepo = "github.com/sgtest/private"
	err = client.WaitForReposToBeCloned(
		"github.com/sgtest/go-diff",
		privateRepo,
	)
	if err != nil {
		t.Fatal(err)
	}

	// Proactively schedule a permissions syncing.
	repo, err := client.Repository(privateRepo)
	if err != nil {
		t.Fatal(err)
	}
	err = client.ScheduleRepositoryPermissionsSync(repo.ID)
	if err != nil {
		t.Fatal(err)
	}

	// Wait up to 30 seconds for the private repository to have permissions synced
	// from the code host at least once.
	err = gqltestutil.Retry(30*time.Second, func() error {
		permsInfo, err := client.RepositoryPermissionsInfo(privateRepo)
		if err != nil {
			t.Fatal(err)
		}

		if permsInfo != nil && !permsInfo.SyncedAt.IsZero() {
			return nil
		}
		return gqltestutil.ErrContinueRetry
	})
	if err != nil {
		t.Fatal("Waiting for repository permissions to be synced:", err)
	}

	// Create a test user (authtest-user-repository) which is not a site admin, the
	// user should only have access to non-private repositories.
	const testUsername = "authtest-user-repository"
	userClient, err := gqltestutil.NewClient(*baseURL)
	require.NoError(t, err)
	require.NoError(t, userClient.SignUp(testUsername+"@sourcegraph.com", testUsername, "mysecurepassword"))
	defer func() {
		err := client.DeleteUser(userClient.AuthenticatedUserID(), true)
		if err != nil {
			t.Fatal(err)
		}
	}()

	t.Run("access repositories", func(t *testing.T) {
		tests := []struct {
			name    string
			repo    string
			wantNil bool
		}{
			{
				name:    "public repository",
				repo:    "github.com/sgtest/go-diff",
				wantNil: false,
			},
			// TODO: Flake: https://github.com/sourcegraph/sourcegraph/issues/28294
			// {
			// 	name:    "private repository",
			// 	repo:    privateRepo,
			// 	wantNil: true,
			// },
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				repo, err := userClient.Repository(test.repo)
				if err != nil {
					t.Fatal(err)
				}

				if diff := cmp.Diff(test.wantNil, repo == nil); diff != "" {
					t.Fatalf("Mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	// TODO: Flake: https://github.com/sourcegraph/sourcegraph/issues/28294
	// t.Run("search repositories", func(t *testing.T) {
	// 	results, err := userClient.SearchRepositories("type:repo sgtest")
	// 	if err != nil {
	// 		t.Fatal(err)
	// 	}
	// 	got := results.Exists(privateRepo)
	// 	want := []string{privateRepo}
	// 	if diff := cmp.Diff(want, got); diff != "" {
	// 		t.Fatalf("Missing mismatch (-want +got):\n%s", diff)
	// 	}
	// })
}
