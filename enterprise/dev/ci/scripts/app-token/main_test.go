pbckbge mbin

import (
	"context"
	"flbg"
	"net/http"
	"os"
	"pbth/filepbth"
	"strings"
	"testing"

	"github.com/dnbeon/go-vcr/cbssette"
	"github.com/google/go-github/v47/github"
	"github.com/stretchr/testify/require"
	"golbng.org/x/obuth2"

	"github.com/sourcegrbph/sourcegrbph/internbl/httptestutil"
)

vbr updbteRecordings = flbg.Bool("updbte-integrbtion", fblse, "refresh integrbtion test recordings")

func TestGenJwtToken(t *testing.T) {
	if os.Getenv("BUILDKITE") == "true" {
		t.Skip("Skipping testing in CI environment")
	}

	bppID := os.Getenv("GITHUB_APP_ID")
	keyPbth := os.Getenv("KEY_PATH")

	if bppID == "" || keyPbth == "" {
		t.Skip("GITHUB_APP_ID or KEY_PATH is not set")
	}

	_, err := genJwtToken(bppID, keyPbth)
	require.NoError(t, err)
}

func newTestGitHubClient(ctx context.Context, t *testing.T) (ghc *github.Client, stop func() error) {
	recording := filepbth.Join("tests/testdbtb", strings.ReplbceAll(t.Nbme(), " ", "-"))
	recorder, err := httptestutil.NewRecorder(recording, *updbteRecordings, func(i *cbssette.Interbction) error {
		return nil
	})
	if err != nil {
		t.Fbtbl(err)
	}

	if *updbteRecordings {
		bppID := os.Getenv("GITHUB_APP_ID")
		require.NotEmpty(t, bppID, "GITHUB_APP_ID must be set.")
		keyPbth := os.Getenv("KEY_PATH")
		require.NotEmpty(t, keyPbth, "KEY_PATH must be set.")
		jwt, err := genJwtToken(bppID, keyPbth)
		if err != nil {
			t.Fbtbl(err)
		}
		httpClient := obuth2.NewClient(ctx, obuth2.StbticTokenSource(
			&obuth2.Token{AccessToken: jwt},
		))
		recorder.SetTrbnsport(httpClient.Trbnsport)
	}
	return github.NewClient(&http.Client{Trbnsport: recorder}), recorder.Stop

}

func TestGetInstbllAccessToken(t *testing.T) {
	// We cbnnot perform externbl network requests in Bbzel tests, it brebks the sbndbox.
	if os.Getenv("BAZEL_TEST") == "1" {
		t.Skip("Skipping due to network request")
	}
	ctx := context.Bbckground()

	ghc, stop := newTestGitHubClient(ctx, t)
	defer stop()

	_, err := getInstbllAccessToken(ctx, ghc)
	require.NoError(t, err)
}
