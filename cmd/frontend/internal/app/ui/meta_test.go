package ui

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/mux"
)

func TestCanonicalRepoURL(t *testing.T) {
	appURL := &url.URL{Scheme: "https", Host: "sourcegraph.com", Path: "/"}
	const (
		repoDefaultBranch = "b"
		resolvedCommitID  = "c"
	)

	tests := map[string]string{
		"/r?utm_source=a": "/r",

		"/r":   "/r",
		"/r@v": "/r@c",
		"/r@b": "/r",
		"/r@c": "/r@c",

		"/r/-/blob/f":   "/r/-/blob/f",
		"/r@v/-/blob/f": "/r@c/-/blob/f",
		"/r@b/-/blob/f": "/r/-/blob/f",
		"/r@c/-/blob/f": "/r@c/-/blob/f",

		"/r/-/tree/f":   "/r/-/tree/f",
		"/r@v/-/tree/f": "/r@c/-/tree/f",
		"/r@b/-/tree/f": "/r/-/tree/f",
		"/r@c/-/tree/f": "/r@c/-/tree/f",

		"/r/-/info/t/u/-/p":   "/r/-/info/t/u/-/p",
		"/r@v/-/info/t/u/-/p": "/r@c/-/info/t/u/-/p",
		"/r@b/-/info/t/u/-/p": "/r/-/info/t/u/-/p",
		"/r@c/-/info/t/u/-/p": "/r@c/-/info/t/u/-/p",

		"/r/-/def/t/u/-/p":   "/r/-/def/t/u/-/p",
		"/r@v/-/def/t/u/-/p": "/r@c/-/def/t/u/-/p",
		"/r@b/-/def/t/u/-/p": "/r/-/def/t/u/-/p",
		"/r@c/-/def/t/u/-/p": "/r@c/-/def/t/u/-/p",
	}
	for orig, want := range tests {
		origURL, err := url.Parse(orig)
		if err != nil {
			t.Fatalf("%s: %s", orig, err)
		}
		var routeMatch mux.RouteMatch
		match := router.Match(&http.Request{Method: "GET", URL: origURL}, &routeMatch)
		if !match || routeMatch.Route == nil {
			t.Fatalf("%s: no match", orig)
		}

		canon := canonicalRepoURL(appURL, routeMatch.Route.GetName(), routeMatch.Vars, origURL.Query(), repoDefaultBranch, resolvedCommitID)
		want = strings.TrimSuffix(appURL.String(), "/") + want
		if canon != want {
			t.Errorf("%s: got %s, want %s", orig, canon, want)
		}
	}
}
