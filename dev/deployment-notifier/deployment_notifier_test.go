package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dnaeon/go-vcr/cassette"
	"github.com/google/go-github/v55/github"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"

	"github.com/sourcegraph/sourcegraph/internal/httptestutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var updateRecordings = flag.Bool("update", false, "update integration test")

func newTestGitHubClient(ctx context.Context, t *testing.T) (ghc *github.Client, stop func() error) {
	recording := filepath.Join("testdata", strings.ReplaceAll(t.Name(), " ", "-"))
	recorder, err := httptestutil.NewRecorder(recording, *updateRecordings, func(i *cassette.Interaction) error {
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if *updateRecordings {
		httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
		))
		recorder.SetTransport(httpClient.Transport)
	}
	return github.NewClient(&http.Client{Transport: recorder}), recorder.Stop
}

func TestDeploymentNotifier(t *testing.T) {
	ctx := context.Background()
	t.Run("OK normal", func(t *testing.T) {
		ghc, stop := newTestGitHubClient(ctx, t)
		defer stop()

		expectedPRs := []int{32996, 32871, 32767}
		expectedServices := []string{
			"frontend",
			"gitserver",
			"searcher",
			"symbols",
			"worker",
		}
		expectedServicesPerPullRequest := map[int][]string{
			32996: {"frontend", "gitserver", "searcher", "symbols", "worker"},
			32871: {"frontend", "gitserver", "searcher", "symbols", "worker"},
			32767: {"gitserver"},
		}

		newCommit := "e1aea6f8d82283695ae4a3b2b5a7a8f36b1b934b"
		oldCommit := "54d527f7f7b5770e0dfd1f56398bf8a2f30b935d"
		olderCommit := "99db56d45299161d3bf62677ba3d3ab701910bb0"

		m := map[string]*ServiceVersionDiff{
			"frontend": {Old: oldCommit, New: newCommit},
			"worker":   {Old: oldCommit, New: newCommit},
			"searcher": {Old: oldCommit, New: newCommit},
			"symbols":  {Old: oldCommit, New: newCommit},
			// This one is older by one PR.
			"gitserver": {Old: olderCommit, New: newCommit},
		}

		dn := NewDeploymentNotifier(
			ghc,
			NewMockManifestDeployementsDiffer(m),
			"tests",
			"",
		)
		report, err := dn.Report(ctx)
		if err != nil {
			t.Fatal(err)
		}

		var prNumbers []int
		for _, pr := range report.PullRequests {
			prNumbers = append(prNumbers, pr.GetNumber())
		}
		assert.EqualValues(t, expectedPRs, prNumbers)
		assert.EqualValues(t, expectedServices, report.Services)
		assert.EqualValues(t, expectedServicesPerPullRequest, report.ServicesPerPullRequest)
	})

	t.Run("OK no relevant changed files", func(t *testing.T) {
		ghc, stop := newTestGitHubClient(ctx, t)
		defer stop()

		m := map[string]*ServiceVersionDiff{}

		dn := NewDeploymentNotifier(
			ghc,
			NewMockManifestDeployementsDiffer(m),
			"tests",
			"",
		)

		_, err := dn.Report(ctx)
		assert.NotNil(t, err)
		assert.True(t, errors.Is(err, ErrNoRelevantChanges))
	})

	t.Run("OK single commit", func(t *testing.T) {
		ghc, stop := newTestGitHubClient(ctx, t)
		defer stop()

		expectedPRs := []int{32996}
		expectedServices := []string{
			"frontend",
			"searcher",
			"symbols",
			"worker",
		}

		newCommit := "e1aea6f8d82283695ae4a3b2b5a7a8f36b1b934b"
		oldCommit := "68374f229042704f1663ca2fd19401ba0772c828"

		m := map[string]*ServiceVersionDiff{
			"frontend": {Old: oldCommit, New: newCommit},
			"worker":   {Old: oldCommit, New: newCommit},
			"searcher": {Old: oldCommit, New: newCommit},
			"symbols":  {Old: oldCommit, New: newCommit},
		}

		dn := NewDeploymentNotifier(
			ghc,
			NewMockManifestDeployementsDiffer(m),
			"tests",
			"",
		)

		report, err := dn.Report(ctx)
		if err != nil {
			t.Fatal(err)
		}

		var prNumbers []int
		for _, pr := range report.PullRequests {
			prNumbers = append(prNumbers, pr.GetNumber())
		}
		assert.EqualValues(t, expectedPRs, prNumbers)
		assert.EqualValues(t, expectedServices, report.Services)
	})

	t.Run("NOK deploying twice", func(t *testing.T) {
		ghc, stop := newTestGitHubClient(ctx, t)
		defer stop()

		newCommit := "e1aea6f8d82283695ae4a3b2b5a7a8f36b1b934b"

		m := map[string]*ServiceVersionDiff{
			"frontend": {Old: newCommit, New: newCommit},
			"worker":   {Old: newCommit, New: newCommit},
			"searcher": {Old: newCommit, New: newCommit},
			"symbols":  {Old: newCommit, New: newCommit},
		}

		dn := NewDeploymentNotifier(
			ghc,
			NewMockManifestDeployementsDiffer(m),
			"tests",
			"",
		)
		_, err := dn.Report(ctx)
		assert.NotNil(t, err)
		assert.True(t, errors.Is(err, ErrNoRelevantChanges))
	})
}

func TestParsePRNumberInMergeCommit(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    int
	}{
		{name: "Merge commit with revert", message: `Revert "Support diffing for unrelated commits. (#32015)" (#32737)`, want: 32737},
		{name: "Normal commit", message: `YOLO I commit on main without PR`, want: 0},
		{name: "Merge commit", message: `batches: Properly quote name in YAML (#32951)`, want: 32951},
		{name: "Merge commit with additional desc", message: `Fix repopendingperms tests (#33247)

* Fix repopendingperms tests`, want: 33247},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := parsePRNumberInMergeCommit(test.message)
			assert.Equal(t, test.want, got)
		})
	}
}
