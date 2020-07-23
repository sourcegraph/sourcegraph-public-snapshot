// +build gqltest

package main

import (
	"fmt"
	"strings"
	"testing"
	"time"

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
		DisplayName: "gqltest-github-search",
		Config: mustMarshalJSONString(struct {
			URL   string   `json:"url"`
			Token string   `json:"token"`
			Repos []string `json:"repos"`
		}{
			URL:   "http://github.com",
			Token: *githubToken,
			Repos: []string{
				"sgtest/java-langserver",
				"sgtest/jsonrpc2",
				"sgtest/go-diff",
				"sgtest/appdash",
				"sgtest/sourcegraph-typescript",
				"sgtest/private",
				"sgtest/mux", // Fork
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
		"github.com/sgtest/java-langserver",
		"github.com/sgtest/jsonrpc2",
		"github.com/sgtest/go-diff",
		"github.com/sgtest/appdash",
		"github.com/sgtest/sourcegraph-typescript",
		"github.com/sgtest/private",
		"github.com/sgtest/mux", // Fork
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
				wantMissing: []string{"github.com/sgtest/private"},
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
				missing := results.Exists("github.com/sgtest/private")
				if diff := cmp.Diff(test.wantMissing, missing); diff != "" {
					t.Fatalf("Missing mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("execute search with search parameters", func(t *testing.T) {
		results, err := client.SearchFiles("repo:^github.com/sgtest/go-diff$ type:file file:.go -file:.md")
		if err != nil {
			t.Fatal(err)
		}

		// Make sure only got .go files and no .md files
		for _, r := range results.Results {
			if !strings.HasSuffix(r.File.Name, ".go") {
				t.Fatalf("Found file name does not end with .go: %s", r.File.Name)
			}
		}
	})

	t.Run("multiple revisions per repository", func(t *testing.T) {
		results, err := client.SearchFiles("repo:sgtest/go-diff$@master:print-options:*refs/heads/ func NewHunksReader")
		if err != nil {
			t.Fatal(err)
		}

		wantExprs := map[string]struct{}{
			"master":        {},
			"print-options": {},

			// These next 2 branches are included because of the *refs/heads/ in the query.
			"test-already-exist-pr": {},
			"bug-fix-wip":           {},
		}

		for _, r := range results.Results {
			delete(wantExprs, r.RevSpec.Expr)
		}

		if len(wantExprs) > 0 {
			missing := make([]string, 0, len(wantExprs))
			for expr := range wantExprs {
				missing = append(missing, expr)
			}
			t.Fatalf("Missing exprs: %v", missing)
		}
	})

	t.Run("repository groups", func(t *testing.T) {
		const repoName = "github.com/sgtest/go-diff"
		err := client.OverwriteSettings(client.AuthenticatedUserID(), fmt.Sprintf(`{"search.repositoryGroups":{"gql_test_group": ["%s"]}}`, repoName))
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			err := client.OverwriteSettings(client.AuthenticatedUserID(), `{}`)
			if err != nil {
				t.Fatal(err)
			}
		}()

		results, err := client.SearchFiles("repogroup:gql_test_group diff.")
		if err != nil {
			t.Fatal(err)
		}

		// Make sure there are results and all results are from the same repository
		if len(results.Results) == 0 {
			t.Fatal("Unexpected zero result")
		}
		for _, r := range results.Results {
			if r.Repository.Name != repoName {
				t.Fatalf("Repository: want %q but got %q", repoName, r.Repository.Name)
			}
		}
	})

	t.Run("search statistics", func(t *testing.T) {
		err := client.OverwriteSettings(client.AuthenticatedUserID(), `{"experimentalFeatures":{"searchStats": true}}`)
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			err := client.OverwriteSettings(client.AuthenticatedUserID(), `{}`)
			if err != nil {
				t.Fatal(err)
			}
		}()

		var lastResult *gqltestutil.SearchStatsResult
		// Retry because the configuration update endpoint is eventually consistent
		err = gqltestutil.Retry(5*time.Second, func() error {
			// This is a substring that appears in the sgtest/go-diff repository.
			// It is OK if it starts to appear in other repositories, the test just
			// checks that it is found in at least 1 Go file.
			result, err := client.SearchStats("Incomplete-Lines")
			if err != nil {
				t.Fatal(err)
			}
			lastResult = result

			for _, lang := range result.Languages {
				if strings.EqualFold(lang.Name, "Go") {
					return nil
				}
			}

			return gqltestutil.ErrContinueRetry
		})
		if err != nil {
			t.Fatal(err, "lastResult:", lastResult)
		}
	})

	t.Run("global search", func(t *testing.T) {
		tests := []struct {
			name          string
			query         string
			zeroResult    bool
			minMatchCount int64
		}{
			// Global search
			{
				name:  "error",
				query: "error",
			},
			{
				name:  "error count:1000",
				query: "error count:1000",
			},
			{
				name:          "something with more than 1000 results and use count:1000",
				query:         ". count:1000",
				minMatchCount: 1001,
			},
			{
				name:  "regular expression without indexed search",
				query: "index:no patterntype:regexp ^func.*$",
			},
			{
				name:  "fork:only",
				query: "fork:only router",
			},
			{
				name:  "double-quoted pattern, nonzero result",
				query: `"func main() {\n" patterntype:regexp count:1 stable:yes`,
			},
			{
				name:  "exclude repo, nonzero result",
				query: `"func main() {\n" -repo:go-diff patterntype:regexp count:1 stable:yes`,
			},
			{
				name:       "fork:no",
				query:      "fork:no FORK_SENTINEL",
				zeroResult: true,
			},
			{
				name:       "random characters, zero results",
				query:      "asdfalksd+jflaksjdfklas patterntype:literal -repo:sourcegraph",
				zeroResult: true,
			},
			// Repo search
			{
				name:  "repo search by name, nonzero result",
				query: "repo:go-diff$",
			},
			{
				name:  "repo search by name, case yes, nonzero result",
				query: `repo:^github\.com/sgtest/go-diff$ String case:yes count:1 stable:yes`,
			},
			{
				name:  "true is an alias for yes when fork is set",
				query: `repo:github\.com/sgtest/mux fork:true`,
			},
			{
				name:  "non-master branch, nonzero result",
				query: `repo:^github\.com/sgtest/java-langserver$@v1 void sendPartialResult(Object requestId, JsonPatch jsonPatch); patterntype:literal count:1 stable:yes`,
			},
			{
				name:  "indexed multiline search, nonzero result",
				query: `repo:^github\.com/sgtest/java-langserver$ \nimport index:only patterntype:regexp count:1 stable:yes`,
			},
			{
				name:  "unindexed multiline search, nonzero result",
				query: `repo:^github\.com/sgtest/java-langserver$ \nimport index:no patterntype:regexp count:1 stable:yes`,
			},
			{
				name:       "random characters, zero result",
				query:      `repo:^github\.com/sgtest/java-langserver$ doesnot734734743734743exist`,
				zeroResult: true,
			},
			// Filename search
			{
				name:  "search for a known file",
				query: "file:doc.go",
			},
			{
				name:       "search for a non-existent file",
				query:      "file:asdfasdf.go",
				zeroResult: true,
			},
			// Symbol search
			{
				name:  "search for a known symbol",
				query: "type:symbol count:100 patterntype:regexp ^newroute",
			},
			{
				name:       "search for a non-existent symbol",
				query:      "type:symbol asdfasdf",
				zeroResult: true,
			},
			// Commit search
			{
				name:  "commit search, nonzero result",
				query: `repo:^github\.com/sgtest/go-diff$ type:commit count:1`,
			},
			// Diff search
			{
				name:  "diff search, nonzero result",
				query: `repo:^github\.com/sgtest/go-diff$ type:diff main count:1`,
			},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				results, err := client.SearchFiles(test.query)
				if err != nil {
					t.Fatal(err)
				}

				if test.zeroResult {
					if len(results.Results) > 0 {
						t.Fatalf("Want zero result but got %d", len(results.Results))
					}
				} else {
					if len(results.Results) == 0 {
						t.Fatal("Want non-zero results but got 0")
					}
				}

				if results.MatchCount < test.minMatchCount {
					t.Fatalf("Want at least %d match count but got %d", test.minMatchCount, results.MatchCount)
				}
			})
		}
	})

	t.Run("timeout search options", func(t *testing.T) {
		alert, err := client.SearchAlert(`router index:no timeout:1ns`)
		if err != nil {
			t.Fatal(err)
		}

		if alert == nil {
			t.Fatal("Want search alert but got nil")
		}
	})
}
