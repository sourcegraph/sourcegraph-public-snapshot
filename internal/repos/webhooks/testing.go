package githubwebhook

import (
	"net/http"
	"net/url"
	"path/filepath"
	"testing"

	"github.com/dnaeon/go-vcr/cassette"

	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/httptestutil"
)

func NewTestClient(t testing.TB, name string, update *bool) *Client {
	t.Helper()

	casette := filepath.Join("testdata/vcr/", name)
	rec, err := httptestutil.NewRecorder(casette, *update)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := rec.Stop(); err != nil {
			t.Errorf("failed to update test data: %s", err)
		}
	})
	rec.SetMatcher(ignoreHostMatcher)

	hc, err := httpcli.NewFactory(nil, httptestutil.NewRecorderOpt(rec)).Doer()
	if err != nil {
		t.Fatal(err)
	}

	cli, err := NewClient(hc)
	if err != nil {
		t.Fatal(err)
	}

	return cli
}

func ignoreHostMatcher(r *http.Request, i cassette.Request) bool {
	if r.Method != i.Method {
		return false
	}
	u, err := url.Parse(i.URL)
	if err != nil {
		return false
	}
	u.Host = r.URL.Host
	u.Scheme = r.URL.Scheme
	return r.URL.String() == u.String()
}
