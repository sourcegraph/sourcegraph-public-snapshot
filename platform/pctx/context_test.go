package pctx

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/conf"
)

// Test_repoFrameBaseURI tests that RepoFrameBaseURI returns the correct URL
// prefix for a repository frame app URI.
func Test_repoFrameBaseURI(t *testing.T) {
	tests := []struct {
		url       string
		expPrefix string
	}{{
		url:       "/github.com/gorilla/mux/-/app/issues",
		expPrefix: "/github.com/gorilla/mux/-/app/issues",
	}, {
		url:       "/github.com/gorilla/mux/-/app/issues/",
		expPrefix: "/github.com/gorilla/mux/-/app/issues",
	}, {
		url:       "/github.com/gorilla/mux/-/app/issues/foo",
		expPrefix: "/github.com/gorilla/mux/-/app/issues",
	}, {
		url:       "/github.com/gorilla/mux@branch/-/app/issues/foo",
		expPrefix: "/github.com/gorilla/mux@branch/-/app/issues",
	}, {
		url:       "/github.com/gorilla/mux@aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa/-/app/issues/foo",
		expPrefix: "/github.com/gorilla/mux@aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa/-/app/issues",
	}, {
		url:       "/github.com/gorilla/mux@branch===aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa/-/app/issues/foo",
		expPrefix: "/github.com/gorilla/mux@branch===aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa/-/app/issues",
	}}

	appURL, err := url.Parse("https://src.foo.com")
	if err != nil {
		panic(err)
	}
	ctx := conf.WithURL(context.Background(), appURL)
	rtr := router.New(nil)
	for _, test := range tests {
		var prefix string
		rtr.Get(router.RepoAppFrame).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p, err := repoFrameBaseURI(ctx, r)
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
