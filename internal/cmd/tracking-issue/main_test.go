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

func TestIntegration(t *testing.T) {
	ti := &TrackingIssue{
		Issue: &Issue{
			Number:    7719,
			Milestone: "3.13",
			Labels:    []string{"tracking", "team/core-services"},
		},
	}

	loadTrackingIssueFixtures(t, "sourcegraph", ti)

	got := ti.Workloads().Markdown()
	path := filepath.Join("testdata", "issue.md")
	testutil.AssertGolden(t, path, *update, got)
}

func loadTrackingIssueFixtures(t testing.TB, org string, issue *TrackingIssue) {
	path := filepath.Join("testdata", "fixtures.json")

	if *updateFixture {
		ctx := context.Background()
		cli := graphql.NewClient(
			"https://api.github.com/graphql",
			graphql.WithHTTPClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
			))),
		)
		err := loadTrackingIssues(ctx, cli, org, []*TrackingIssue{issue})
		if err != nil {
			t.Fatal(err)
		}

		for _, issue := range issue.Issues {
			issue.Redact()
		}

		for _, pr := range issue.PRs {
			pr.Redact()
		}

		testutil.AssertGolden(t, path, true, issue)
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(issue); err != nil {
		t.Fatal(err)
	}
}
