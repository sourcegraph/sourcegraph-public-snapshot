package bitbucketcloud

import (
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/httptestutil"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
)

func GetenvTestBitbucketCloudUsername() string {
	username := os.Getenv("BITBUCKET_CLOUD_USERNAME")
	if username == "" {
		username = "unknwon"
	}
	return username
}

// NewTestClient returns a bitbucketcloud.Client that records its interactions
// to testdata/vcr/.
func NewTestClient(t testing.TB, name string, update bool, apiURL *url.URL) (*Client, func()) {
	t.Helper()

	cassete := filepath.Join("testdata/vcr/", normalize(name))
	rec, err := httptestutil.NewRecorder(cassete, update)
	if err != nil {
		t.Fatal(err)
	}

	hc, err := httpcli.NewFactory(nil, httptestutil.NewRecorderOpt(rec)).Doer()
	if err != nil {
		t.Fatal(err)
	}

	cli := NewClient(apiURL, hc)
	cli.Username = GetenvTestBitbucketCloudUsername()
	cli.AppPassword = os.Getenv("BITBUCKET_CLOUD_APP_PASSWORD")

	return cli, func() {
		if err := rec.Stop(); err != nil {
			t.Errorf("failed to update test data: %s", err)
		}
	}
}

var normalizer = lazyregexp.New("[^A-Za-z0-9-]+")

func normalize(path string) string {
	return normalizer.ReplaceAllLiteralString(path, "-")
}
