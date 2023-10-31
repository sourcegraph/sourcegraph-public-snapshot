package main

import (
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
)

func TestExternalService(t *testing.T) {
	t.Skip("for now")
	if len(*githubToken) == 0 {
		t.Skip("Environment variable GITHUB_TOKEN is not set")
	}

	t.Run("repositoryPathPattern", func(t *testing.T) {
		const repo = "sgtest/go-diff" // Tiny repo, fast to clone
		const slug = "github.com/" + repo
		// Set up external service
		esID, err := client.AddExternalService(gqltestutil.AddExternalServiceInput{
			Kind:        extsvc.KindGitHub,
			DisplayName: "gqltest-github-repoPathPattern",
			Config: mustMarshalJSONString(struct {
				URL                   string   `json:"url"`
				Token                 string   `json:"token"`
				Repos                 []string `json:"repos"`
				RepositoryPathPattern string   `json:"repositoryPathPattern"`
			}{
				URL:                   "https://ghe.sgdev.org/",
				Token:                 *githubToken,
				Repos:                 []string{repo},
				RepositoryPathPattern: "github.com/{nameWithOwner}",
			}),
		})
		// The repo-updater might not be up yet, but it will eventually catch up for the external
		// service we just added, thus it is OK to ignore this transient error.
		if err != nil && !strings.Contains(err.Error(), "/sync-external-service") {
			t.Fatal(err)
		}
		removeExternalServiceAfterTest(t, esID)

		err = client.WaitForReposToBeCloned(slug)
		if err != nil {
			t.Fatal(err)
		}

		// The request URL should be redirected to the new path
		origURL := *baseURL + "/" + slug
		resp, err := client.Get(origURL)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = resp.Body.Close() }()

		wantURL := *baseURL + "/" + slug // <baseURL>/github.com/sgtest/go-diff
		if diff := cmp.Diff(wantURL, resp.Request.URL.String()); diff != "" {
			t.Fatalf("URL mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestExternalService_AWSCodeCommit(t *testing.T) {
	t.Skip("for now")
	if len(*awsAccessKeyID) == 0 || len(*awsSecretAccessKey) == 0 ||
		len(*awsCodeCommitUsername) == 0 || len(*awsCodeCommitPassword) == 0 {
		t.Skip("Environment variable AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_CODE_COMMIT_USERNAME or AWS_CODE_COMMIT_PASSWORD is not set")
	}

	// Set up external service
	esID, err := client.AddExternalService(gqltestutil.AddExternalServiceInput{
		Kind:        extsvc.KindAWSCodeCommit,
		DisplayName: "gqltest-aws-code-commit",
		Config: mustMarshalJSONString(struct {
			Region                string            `json:"region"`
			AccessKeyID           string            `json:"accessKeyID"`
			SecretAccessKey       string            `json:"secretAccessKey"`
			RepositoryPathPattern string            `json:"repositoryPathPattern"`
			GitCredentials        map[string]string `json:"gitCredentials"`
		}{
			Region:                "us-west-1",
			AccessKeyID:           *awsAccessKeyID,
			SecretAccessKey:       *awsSecretAccessKey,
			RepositoryPathPattern: "aws/{name}",
			GitCredentials: map[string]string{
				"username": *awsCodeCommitUsername,
				"password": *awsCodeCommitPassword,
			},
		}),
	})
	// The repo-updater might not be up yet, but it will eventually catch up for the external
	// service we just added, thus it is OK to ignore this transient error.
	if err != nil && !strings.Contains(err.Error(), "/sync-external-service") {
		t.Fatal(err)
	}
	removeExternalServiceAfterTest(t, esID)

	const repoName = "aws/test"
	err = client.WaitForReposToBeCloned(repoName)
	if err != nil {
		t.Fatal(err)
	}

	blob, err := client.GitBlob(repoName, "master", "README")
	if err != nil {
		t.Fatal(err)
	}

	wantBlob := "README\n\nchange"
	if diff := cmp.Diff(wantBlob, blob); diff != "" {
		t.Fatalf("Blob mismatch (-want +got):\n%s", diff)
	}
}

func TestExternalService_BitbucketServer(t *testing.T) {
	t.Skip("for now")
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
	// The repo-updater might not be up yet, but it will eventually catch up for the external
	// service we just added, thus it is OK to ignore this transient error.
	if err != nil && !strings.Contains(err.Error(), "/sync-external-service") {
		t.Fatal(err)
	}
	removeExternalServiceAfterTest(t, esID)

	const repoName = "bbs/SOURCEGRAPH/jsonrpc2"
	err = client.WaitForReposToBeCloned(repoName)
	if err != nil {
		t.Fatal(err)
	}

	blob, err := client.GitBlob(repoName, "master", ".travis.yml")
	if err != nil {
		t.Fatal(err)
	}

	wantBlob := "language: go\ngo: \n - 1.x\n\nscript:\n - go test -race -v ./...\n"
	if diff := cmp.Diff(wantBlob, blob); diff != "" {
		t.Fatalf("Blob mismatch (-want +got):\n%s", diff)
	}
}

func TestExternalService_Perforce(t *testing.T) {
	t.Skip("for now")
	for _, tc := range []struct {
		name       string
		depot      string
		useFusion  bool
		headBranch string
		blobPath   string
		wantBlob   string
	}{
		{
			name:       "git p4",
			depot:      "test-perms",
			useFusion:  false,
			blobPath:   "README.md",
			headBranch: "master",
			wantBlob: `This depot is used to test user and group permissions.
`,
		},
		{
			name:       "p4 fusion",
			depot:      "integration-test-depot",
			useFusion:  true,
			blobPath:   "path.txt",
			headBranch: "main",
			wantBlob: `./
`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repoName := "perforce/" + tc.depot
			checkPerforceEnvironment(t)
			cleanup := createPerforceExternalService(t, tc.depot, tc.useFusion)
			t.Cleanup(cleanup)

			err := client.WaitForReposToBeCloned(repoName)
			if err != nil {
				t.Fatal(err)
			}

			blob, err := client.GitBlob(repoName, tc.headBranch, tc.blobPath)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, tc.wantBlob, blob)
		})
	}
}

func checkPerforceEnvironment(t *testing.T) {
	if len(*perforcePort) == 0 || len(*perforceUser) == 0 || len(*perforcePassword) == 0 {
		t.Skip("Environment variables PERFORCE_PORT, PERFORCE_USER or PERFORCE_PASSWORD are not set")
	}
}

// createPerforceExternalService creates an Perforce external service that
// includes the supplied depot. It returns a function to cleanup after the test
// which will delete the depot from disk and remove the external service.
func createPerforceExternalService(t *testing.T, depot string, useP4Fusion bool) func() {
	t.Helper()

	type Authorization = struct {
		SubRepoPermissions bool `json:"subRepoPermissions"`
	}
	type FusionClient = struct {
		Enabled   bool `json:"enabled"`
		LookAhead int  `json:"lookAhead,omitempty"`
	}

	// Set up external service
	esID, err := client.AddExternalService(gqltestutil.AddExternalServiceInput{
		Kind:        extsvc.KindPerforce,
		DisplayName: "gqltest-perforce-server",
		Config: mustMarshalJSONString(struct {
			P4Port                string        `json:"p4.port"`
			P4User                string        `json:"p4.user"`
			P4Password            string        `json:"p4.passwd"`
			Depots                []string      `json:"depots"`
			RepositoryPathPattern string        `json:"repositoryPathPattern"`
			FusionClient          FusionClient  `json:"fusionClient"`
			Authorization         Authorization `json:"authorization"`
		}{
			P4Port:                *perforcePort,
			P4User:                *perforceUser,
			P4Password:            *perforcePassword,
			Depots:                []string{"//" + depot + "/"},
			RepositoryPathPattern: "perforce/{depot}",
			FusionClient: FusionClient{
				Enabled:   useP4Fusion,
				LookAhead: 2000,
			},
			Authorization: Authorization{
				SubRepoPermissions: true,
			},
		}),
	})

	// The repo-updater might not be up yet but it will eventually catch up for the
	// external service we just added, thus it is OK to ignore this transient error.
	if err != nil && !strings.Contains(err.Error(), "/sync-external-service") {
		t.Fatal(err)
	}

	return func() {
		if err := client.DeleteRepoFromDiskByName("perforce/" + depot); err != nil {
			t.Fatalf("removing depot from disk: %v", err)
		}

		if err := client.DeleteExternalService(esID, false); err != nil {
			t.Fatalf("removing external service: %v", err)
		}
	}
}

func TestExternalService_AsyncDeletion(t *testing.T) {
	t.Skip("for now")
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
	// The repo-updater might not be up yet, but it will eventually catch up for the external
	// service we just added, thus it is OK to ignore this transient error.
	if err != nil && !strings.Contains(err.Error(), "/sync-external-service") {
		t.Fatal(err)
	}

	err = client.DeleteExternalService(esID, true)
	if err != nil {
		t.Fatal(err)
	}

	// This call should return not found error.
	// Retrying to wait for async deletion to finish. Deletion
	// might be blocked on the first sync of the external service still
	// running. It cancels that syncs and waits for it to finish.
	err = gqltestutil.Retry(30*time.Second, func() error {
		err = client.CheckExternalService(esID)
		if err == nil {
			return gqltestutil.ErrContinueRetry
		}
		return err
	})
	if err == nil || err == gqltestutil.ErrContinueRetry {
		t.Fatal("Deleted service should not be found")
	}
	if !strings.Contains(err.Error(), "external service not found") {
		t.Fatalf("Not found error should be returned, got: %s", err.Error())
	}
}

func removeExternalServiceAfterTest(t *testing.T, esID string) {
	t.Helper()
	t.Cleanup(func() {
		err := client.DeleteExternalService(esID, false)
		if err != nil {
			t.Fatal(err)
		}
	})
}
