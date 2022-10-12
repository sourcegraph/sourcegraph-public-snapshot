package main

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dnaeon/go-vcr/cassette"
	"github.com/google/go-github/v47/github"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/sourcegraph/sourcegraph/internal/httptestutil"
)

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

func TestGenJwtToken(t *testing.T) {

	appID := os.Getenv("GITHUB_APP_ID")
	require.NotEmpty(t, appID, "GITHUB_APP_ID must be set.")
	keyPath := os.Getenv("KEY_PATH")
	require.NotEmpty(t, keyPath, "KEY_PATH must be set.")

	jwt, err := genJwtToken(appID, keyPath)
	require.NoError(t, err)
	t.Log("%+s", jwt)
}

func TestGetInstallAccessToken(t *testing.T) {

}
