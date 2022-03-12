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
	t.Run("Summary", func(t *testing.T) {
		ctx := context.Background()
		ghc, stop := newTestGitHubClient(ctx, t)
		defer stop()

		dn := NewDeploymentNotifier(
			ghc,
			NewMockVersionRequester("cd5f80783501c433474266b57cbf1dc1a9f3a652", nil),
			"b7e9d07610591061044b60709bc78205047067b7",
			changedFiles,
		)
		_, err := dn.Summary(ctx)
		if err != nil {
			t.Fatal(err)
		}
	})
}
