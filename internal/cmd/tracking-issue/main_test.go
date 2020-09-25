package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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
			Number:    13987,
			Milestone: "3.21",
			Labels:    []string{"tracking", "team/code-intelligence"},
		},
	}

	// Mock current time to be the last update
	lastUpdate, err := getOrUpdateLastUpdateTime(*update)
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}
	now = func() time.Time { return lastUpdate }

	loadTrackingIssueFixtures(t, "sourcegraph", ti)
	got := ti.Workloads().Markdown(ti.LabelAllowlist)
	testutil.AssertGolden(t, filepath.Join("testdata", "issue.md"), *update, got)
}

func getOrUpdateLastUpdateTime(update bool) (time.Time, error) {
	lastUpdateFile := filepath.Join("testdata", "last-update.txt")

	if update {
		now := time.Now().UTC()

		if err := ioutil.WriteFile(
			lastUpdateFile,
			[]byte(now.Format(time.RFC3339)),
			os.ModePerm,
		); err != nil {
			return time.Time{}, err
		}

		return now, nil
	}

	content, err := ioutil.ReadFile(lastUpdateFile)
	if err != nil {
		return time.Time{}, err
	}

	return time.Parse(time.RFC3339, string(content))
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

		var q strings.Builder
		fmt.Fprintf(&q, "org:sourcegraph label:tracking is:open")
		for _, label := range issue.Labels {
			fmt.Fprintf(&q, " label:%s", label)
		}

		tracking, err := listTrackingIssues(ctx, cli, q.String())
		if err != nil {
			t.Fatal(err)
		}

		found := false
		for _, ti := range tracking {
			if ti.Number == issue.Number {
				issue = ti
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("unknown tracking issue %d\n", issue.Number)
		}

		err = loadTrackingIssues(ctx, cli, org, []*TrackingIssue{issue})
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

	issue.FillLabelAllowlist()
}
