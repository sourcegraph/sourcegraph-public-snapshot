package main

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/machinebox/graphql"
	"golang.org/x/oauth2"

	"github.com/sourcegraph/sourcegraph/internal/testutil"
)

var (
	updateFixture = flag.Bool("update.fixture", false, "update testdata input")
	update        = flag.Bool("update", false, "update testdata golden")
)

func TestGenerate(t *testing.T) {
	milestone := "3.13"
	issues, prs := getIssuesAndPullRequestsFixtures(t, "sourcegraph", milestone, []string{"team/core-services"})
	got := generate(workloads(issues, prs, milestone))
	path := filepath.Join("testdata", "issue.md")
	testutil.AssertGolden(t, path, *update, got)
}

func getIssuesAndPullRequestsFixtures(t testing.TB, org, milestone string, labels []string) ([]*Issue, []*PullRequest) {
	type fixtures struct {
		Issues       []*Issue
		PullRequests []*PullRequest
	}

	path := filepath.Join("testdata", "fixtures.json")

	if *updateFixture {
		ctx := context.Background()
		cli := graphql.NewClient(
			"https://api.github.com/graphql",
			graphql.WithHTTPClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
			))),
		)
		issues, prs, err := listIssuesAndPullRequests(ctx, cli, org, milestone, labels)
		if err != nil {
			t.Fatal(err)
		}

		for _, issue := range issues {
			issue.Redact()
		}

		for _, pr := range prs {
			pr.Redact()
		}

		testutil.AssertGolden(t, path, true, fixtures{issues, prs})
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	var v fixtures
	if err := json.NewDecoder(f).Decode(&v); err != nil {
		t.Fatal(err)
	}

	return v.Issues, v.PullRequests
}
