package ui

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestLegacyExtensionsRedirects(t *testing.T) {
	InitRouter(database.NewMockDB(), nil)
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

func TestLegacyExtensionsRedirectsWithExtensionsEnabled(t *testing.T) {
	enableLegacyExtensions()
	defer conf.Mock(nil)

	InitRouter(database.NewMockDB(), nil)
	router := Router()

	tests := []string{
		"/extensions",
		"/extensions/sourcegraph/codecov",
		"/extensions/sourcegraph/codecov/-/manifest",
		"/-/static/extension/13594-sourcegraph-codecov.js",
	}
	for i := range tests {
		rw := httptest.NewRecorder()
		req, err := http.NewRequest("GET", tests[i], nil)
		if err != nil {
			t.Error(err)
			continue
		}
		router.ServeHTTP(rw, req)

		if got := rw.Header().Get("location"); got != "" {
			t.Errorf("%s: expected router to not redirect to root page but got %s", tests[i], got)
		}
	}
}

func enableLegacyExtensions() {
	enableLegacyExtensionsVar := true
	conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{
		ExperimentalFeatures: &schema.ExperimentalFeatures{
			EnableLegacyExtensions: &enableLegacyExtensionsVar,
		},
	}})
}
