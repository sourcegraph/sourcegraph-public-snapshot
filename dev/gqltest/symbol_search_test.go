package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestSymbolSearch(t *testing.T) {
	if len(*githubToken) == 0 {
		t.Skip("Environment variable GITHUB_TOKEN is not set")
	}

	// Set up external service
	esID, err := client.AddExternalService(gqltestutil.AddExternalServiceInput{
		Kind:        extsvc.KindGitHub,
		DisplayName: "gqltest-github-search",
		Config: mustMarshalJSONString(&schema.GitHubConnection{
			Url:   "https://ghe.sgdev.org/",
			Token: *githubToken,
			Repos: []string{
				"sgtest/java-langserver",
				"sgtest/go-diff",
			},
			RepositoryPathPattern: "github.com/{nameWithOwner}",
		}),
	})
	require.NoError(t, err)

	removeExternalServiceAfterTest(t, esID)

	err = client.WaitForReposToBeIndexed(
		"github.com/sgtest/java-langserver",
		"github.com/sgtest/go-diff",
	)
	require.NoError(t, err)

	cases := []struct {
		name          string
		query         string
		zeroResult    bool
		wantExtension string
	}{
		{
			name:  "search for a known symbol",
			query: "type:symbol count:100 patterntype:regexp ^newroute",
		},
		{
			name:       "search for a non-existent symbol",
			query:      "type:symbol asdfasdf",
			zeroResult: true,
		},
		{
			name:  "search using boolean expression",
			query: "type:symbol count:100 newroute OR nonexistentsymbol",
		},
		{
			// Mimic the web client's search for 'go to definition'
			name:          "code nav: go to definition",
			query:         "type:symbol ^readLine$ lang:go case:yes patterntype:regexp",
			wantExtension: ".go",
		},
		{
			// Mimic the web client's search for 'find references'
			name:          "code nav: find references",
			query:         "type:file \\bBYTES_TO_GIGABYTES\\b lang:java case:yes patterntype:regexp",
			wantExtension: ".java",
		},
		{
			// Mimic the web client's search for 'find references' with unrecognized language
			name:          "code nav: find references with unrecognized language",
			query:         "type:file \\bLIGHTSTEP_INCLUDE_SENSITIVE\\b file:\\.yml case:yes patterntype:regexp",
			wantExtension: ".yml",
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			results, err := client.SearchFiles(test.query)
			require.NoError(t, err)
			require.Nil(t, results.Alert)

			if test.zeroResult {
				if len(results.Results) > 0 {
					t.Fatalf("Want zero result but got %d", len(results.Results))
				}
			} else {
				if len(results.Results) == 0 {
					t.Fatal("Want non-zero results but got 0")
				}
			}

			for _, r := range results.Results {
				if !strings.HasSuffix(strings.ToLower(r.File.Name), test.wantExtension) {
					t.Fatalf("Found file name that does not end with %s: %s", test.wantExtension, r.File.Name)
				}
			}
		})
	}
}
