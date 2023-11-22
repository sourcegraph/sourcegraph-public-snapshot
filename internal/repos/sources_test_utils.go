package repos

import (
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/grafana/regexp"

	"github.com/dnaeon/go-vcr/cassette"
	"github.com/dnaeon/go-vcr/recorder"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/httptestutil"
)

func NewClientFactory(t testing.TB, name string, mws ...httpcli.Middleware) (*httpcli.Factory, func(testing.TB)) {
	mw, rec := TestClientFactorySetup(t, name, mws...)
	return httpcli.NewFactory(mw, httptestutil.NewRecorderOpt(rec)),
		func(t testing.TB) { Save(t, rec) }
}

func Save(t testing.TB, rec *recorder.Recorder) {
	if err := rec.Stop(); err != nil {
		t.Errorf("failed to update test data: %s", err)
	}
}

var updateRegex *string

func Update(name string) bool {
	if updateRegex == nil || *updateRegex == "" {
		return false
	}
	return regexp.MustCompile(*updateRegex).MatchString(name)
}

func TestClientFactorySetup(t testing.TB, name string, mws ...httpcli.Middleware) (httpcli.Middleware, *recorder.Recorder) {
	cassete := filepath.Join("testdata", "sources", strings.ReplaceAll(name, " ", "-"))
	rec := NewRecorder(t, cassete, Update(name))
	mws = append(mws, GitserverRedirectMiddleware)
	mw := httpcli.NewMiddleware(mws...)
	return mw, rec
}

func GitserverRedirectMiddleware(cli httpcli.Doer) httpcli.Doer {
	return httpcli.DoerFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Hostname() == "gitserver" {
			// Start local git server first
			req.URL.Host = "127.0.0.1:3178"
			req.URL.Scheme = "http"
		}
		return cli.Do(req)
	})
}

func NewRecorder(t testing.TB, file string, record bool) *recorder.Recorder {
	rec, err := httptestutil.NewRecorder(file, record, func(i *cassette.Interaction) error {
		// The ratelimit.Monitor type resets its internal timestamp if it's
		// updated with a timestamp in the past. This makes tests ran with
		// recorded interations just wait for a very long time. Removing
		// these headers from the casseste effectively disables rate-limiting
		// in tests which replay HTTP interactions, which is desired behaviour.
		for _, name := range [...]string{
			"RateLimit-Limit",
			"RateLimit-Observed",
			"RateLimit-Remaining",
			"RateLimit-Reset",
			"RateLimit-Resettime",
			"X-RateLimit-Limit",
			"X-RateLimit-Remaining",
			"X-RateLimit-Reset",
		} {
			i.Response.Headers.Del(name)
		}

		// Phabricator requests include a token in the form and body.
		ua := i.Request.Headers.Get("User-Agent")
		if strings.Contains(strings.ToLower(ua), extsvc.TypePhabricator) {
			i.Request.Body = ""
			i.Request.Form = nil
		}

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	return rec
}
