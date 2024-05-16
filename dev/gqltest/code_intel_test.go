package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
	"github.com/sourcegraph/sourcegraph/schema"
)

func unwrap[T any](v T, err error) func(*testing.T) T {
	return func(t *testing.T) T {
		require.NoError(t, err)
		return v
	}
}

func TestCodeGraphAPIs(t *testing.T) {
	if len(*githubToken) == 0 {
		t.Skip("Environment variable GITHUB_TOKEN is not set")
	}

	reset, err := client.ModifySiteConfiguration(func(siteConfig *schema.SiteConfiguration) {
		siteConfig.CodeIntelAutoIndexingEnabled = pointers.Ptr(true)
	})
	require.NoError(t, err)
	if reset != nil {
		t.Cleanup(func() {
			require.NoError(t, reset())
		})
	}

	extSvcID := unwrap(client.AddExternalService(gqltestutil.AddExternalServiceInput{
		Kind:        extsvc.KindGitHub,
		DisplayName: "gqltest-code-graph-apis",
		Config: mustMarshalJSONString(&schema.GitHubConnection{
			Authorization: &schema.GitHubAuthorization{},
			Url:           "https://ghe.sgdev.org/",
			Token:         *githubToken,
			Repos: []string{
				"sgtest/go-diff",
			},
			RepositoryPathPattern: "github.com/{nameWithOwner}",
		}),
	}))(t)

	removeExternalServiceAfterTest(t, extSvcID)
	start := time.Now()
	err = client.WaitForReposToBeCloned("github.com/sgtest/go-diff")
	cloneTime := time.Since(start)
	require.NoError(t, err)

	jobs := unwrap(client.TriggerAutoIndexing("github.com/sgtest/go-diff"))(t)

	timeout := 2 * time.Minute
	start = time.Now()
	jobStateMap, err := client.WaitForAutoIndexingJobsToComplete(jobs, timeout)
	require.NoErrorf(t, err, "jobStateMap: %v", jobStateMap)
	autoIndexingTime := time.Since(start)
	panic(fmt.Sprintf("Auto-indexing timing: indexing time - %s, clone time - %s", autoIndexingTime, cloneTime))
	panic("Auto-indexing job completed!")
}
