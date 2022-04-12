package bitbucketcloud

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/httptestutil"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/schema"
)

func GetenvTestBitbucketCloudUsername() string {
	username := os.Getenv("BITBUCKET_CLOUD_USERNAME")
	if username == "" {
		username = "sourcegraph-testing"
	}
	return username
}

// NewTestClient returns a bitbucketcloud.Client that records its interactions
// to testdata/vcr/.
func NewTestClient(t testing.TB, name string, update bool) (*Client, func()) {
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

	cli, err := NewClient("urn", &schema.BitbucketCloudConnection{
		ApiURL:      "https://api.bitbucket.org",
		Username:    GetenvTestBitbucketCloudUsername(),
		AppPassword: os.Getenv("BITBUCKET_CLOUD_APP_PASSWORD"),
	}, hc)
	if err != nil {
		t.Fatal(err)
	}

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
