package ui

import (
	"bytes"
	"net/http"
	"net/http/httptest"
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

// random will create a file of size bytes (rounded up to next 1024 size)
func random_292(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
