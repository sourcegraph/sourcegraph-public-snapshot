package ui

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/kr/pretty"
)

func TestRouter(t *testing.T) {
	tests := []struct {
		path          string
		wantNoMatch   bool
		wantRouteName string
		wantVars      map[string]string
		wantPath      string
	}{
		// Repo
		{
			path:          "/r",
			wantRouteName: routeRepo,
			wantVars:      map[string]string{"Repo": "r", "Rev": ""},
		},
		{
			path:          "/r/r",
			wantRouteName: routeRepo,
			wantVars:      map[string]string{"Repo": "r/r", "Rev": ""},
		},
		{
			path:          "/r/r@v",
			wantRouteName: routeRepo,
			wantVars:      map[string]string{"Repo": "r/r", "Rev": "@v"},
		},
		{
			path:          "/r/r@v/v",
			wantRouteName: routeRepo,
			wantVars:      map[string]string{"Repo": "r/r", "Rev": "@v/v"},
		},
		{
			path:        "/r/-/invalidroute",
			wantNoMatch: true,
		},
		{
			path:        "/r/-",
			wantNoMatch: true,
		},

		// Repo sub-routes
		{
			path:          "/r/-/builds",
			wantRouteName: routeRepoBuilds,
			wantVars:      map[string]string{"Repo": "r"},
		},
		{
			path:          "/r/-/builds/123",
			wantRouteName: routeBuild,
			wantVars:      map[string]string{"Repo": "r", "Build": "123"},
		},
		{
			path:        "/r@v/-/builds",
			wantNoMatch: true,
		},
		{
			path:        "/r@v/-/builds/123",
			wantNoMatch: true,
		},

		// Tree
		{
			path:          "/r@v/-/tree",
			wantRouteName: routeTree,
			wantVars:      map[string]string{"Repo": "r", "Rev": "@v", "Path": ""},
		},
		{
			path:          "/r@v/-/tree/d",
			wantRouteName: routeTree,
			wantVars:      map[string]string{"Repo": "r", "Rev": "@v", "Path": "/d"},
		},
		{
			path:          "/r@v/-/tree/d/d",
			wantRouteName: routeTree,
			wantVars:      map[string]string{"Repo": "r", "Rev": "@v", "Path": "/d/d"},
		},

		// Blob
		{
			path:          "/r@v/-/blob/f",
			wantRouteName: routeBlob,
			wantVars:      map[string]string{"Repo": "r", "Rev": "@v", "Path": "/f"},
		},
		{
			path:          "/r@v/-/blob/d/f",
			wantRouteName: routeBlob,
			wantVars:      map[string]string{"Repo": "r", "Rev": "@v", "Path": "/d/f"},
		},

		// Def
		{
			path:          "/r@v/-/def/t/u/-/p",
			wantRouteName: routeDef,
			wantVars:      map[string]string{"Repo": "r", "UnitType": "t", "Unit": "u", "Path": "p", "Rev": "@v"},
		},
		{
			path:          "/r@v/-/info/t/u/-/p",
			wantRouteName: routeDefLanding,
			wantVars:      map[string]string{"Repo": "r", "UnitType": "t", "Unit": "u", "Path": "p", "Rev": "@v"},
		},
	}
	for _, test := range tests {
		var routeMatch mux.RouteMatch
		match := router.Match(&http.Request{Method: "GET", URL: &url.URL{Path: test.path}}, &routeMatch)

		// Treat fallback/catch-all matches as non-matches.
		if match && (routeMatch.Route == nil || routeMatch.Route.GetName() == "") {
			match = false
		}

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
