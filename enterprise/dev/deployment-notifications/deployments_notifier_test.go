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
	"github.com/google/go-github/v41/github"
	"github.com/sourcegraph/sourcegraph/internal/httptestutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

var updateRecordings = flag.Bool("update", false, "update integration test")

var changedFiles = []string{
	".buildkite/kubeval.sh",
	".buildkite/verify-yaml.sh",
	"base/frontend/sourcegraph-frontend-internal.Deployment.yaml",
	"base/frontend/sourcegraph-frontend.Deployment.yaml",
	"base/github-proxy/github-proxy.Deployment.yaml",
	"base/migrator/migrator.Job.yaml",
	"base/precise-code-intel/worker.Deployment.yaml",
	"base/searcher/searcher.Deployment.yaml",
	"base/symbols/symbols.Deployment.yaml",
	"base/syntect-server/syntect-server.Deployment.yaml",
	"base/worker/worker.Deployment.yaml",
	"configure/postgres-exporter/postgres-exporter-codeintel.Deployment.yaml",
	"configure/postgres-exporter/postgres-exporter.Deployment.yaml",
}

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

		changedFiles := []string{
			"base/frontend/sourcegraph-frontend-internal.Deployment.yaml",
			"base/frontend/sourcegraph-frontend.Deployment.yaml",
			"base/precise-code-intel/worker.Deployment.yaml",
			"base/searcher/searcher.Deployment.yaml",
			"base/symbols/symbols.Deployment.yaml",
		}
		expectedPRs := []int{32512, 32533, 32516, 32525}
		expectedApps := []string{
			"sourcegraph-frontend-internal",
			"sourcegraph-frontend",
			"worker",
			"searcher",
			"symbols",
		}

		dn := NewDeploymentNotifier(
			ghc,
			NewMockVersionRequester("cd5f80783501c433474266b57cbf1dc1a9f3a652", nil),
			"b7e9d07610591061044b60709bc78205047067b7",
			changedFiles,
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
		assert.EqualValues(t, expectedApps, report.Apps)
	})

	t.Run("OK no relevant changed files", func(t *testing.T) {
		ghc, stop := newTestGitHubClient(ctx, t)
		defer stop()

		changedFiles := []string{
			".buildkite/kubeval.sh",
			".buildkite/verify-yaml.sh",
			"base/migrator/migrator.Job.yaml",
		}
		expectedApps := []string(nil)

		dn := NewDeploymentNotifier(
			ghc,
			NewMockVersionRequester("cd5f80783501c433474266b57cbf1dc1a9f3a652", nil),
			"b7e9d07610591061044b60709bc78205047067b7",
			changedFiles,
		)
		report, err := dn.Report(ctx)
		if err != nil {
			t.Fatal(err)
		}

		assert.EqualValues(t, expectedApps, report.Apps)
	})

	t.Run("OK single commit", func(t *testing.T) {
		ghc, stop := newTestGitHubClient(ctx, t)
		defer stop()

		changedFiles := []string{
			"base/frontend/sourcegraph-frontend-internal.Deployment.yaml",
			"base/frontend/sourcegraph-frontend.Deployment.yaml",
			"base/precise-code-intel/worker.Deployment.yaml",
			"base/searcher/searcher.Deployment.yaml",
			"base/symbols/symbols.Deployment.yaml",
		}
		expectedPRs := []int{32512}
		expectedApps := []string{
			"sourcegraph-frontend-internal",
			"sourcegraph-frontend",
			"worker",
			"searcher",
			"symbols",
		}

		dn := NewDeploymentNotifier(
			ghc,
			NewMockVersionRequester("c9e25b6eef532bccc47c9dfe149265abad939239", nil),
			"b7e9d07610591061044b60709bc78205047067b7",
			changedFiles,
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
		assert.EqualValues(t, expectedApps, report.Apps)
	})

	t.Run("NOK deploying twice", func(t *testing.T) {
		ghc, stop := newTestGitHubClient(ctx, t)
		defer stop()

		changedFiles := []string{
			"base/frontend/sourcegraph-frontend-internal.Deployment.yaml",
			"base/frontend/sourcegraph-frontend.Deployment.yaml",
			"base/precise-code-intel/worker.Deployment.yaml",
			"base/searcher/searcher.Deployment.yaml",
			"base/symbols/symbols.Deployment.yaml",
		}

		dn := NewDeploymentNotifier(
			ghc,
			NewMockVersionRequester("cd5f80783501c433474266b57cbf1dc1a9f3a652", nil),
			"cd5f80783501c433474266b57cbf1dc1a9f3a652",
			changedFiles,
		)
		_, err := dn.Report(ctx)
		assert.NotNil(t, err)
		assert.True(t, errors.Is(err, ErrAlreadyDeployed))
	})

	t.Run("NOK failed to request current version", func(t *testing.T) {
		dn := NewDeploymentNotifier(
			nil,
			NewMockVersionRequester("cd5f80783501c433474266b57cbf1dc1a9f3a652", errors.New("500")),
			"cd5f80783501c433474266b57cbf1dc1a9f3a652",
			nil,
		)
		_, err := dn.Report(ctx)
		assert.NotNil(t, err)
	})
}
