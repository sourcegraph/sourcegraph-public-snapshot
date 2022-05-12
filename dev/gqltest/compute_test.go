package main

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
)

// computeClient is an interface so we can swap out a streaming vs grahql based
// compute API. It only supports the methods that streaming supports.
type computeClient interface {
	Compute(query string) ([]*gqltestutil.ComputeResult, error)
}

func testComputeClient(t *testing.T, client computeClient) {
	t.Run("errors", func(t *testing.T) {
		tests := []struct {
			query string
		}{
			{
				// Need an erroring query.
				query: "",
			},
		}
		for _, test := range tests {
			t.Run(test.query, func(t *testing.T) {
				// TODO: not actually sure how graphQL compute endpoint handles errors.
				results, err := client.Compute(test.query)
				if len(results) != 0 {
					t.Errorf("Expected err, got results: %v", results)
				}
				if err == nil {
					t.Error("Expected err, got nil")
				}
			})
		}
	})

}

func TestCompute(t *testing.T) {
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

	t.Run("graphql", func(t *testing.T) {
		testComputeClient(t, client)
	})

	streamClient := &gqltestutil.ComputeStreamClient{Client: client}
	t.Run("stream", func(t *testing.T) {
		testComputeClient(t, streamClient)
	})
}
