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
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/sourcegraph/sourcegraph/internal/httptestutil"
)

var updateRecordings = flag.Bool("update-integration", false, "refresh integration test recordings")

func TestGenJwtToken(t *testing.T) {
	if os.Getenv("BUILDKITE") == "true" {
		t.Skip("Skipping testing in CI environment")
	}

	appID := os.Getenv("GITHUB_APP_ID")
	keyPath := os.Getenv("KEY_PATH")

	if appID == "" || keyPath == "" {
		t.Skip("GITHUB_APP_ID or KEY_PATH is not set")
	}

	_, err := genJwtToken(appID, keyPath)
	require.NoError(t, err)
}

func newTestGitHubClient(ctx context.Context, t *testing.T) (ghc *github.Client, stop func() error) {
	recording := filepath.Join("tests/testdata", strings.ReplaceAll(t.Name(), " ", "-"))
	recorder, err := httptestutil.NewRecorder(recording, *updateRecordings, func(i *cassette.Interaction) error {
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if *updateRecordings {
		appID := os.Getenv("GITHUB_APP_ID")
		require.NotEmpty(t, appID, "GITHUB_APP_ID must be set.")
		keyPath := os.Getenv("KEY_PATH")
		require.NotEmpty(t, keyPath, "KEY_PATH must be set.")
		jwt, err := genJwtToken(appID, keyPath)
		if err != nil {
			t.Fatal(err)
		}
		httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: jwt},
		))
		recorder.SetTransport(httpClient.Transport)
	}
	return github.NewClient(&http.Client{Transport: recorder}), recorder.Stop
}

func TestGetInstallAccessToken(t *testing.T) {
	// We cannot perform external network requests in Bazel tests, it breaks the sandbox.
	if os.Getenv("BAZEL_TEST") == "1" {
		t.Skip("Skipping due to network request")
	}
	ctx := context.Background()

	ghc, stop := newTestGitHubClient(ctx, t)
	defer stop()

	_, err := getInstallAccessToken(ctx, ghc)
	require.NoError(t, err)
}
