package router

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/kr/pretty"
	"github.com/sourcegraph/mux"
)

const commitID = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

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
			path:        "/repos/a.com/b@mycommitid", // doesn't accept a commit ID
			wantNoMatch: true,
		},
		{
			path:        "/repos/.invalidrepo",
			wantNoMatch: true,
		},

		// Repo sub-routes
		{
			path:          "/repos/a.com/b/.builds/123",
			wantRouteName: Build,
			wantVars:      map[string]string{"Repo": "a.com/b", "Build": "123"},
		},

		// Repo sub-routes that don't allow an "@REVSPEC" revision.
		{
			path:          "/repos/repohost.com/foo/.tags",
			wantRouteName: RepoTags,
			wantVars:      map[string]string{"Repo": "repohost.com/foo"},
		},
		{
			path:        "/repos/repohost.com/foo@myrevspec/.tags", // no @REVSPEC match
			wantNoMatch: true,
		},

		// Defs
		{
			path:          "/repos/repohost.com/foo@mycommitid/.defs/.t/.def/p",
			wantRouteName: Def,
			wantVars:      map[string]string{"Repo": "repohost.com/foo", "UnitType": "t", "Unit": ".", "Path": "p", "Rev": "mycommitid"},
		},
		{
			path:          "/repos/repohost.com/foo@myrev/mysubrev/.defs/.t/.def/p",
			wantRouteName: Def,
			wantVars:      map[string]string{"Repo": "repohost.com/foo", "UnitType": "t", "Unit": ".", "Path": "p", "Rev": "myrev/mysubrev"},
		},
		{
			path:          "/repos/repohost.com/foo/.defs/.t/.def/p",
			wantRouteName: Def,
			wantVars:      map[string]string{"Repo": "repohost.com/foo", "UnitType": "t", "Unit": ".", "Path": "p"},
		},
		{
			path:          "/repos/repohost.com/foo/.defs/.t/.def", // empty path
			wantRouteName: Def,
			wantVars:      map[string]string{"Repo": "repohost.com/foo", "UnitType": "t", "Unit": ".", "Path": "."},
		},
		{
			path:          "/repos/repohost.com/foo/.defs/.t/u1/.def/p",
			wantRouteName: Def,
			wantVars:      map[string]string{"Repo": "repohost.com/foo", "UnitType": "t", "Unit": "u1", "Path": "p"},
		},
		{
			path:          "/repos/repohost.com/foo/.defs/.t/u1/u2/.def/p1/p2",
			wantRouteName: Def,
			wantVars:      map[string]string{"Repo": "repohost.com/foo", "UnitType": "t", "Unit": "u1/u2", "Path": "p1/p2"},
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
