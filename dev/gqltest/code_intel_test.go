package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
)

func unwrap[T any](v T, err error) func(*testing.T) T {
	return func(t *testing.T) T {
		require.NoError(t, err)
		return v
	}
}

func TestCodeGraphAPIs(t *testing.T) {
	if len(*githubToken) == 0 {
		t.Fatal("Testing code graph APIs missing token")
		t.Skip("Environment variable GITHUB_TOKEN is not set")
	}

	extSvcID := unwrap(client.AddExternalService(gqltestutil.AddExternalServiceInput{
		Kind:        extsvc.KindGitHub,
		DisplayName: "gqltest-code-graph-apis",
		Config: mustMarshalJSONString(struct {
			URL                   string   `json:"url"`
			Token                 string   `json:"token"`
			Repos                 []string `json:"repos"`
			RepositoryPathPattern string   `json:"repositoryPathPattern"`
		}{
			URL:   "https://ghe.sgdev.org/",
			Token: *githubToken,
			Repos: []string{
				"sourcegraph/conc",
			},
			RepositoryPathPattern: "github.com/{nameWithOwner}",
		}),
	}))(t)

	removeExternalServiceAfterTest(t, extSvcID)
	err := client.WaitForReposToBeCloned("github.com/sourcegraph/conc")
	require.NoError(t, err)

	jobs := unwrap(client.TriggerAutoIndexing("github.com/sourcegraph/conc"))(t)

	timeout := 5 * time.Minute
	jobStates := unwrap(client.WaitForAutoIndexingJobsToComplete(jobs, timeout))(t)

	fmt.Printf("%v\n", jobStates)
	t.Fatal("Testing if this test is actually running")
	// Now run precise code nav queries against this code! :)
}
