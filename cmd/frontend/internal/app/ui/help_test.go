package ui

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/pkg/version"
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
			version.Mock("dev")
			defer version.Mock(orig) // reset
		}

		rw := httptest.NewRecorder()
		rw.Body = new(bytes.Buffer)
		req, _ := http.NewRequest("GET", "/help/foo/bar", nil)
		serveHelp(rw, req)

		if want := http.StatusNotImplemented; rw.Code != want {
			t.Errorf("got %d, want %d", rw.Code, want)
		}
		if got, want := rw.Body.String(), "<a href=\"https://docs.sourcegraph.com/foo/bar"; !strings.Contains(got, want) {
			t.Errorf("got %q, want to contain %q", got, want)
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
			version.Mock("1.2.3")
			defer version.Mock(orig) // reset
		}

		rw := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/help/foo/bar", nil)
		serveHelp(rw, req)

		if want := http.StatusTemporaryRedirect; rw.Code != want {
			t.Errorf("got %d, want %d", rw.Code, want)
		}
		if got, want := rw.Header().Get("Location"), "https://docs.sourcegraph.com/@v1.2.3/foo/bar"; got != want {
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
