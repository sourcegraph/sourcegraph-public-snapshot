package bitbucketcloud

import (
	"flag"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/httptestutil"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/schema"
)

var updateRegex = flag.String("update", "", "Update testdata of tests matching the given regex")

func update(name string) bool {
	if updateRegex == nil || *updateRegex == "" {
		return false
	}
	return regexp.MustCompile(*updateRegex).MatchString(name)
}

func GetenvTestBitbucketCloudUsername() string {
	username := os.Getenv("BITBUCKET_CLOUD_USERNAME")
	if username == "" {
		username = "sourcegraph-testing"
	}
	return username
}

// newTestClient returns a bitbucketcloud.Client that records its interactions
// to testdata/vcr/.
func newTestClient(t testing.TB) (*Client, func()) {
	t.Helper()

	cassete := filepath.Join("testdata/vcr/", normalize(t.Name()))
	rec, err := httptestutil.NewRecorder(cassete, update(t.Name()))
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
