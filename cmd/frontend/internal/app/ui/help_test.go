package ui

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/version"
)

func TestServeHelp(t *testing.T) {
	t.Run("unreleased dev version", func(t *testing.T) {
		{
			orig := envvar.SourcegraphDotComMode()
			envvar.MockSourcegraphDotComMode(false)
			defer envvar.MockSourcegraphDotComMode(orig) // reset
		}
		{
			orig := version.Version()
			version.Mock("0.0.0+dev")
			defer version.Mock(orig) // reset
		}

		rw := httptest.NewRecorder()
		rw.Body = new(bytes.Buffer)
		req, _ := http.NewRequest("GET", "/help/foo/bar", nil)
		serveHelp(rw, req)

		if want := http.StatusTemporaryRedirect; rw.Code != want {
			t.Errorf("got %d, want %d", rw.Code, want)
		}
		if got, want := rw.Header().Get("Location"), "http://localhost:5080/foo/bar"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("released version", func(t *testing.T) {
		{
			orig := envvar.SourcegraphDotComMode()
			envvar.MockSourcegraphDotComMode(false)
			defer envvar.MockSourcegraphDotComMode(orig) // reset
		}
		{
			orig := version.Version()
			version.Mock("3.39.1")
			defer version.Mock(orig) // reset
		}

		rw := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/help/dev", nil)
		serveHelp(rw, req)

		if want := http.StatusTemporaryRedirect; rw.Code != want {
			t.Errorf("got %d, want %d", rw.Code, want)
		}
		if got, want := rw.Header().Get("Location"), "https://docs.sourcegraph.com/@3.39/dev"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("Sourcegraph.com", func(t *testing.T) {
		orig := envvar.SourcegraphDotComMode()
		envvar.MockSourcegraphDotComMode(true)
		defer envvar.MockSourcegraphDotComMode(orig) // reset

		rw := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/help/foo/bar", nil)
		serveHelp(rw, req)

		if want := http.StatusTemporaryRedirect; rw.Code != want {
			t.Errorf("got %d, want %d", rw.Code, want)
		}
		if got, want := rw.Header().Get("Location"), "https://docs.sourcegraph.com/foo/bar"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}
