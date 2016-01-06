package pctx

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"golang.org/x/net/context"

	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/conf"
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

	appURL, err := url.Parse("https://src.foo.com")
	if err != nil {
		panic(err)
	}
	ctx := conf.WithURL(context.Background(), appURL, nil)
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

		expPrefix := appURL
		expPrefix.Path = test.expPrefix
		if prefix != expPrefix.String() {
			t.Errorf("expected prefix %s, got %s", test.expPrefix, prefix)
		}
	}
}
