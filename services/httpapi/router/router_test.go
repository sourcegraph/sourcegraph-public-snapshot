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
		// Repo
		{
			path:          "/repos/repohost.com/foo",
			wantRouteName: Repo,
			wantVars:      map[string]string{"Repo": "repohost.com/foo"},
		},
		{
			path:          "/repos/a/b/c",
			wantRouteName: Repo,
			wantVars:      map[string]string{"Repo": "a/b/c"},
		},
		{
			path:          "/repos/a.com/b",
			wantRouteName: Repo,
			wantVars:      map[string]string{"Repo": "a.com/b"},
		},
		{
			path:        "/repos/-/myrepo",
			wantNoMatch: true,
		},
		{
			path:        "/repos/myrepo/-/invalidroute",
			wantNoMatch: true,
		},
		{
			path:        "/repos/myrepo/-",
			wantNoMatch: true,
		},

		// Repo sub-routes
		{
			path:          "/repos/a.com/b/-/builds/123",
			wantRouteName: RepoBuild,
			wantVars:      map[string]string{"Repo": "a.com/b", "Build": "123"},
		},
		{
			path:          "/repos/repohost.com/foo/-/tags",
			wantRouteName: RepoTags,
			wantVars:      map[string]string{"Repo": "repohost.com/foo"},
		},
		{
			path:        "/repos/repohost.com/foo/-/tags/myrevspec",
			wantNoMatch: true,
		},

		// Defs
		{
			path:          "/repos/repohost.com/foo@mycommitid/-/def/t/u/-/p",
			wantRouteName: Def,
			wantVars:      map[string]string{"Repo": "repohost.com/foo", "UnitType": "t", "Unit": "u", "Path": "p", "Rev": "@mycommitid"},
		},
		{
			path:          "/repos/repohost.com/foo@myrev/subrev/-/def/t/u/-/p",
			wantRouteName: Def,
			wantVars:      map[string]string{"Repo": "repohost.com/foo", "UnitType": "t", "Unit": "u", "Path": "p", "Rev": "@myrev/subrev"},
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
