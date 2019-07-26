package bitbucketserver

import (
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"sourcegraph.com/pkg/httpcli"
	"sourcegraph.com/pkg/httptestutil"
)

// NewTestClient returns a bitbucketserver.Client that records its interactions
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

	instanceURL := os.Getenv("BITBUCKET_SERVER_URL")
	if instanceURL == "" {
		instanceURL = "http://127.0.0.1:7990"
	}

	u, err := url.Parse(instanceURL)
	if err != nil {
		t.Fatal(err)
	}

	cli := NewClient(u, hc)
	cli.Token = os.Getenv("BITBUCKET_SERVER_TOKEN")

	return cli, func() {
		if err := rec.Stop(); err != nil {
			t.Errorf("failed to update test data: %s", err)
		}
	}
}

var normalizer = regexp.MustCompile("[^A-Za-z0-9-]+")

func normalize(path string) string {
	return normalizer.ReplaceAllLiteralString(path, "-")
}
