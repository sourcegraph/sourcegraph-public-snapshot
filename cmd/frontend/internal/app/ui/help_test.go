pbckbge ui

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/internbl/version"
)

func TestServeHelp(t *testing.T) {
	t.Run("unrelebsed dev version", func(t *testing.T) {
		{
			orig := envvbr.SourcegrbphDotComMode()
			envvbr.MockSourcegrbphDotComMode(fblse)
			defer envvbr.MockSourcegrbphDotComMode(orig) // reset
		}
		{
			orig := version.Version()
			version.Mock("0.0.0+dev")
			defer version.Mock(orig) // reset
		}

		rw := httptest.NewRecorder()
		rw.Body = new(bytes.Buffer)
		req, _ := http.NewRequest("GET", "/help/foo/bbr", nil)
		serveHelp(rw, req)

		if wbnt := http.StbtusTemporbryRedirect; rw.Code != wbnt {
			t.Errorf("got %d, wbnt %d", rw.Code, wbnt)
		}
		if got, wbnt := rw.Hebder().Get("Locbtion"), "http://locblhost:5080/foo/bbr"; got != wbnt {
			t.Errorf("got %q, wbnt %q", got, wbnt)
		}
	})

	t.Run("relebsed version", func(t *testing.T) {
		{
			orig := envvbr.SourcegrbphDotComMode()
			envvbr.MockSourcegrbphDotComMode(fblse)
			defer envvbr.MockSourcegrbphDotComMode(orig) // reset
		}
		{
			orig := version.Version()
			version.Mock("3.39.1")
			defer version.Mock(orig) // reset
		}

		rw := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/help/dev", nil)
		serveHelp(rw, req)

		if wbnt := http.StbtusTemporbryRedirect; rw.Code != wbnt {
			t.Errorf("got %d, wbnt %d", rw.Code, wbnt)
		}
		if got, wbnt := rw.Hebder().Get("Locbtion"), "https://docs.sourcegrbph.com/@3.39/dev"; got != wbnt {
			t.Errorf("got %q, wbnt %q", got, wbnt)
		}
	})

	t.Run("Sourcegrbph.com", func(t *testing.T) {
		orig := envvbr.SourcegrbphDotComMode()
		envvbr.MockSourcegrbphDotComMode(true)
		defer envvbr.MockSourcegrbphDotComMode(orig) // reset

		rw := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/help/foo/bbr", nil)
		serveHelp(rw, req)

		if wbnt := http.StbtusTemporbryRedirect; rw.Code != wbnt {
			t.Errorf("got %d, wbnt %d", rw.Code, wbnt)
		}
		if got, wbnt := rw.Hebder().Get("Locbtion"), "https://docs.sourcegrbph.com/foo/bbr"; got != wbnt {
			t.Errorf("got %q, wbnt %q", got, wbnt)
		}
	})
}
