package main

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
)

type computeClient interface {
	Compute(query string) ([]gqltestutil.MatchContext, error)
}

func testComputeClient(t *testing.T, client computeClient) {
	t.Run("compute endpoint returns results", func(t *testing.T) {
		results, err := client.Compute(`repo:^github.com/sgtest/go-diff$ file:\.go func Parse(\w+)`)
		if err != nil {
			t.Errorf("unexpected error: %+v", err)
		}
		if len(results) == 0 {
			t.Error("Expected results, got none")
		}
	})
}

func TestCompute(t *testing.T) {
	t.Skip("for now")
	if len(*githubToken) == 0 {
		t.Skip("Environment variable GITHUB_TOKEN is not set")
	}

	// Set up external service
	_, err := client.AddExternalService(gqltestutil.AddExternalServiceInput{
		Kind:        extsvc.KindGitHub,
		DisplayName: "gqltest-github-search",
		Config: mustMarshalJSONString(struct {
			URL                   string   `json:"url"`
			Token                 string   `json:"token"`
			Repos                 []string `json:"repos"`
			RepositoryPathPattern string   `json:"repositoryPathPattern"`
		}{
			URL:   "https://ghe.sgdev.org/",
			Token: *githubToken,
			Repos: []string{
				"sgtest/java-langserver",
				"sgtest/jsonrpc2",
				"sgtest/go-diff",
				"sgtest/appdash",
				"sgtest/sourcegraph-typescript",
			},
			RepositoryPathPattern: "github.com/{nameWithOwner}",
		}),
	})
	if err != nil {
		t.Fatal(err)
	}

	err = client.WaitForReposToBeCloned(
		"github.com/sgtest/java-langserver",
		"github.com/sgtest/jsonrpc2",
		"github.com/sgtest/go-diff",
		"github.com/sgtest/appdash",
		"github.com/sgtest/sourcegraph-typescript",
	)
	if err != nil {
		t.Fatal(err)
	}

	err = client.WaitForReposToBeIndexed(
		"github.com/sgtest/java-langserver",
	)
	if err != nil {
		t.Fatal(err)
	}

	streamClient := &gqltestutil.ComputeStreamClient{Client: client}
	t.Run("stream", func(t *testing.T) {
		testComputeClient(t, streamClient)
	})
}
