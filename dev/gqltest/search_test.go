// +build gqltest

package main

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
)

func TestSearch(t *testing.T) {
	if len(*githubToken) == 0 {
		t.Skip("Environment variable GITHUB_TOKEN is not set")
	}

	// Set up external service
	esID, err := client.AddExternalService(gqltestutil.AddExternalServiceInput{
		Kind:        extsvc.KindGitHub,
		DisplayName: "e2e-test-github",
		Config: mustMarshalJSONString(struct {
			URL   string   `json:"url"`
			Token string   `json:"token"`
			Repos []string `json:"repos"`
		}{
			URL:   "http://github.com",
			Token: *githubToken,
			Repos: []string{
				"sourcegraph/java-langserver",
				"gorilla/mux",
				"gorilla/securecookie",
				"sourcegraph/jsonrpc2",
				"sourcegraph/go-diff",
				"sourcegraph/appdash",
				"sourcegraph/sourcegraph-typescript",
				"sourcegraph-testing/automation-e2e-test",
				"sourcegraph/e2e-test-private-repository",
			},
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := client.DeleteExternalService(esID)
		if err != nil {
			t.Fatal(err)
		}
	}()

	err = client.WaitForReposToBeCloned(
		"github.com/sourcegraph/java-langserver",
		"github.com/gorilla/mux",
		"github.com/gorilla/securecookie",
		"github.com/sourcegraph/jsonrpc2",
		"github.com/sourcegraph/go-diff",
		"github.com/sourcegraph/appdash",
		"github.com/sourcegraph/sourcegraph-typescript",
		"github.com/sourcegraph-testing/automation-e2e-test",
		"github.com/sourcegraph/e2e-test-private-repository",
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("visibility", func(t *testing.T) {
		tests := []struct {
			query       string
			wantMissing []string
		}{
			{
				query:       "type:repo visibility:private",
				wantMissing: []string{},
			},
			{
				query:       "type:repo visibility:public",
				wantMissing: []string{"github.com/sourcegraph/e2e-test-private-repository"},
			},
			{
				query:       "type:repo visibility:any",
				wantMissing: []string{},
			},
		}
		for _, test := range tests {
			t.Run(test.query, func(t *testing.T) {
				results, err := client.SearchRepositories(test.query)
				if err != nil {
					t.Fatal(err)
				}
				missing := results.Exists("github.com/sourcegraph/e2e-test-private-repository")
				if diff := cmp.Diff(test.wantMissing, missing); diff != "" {
					t.Fatalf("Missing mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("execute search with search parameters", func(t *testing.T) {
		results, err := client.SearchFiles("repo:^github.com/sourcegraph/go-diff$ type:file file:.go -file:.md")
		if err != nil {
			t.Fatal(err)
		}

		// Make sure only got .go files and no .md files
		for _, r := range results {
			if !strings.HasSuffix(r.Name, ".go") {
				t.Fatalf("Found file name does not end with .go: %s", r.Name)
			}
		}
	})
}
