package pctx

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
)

// Test_repoFrameBaseURI tests that RepoFrameBaseURI returns the correct URL
// prefix for a repository frame app URI.
func Test_repoFrameBaseURI(t *testing.T) {
	tests := []struct {
		url       string
		expPrefix string
	}{{
		url:       "/github.com/gorilla/mux/.issues",
		expPrefix: "/github.com/gorilla/mux/.issues",
	}, {
		url:       "/github.com/gorilla/mux/.issues/",
		expPrefix: "/github.com/gorilla/mux/.issues",
	}, {
		url:       "/github.com/gorilla/mux/.issues/foo",
		expPrefix: "/github.com/gorilla/mux/.issues",
	}, {
		url:       "/github.com/gorilla/mux/.issues@branch/foo",
		expPrefix: "/github.com/gorilla/mux/.issues@branch",
	}, {
		url:       "/github.com/gorilla/mux/.issues@aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa/foo",
		expPrefix: "/github.com/gorilla/mux/.issues@aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}, {
		url:       "/github.com/gorilla/mux/.issues@branch===aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa/foo",
		expPrefix: "/github.com/gorilla/mux/.issues@branch===aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}}

	rtr := router.New(nil)
	for _, test := range tests {
		var prefix string
		rtr.Get(router.RepoAppFrame).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p, err := repoFrameBaseURI(r)
			if err != nil {
				t.Errorf("unexpected error computing repo frame base URL: %s", err)
				return
			}
			prefix = p
		})

		req, _ := http.NewRequest("GET", test.url, nil)
		rw := httptest.NewRecorder()
		rtr.ServeHTTP(rw, req)

		if prefix != test.expPrefix {
			t.Errorf("expected prefix %s, got %s", test.expPrefix, prefix)
		}
	}
}
