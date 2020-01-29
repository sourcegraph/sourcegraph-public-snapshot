package main

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shurcooL/githubv4"
	"github.com/sourcegraph/sourcegraph/internal/testutil"
	"golang.org/x/oauth2"
)

var (
	updateFixture = flag.Bool("update.fixture", false, "update testdata input")
	update        = flag.Bool("update", false, "update testdata golden")
)

func TestGenerate(t *testing.T) {
	milestone := "3.13"
	issues := getIssuesFixture(t, "sourcegraph", milestone, []string{"team/core-services"})
	got := generate(issues, milestone)
	path := filepath.Join("testdata", "issue.md")
	testutil.AssertGolden(t, path, *update, got)
}

func getIssuesFixture(t testing.TB, org, milestone string, labels []string) []*Issue {
	path := filepath.Join("testdata", "issues.json")
	if *updateFixture {
		ctx := context.Background()
		cli := githubv4.NewClient(
			oauth2.NewClient(ctx, oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
			)),
		)
		issues, err := listIssues(ctx, cli, org, milestone, labels)
		if err != nil {
			t.Fatal(err)
		}
		for _, issue := range issues {
			if issue.Private {
				// Whitelist of fields to prevent leaking data in fixture.
				labels := issue.Labels[:0]
				for _, label := range issue.Labels {
					if strings.HasPrefix(label, "estimate/") || strings.HasPrefix(label, "planned/") {
						labels = append(labels, label)
					}
				}
				*issue = Issue{
					Title:      "REDACTED",
					Private:    true,
					Labels:     labels,
					Number:     issue.Number,
					URL:        issue.URL,
					State:      issue.State,
					Repository: issue.Repository,
					Assignees:  issue.Assignees,
					Milestone:  issue.Milestone,
				}
			}
		}
		testutil.AssertGolden(t, path, true, issues)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	var issues []*Issue
	if err := json.NewDecoder(f).Decode(&issues); err != nil {
		t.Fatal(err)
	}

	return issues
}
