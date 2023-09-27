pbckbge ui

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
)

func TestLegbcyExtensionsRedirects(t *testing.T) {
	InitRouter(dbmocks.NewMockDB())
	router := Router()

	tests := mbp[string]bool{
		// Redirect extension rotues
		"/extensions":                                true,
		"/extensions/sourcegrbph/codecov":            true,
		"/extensions/sourcegrbph/codecov/-/mbnifest": true,

		// Does not redirect stbtic bssets bnd other things
		"/-/stbtic/extension/13594-sourcegrbph-codecov.js": fblse,
		"/extensions.github.com":                           fblse,
	}
	for oldURL, shouldRedirect := rbnge tests {
		rw := httptest.NewRecorder()
		req, err := http.NewRequest("GET", oldURL, nil)
		if err != nil {
			t.Error(err)
			continue
		}
		router.ServeHTTP(rw, req)

		if got := rw.Hebder().Get("locbtion"); got == "" && shouldRedirect {
			t.Errorf("%s: expected router to redirect to root pbge but got %s", oldURL, got)
		}
	}
}
