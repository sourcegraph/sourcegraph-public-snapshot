package ui

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
)

func TestLegacyExtensionsRedirects(t *testing.T) {
	InitRouter(dbmocks.NewMockDB())
	router := Router()

	tests := map[string]bool{
		// Redirect extension rotues
		"/extensions":                                true,
		"/extensions/sourcegraph/codecov":            true,
		"/extensions/sourcegraph/codecov/-/manifest": true,

		// Does not redirect static assets and other things
		"/-/static/extension/13594-sourcegraph-codecov.js": false,
		"/extensions.github.com":                           false,
	}
	for oldURL, shouldRedirect := range tests {
		rw := httptest.NewRecorder()
		req, err := http.NewRequest("GET", oldURL, nil)
		if err != nil {
			t.Error(err)
			continue
		}
		router.ServeHTTP(rw, req)

		if got := rw.Header().Get("location"); got == "" && shouldRedirect {
			t.Errorf("%s: expected router to redirect to root page but got %s", oldURL, got)
		}
	}
}
