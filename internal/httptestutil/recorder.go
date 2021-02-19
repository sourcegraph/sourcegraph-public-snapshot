package httptestutil

import (
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dnaeon/go-vcr/cassette"
	"github.com/dnaeon/go-vcr/recorder"

	"github.com/sourcegraph/sourcegraph/internal/httpcli"
)

// NewRecorder returns an HTTP interaction recorder with the given record mode and filters. It strips away the HTTP Authorization and Set-Cookie headers.
func NewRecorder(file string, record bool, filters ...cassette.Filter) (*recorder.Recorder, error) {
	mode := recorder.ModeReplaying
	if record {
		mode = recorder.ModeRecording
	}

	rec, err := recorder.NewAsMode(file, mode, nil)
	if err != nil {
		return nil, err
	}

	filters = append(filters, func(i *cassette.Interaction) error {
		delete(i.Request.Headers, "Authorization")
		// This is used for GitLab.
		delete(i.Request.Headers, "Private-Token")
		delete(i.Response.Headers, "Set-Cookie")
		return nil
	})

	for _, f := range filters {
		rec.AddFilter(f)
	}

	return rec, nil
}

// NewRecorderOpt returns an httpcli.Opt that wraps the Transport
// of an http.Client with the given recorder.
func NewRecorderOpt(rec *recorder.Recorder) httpcli.Opt {
	return func(c *http.Client) error {
		tr := c.Transport
		if tr == nil {
			tr = http.DefaultTransport
		}

		rec.SetTransport(tr)
		c.Transport = rec

		return nil
	}
}

// NewGitHubRecorderFactory returns a *http.Factory that rewrites HTTP requests to
// github-proxy to github.com and records all HTTP requests in "testdata/vcr/{name}"
// with {name} being the name that's passed in.
// If update is true, the HTTP requests are recorded, otherwise they're
// replayed from the recorded cassete.
func NewGitHubRecorderFactory(t testing.TB, update bool, name string) (*httpcli.Factory, func()) {
	t.Helper()

	cassete := filepath.Join("testdata/vcr/", strings.Replace(name, " ", "-", -1))

	rec, err := NewRecorder(cassete, update, func(i *cassette.Interaction) error {
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	mw := httpcli.NewMiddleware(httpcli.GitHubProxyRedirectMiddleware)

	hc := httpcli.NewFactory(mw, NewRecorderOpt(rec))

	return hc, func() {
		if err := rec.Stop(); err != nil {
			t.Errorf("failed to update test data: %s", err)
		}
	}
}
