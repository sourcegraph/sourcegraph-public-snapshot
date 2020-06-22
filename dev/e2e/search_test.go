// +build e2e

package main

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/e2eutil"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
)

func TestSearch(t *testing.T) {
	if len(*githubToken) == 0 {
		t.Fatal("Environment variable GITHUB_TOKEN is not set")
	}

	// Set up external service
	esID, err := client.AddExternalService(e2eutil.AddExternalServiceInput{
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

	t.Run("type:repo visibility:private", func(t *testing.T) {
		results, err := client.SearchRepositories("type:repo visibility:private")
		if err != nil {
			t.Fatal(err)
		}
		missing := results.Exists("github.com/sourcegraph/e2e-test-private-repository")
		if len(missing) > 0 {
			t.Fatalf("private repository not found: %v", missing)
		}
	})
}
