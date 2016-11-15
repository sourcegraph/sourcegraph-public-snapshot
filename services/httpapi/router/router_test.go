package router

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/kr/pretty"
)

func TestRouter(t *testing.T) {
	router := New(nil)
	tests := []struct {
		path          string
		wantNoMatch   bool
		wantRouteName string
		wantVars      map[string]string
		wantPath      string
	}{
		// Tree
		{
			path:          "/repos/r@v/-/tree",
			wantRouteName: RepoTree,
			wantVars:      map[string]string{"Repo": "r", "Rev": "@v", "Path": ""},
		},
		{
			path:          "/repos/r@v/-/tree/a/b.txt",
			wantRouteName: RepoTree,
			wantVars:      map[string]string{"Repo": "r", "Rev": "@v", "Path": "/a/b.txt"},
		},
	}
	for _, test := range tests {
		var routeMatch mux.RouteMatch
		match := router.Match(&http.Request{Method: "GET", URL: &url.URL{Path: test.path}}, &routeMatch)

		if match && test.wantNoMatch {
			t.Errorf("%s: got match (route %q), want no match", test.path, routeMatch.Route.GetName())
		}
		if !match && !test.wantNoMatch {
			t.Errorf("%s: got no match, wanted match", test.path)
		}
		if !match || test.wantNoMatch {
			continue
		}

		if routeName := routeMatch.Route.GetName(); routeName != test.wantRouteName {
			t.Errorf("%s: got matched route %q, want %q", test.path, routeName, test.wantRouteName)
		}

		if diff := pretty.Diff(routeMatch.Vars, test.wantVars); len(diff) > 0 {
			t.Errorf("%s: vars don't match expected:\n%s", test.path, strings.Join(diff, "\n"))
		}

		// Check that building the URL yields the original path.
		var pairs []string
		for k, v := range test.wantVars {
			pairs = append(pairs, k, v)
		}
		path, err := routeMatch.Route.URLPath(pairs...)
		if err != nil {
			t.Errorf("%s: URLPath(%v) failed: %s", test.path, pairs, err)
			continue
		}
		var wantPath string
		if test.wantPath != "" {
			wantPath = test.wantPath
		} else {
			wantPath = test.path
		}
		if path.Path != wantPath {
			t.Errorf("got generated path %q, want %q", path, wantPath)
		}
	}
}
