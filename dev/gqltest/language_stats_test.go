package main

import (
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
	"github.com/sourcegraph/sourcegraph/schema"
	"testing"
)

func TestRepositoryLanguageStats(t *testing.T) {
	if len(*githubToken) == 0 {
		t.Skip("Environment variable GITHUB_TOKEN is not set")
	}

	// Set up external service
	esID, err := client.AddExternalService(gqltestutil.AddExternalServiceInput{
		Kind:        extsvc.KindGitHub,
		DisplayName: "gqltest-github-repository-lang-stats",
		Config: mustMarshalJSONString(&schema.GitHubConnection{
			Url:   "https://ghe.sgdev.org/",
			Token: *githubToken,
			Repos: []string{
				"sgtest/go-diff",
			},
			RepositoryPathPattern: "github.com/{nameWithOwner}",
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	removeExternalServiceAfterTest(t, esID)

	err = client.WaitForReposToBeCloned(
		"github.com/sgtest/go-diff",
	)
	if err != nil {
		t.Fatal(err)
	}

	var res struct {
		Repository struct {
			Name          string
			LanguageStats []struct {
				Name       string
				TotalBytes float64
				TotalLines int
			}
		} `graphql:"repository(name: $name)"`
	}
	variables := map[string]any{
		"name": "github.com/sgtest/go-diff",
	}
	q := `
		query LangStatsInsightContent($query: String!) {
			search(query: $query) {
				results {
					limitHit
				}
				stats {
					languages {
						name
						totalLines
					}
				}
			}
		}`
	err = client.GraphQL("", q, variables, &res)
	if err != nil {
		t.Fatal(err)
	}

	// Check the repository name
	if res.Repository.Name != "github.com/sgtest/go-diff" {
		t.Fatalf("unexpected repository name. want=%q, have=%q", "github.com/sgtest/go-diff", res.Repository.Name)
	}

	// Check language stats
	expectedStats := []struct {
		Name string
	}{
		{Name: "Go"},
	}
	if len(res.Repository.LanguageStats) != len(expectedStats) {
		t.Fatalf("unexpected number of language stats. want=%d, have=%d", len(expectedStats), len(res.Repository.LanguageStats))
	}
	for i, expectedStat := range expectedStats {
		if res.Repository.LanguageStats[i].Name != expectedStat.Name {
			t.Errorf("unexpected language name at position %d. want=%q, have=%q", i, expectedStat.Name, res.Repository.LanguageStats[i].Name)
		}
		if res.Repository.LanguageStats[i].TotalBytes <= 0 {
			t.Errorf("unexpected total bytes at position %d. want > 0, have=%f", i, res.Repository.LanguageStats[i].TotalBytes)
		}
		if res.Repository.LanguageStats[i].TotalLines <= 0 {
			t.Errorf("unexpected total lines at position %d. want > 0, have=%d", i, res.Repository.LanguageStats[i].TotalLines)
		}
	}

	// make it fail
	t.Error("make it fail")
}
