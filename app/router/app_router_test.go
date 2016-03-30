package router

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/kr/pretty"
	"github.com/sourcegraph/mux"
)

func TestMatch(t *testing.T) {
	tests := []struct {
		path          string
		wantNoMatch   bool
		wantRouteName string
		wantVars      map[string]string
		wantPath      string
	}{
		// Home
		{
			path:          "/",
			wantRouteName: Home,
		},

		// Repo
		{
			path:          "/repohost.com/foo",
			wantRouteName: Repo,
			wantVars:      map[string]string{"Repo": "repohost.com/foo", "Rev": ""},
		},
		{
			path:          "/a/b/c/d",
			wantRouteName: Repo,
			wantVars:      map[string]string{"Repo": "a/b/c/d", "Rev": ""},
		},
		{
			path:          "/a/b@mycommitid",
			wantRouteName: Repo,
			wantVars:      map[string]string{"Repo": "a/b", "Rev": "@mycommitid"},
		},
		{
			path:          "/a/b@myrev/subrev",
			wantRouteName: Repo,
			wantVars:      map[string]string{"Repo": "a/b", "Rev": "@myrev/subrev"},
		},
		{
			path:          "/a/b@myrev/subrev1/subrev2",
			wantRouteName: Repo,
			wantVars:      map[string]string{"Repo": "a/b", "Rev": "@myrev/subrev1/subrev2"},
		},

		// Repo sub-routes
		{
			path:          "/repohost.com/foo@mybranch/-/commits",
			wantRouteName: RepoRevCommits,
			wantVars:      map[string]string{"Repo": "repohost.com/foo", "Rev": "@mybranch"},
		},
		{
			path:          "/repohost.com/foo@mycommit/-/commit",
			wantRouteName: RepoCommit,
			wantVars:      map[string]string{"Repo": "repohost.com/foo", "Rev": "@mycommit"},
		},
		{
			path:          "/repohost.com/foo@branch/with/slash/-/commit",
			wantRouteName: RepoCommit,
			wantVars:      map[string]string{"Repo": "repohost.com/foo", "Rev": "@branch/with/slash"},
		},
		{
			path:          "/repohost.com/foo/-/tags",
			wantRouteName: RepoTags,
			wantVars:      map[string]string{"Repo": "repohost.com/foo"},
		},

		// Repo app sub-routes
		{
			path:          "/repohost.com/foo/-/app/myapp",
			wantRouteName: RepoAppFrame,
			wantVars:      map[string]string{"Repo": "repohost.com/foo", "Rev": "", "App": "myapp", "AppPath": ""},
		},
		{
			path:          "/repohost.com/foo/-/app/myapp/foo/.bar/baz",
			wantRouteName: RepoAppFrame,
			wantVars:      map[string]string{"Repo": "repohost.com/foo", "Rev": "", "App": "myapp", "AppPath": "/foo/.bar/baz"},
		},

		// Repo tree
		{
			path:          "/repohost.com/foo@mycommitid/-/tree",
			wantRouteName: RepoTree,
			wantVars:      map[string]string{"Repo": "repohost.com/foo", "Rev": "@mycommitid", "Path": ""},
		},
		{
			path:          "/repohost.com/foo@my-commit.id_2/-/tree",
			wantRouteName: RepoTree,
			wantVars:      map[string]string{"Repo": "repohost.com/foo", "Rev": "@my-commit.id_2", "Path": ""},
		},
		{
			path:          "/repohost.com/foo@mycommitid/-/tree",
			wantRouteName: RepoTree,
			wantVars:      map[string]string{"Repo": "repohost.com/foo", "Rev": "@mycommitid", "Path": ""},
			wantPath:      "/repohost.com/foo@mycommitid/-/tree",
		},
		{
			path:          "/repohost.com/foo@mycommitid/-/tree/my/file",
			wantRouteName: RepoTree,
			wantVars:      map[string]string{"Repo": "repohost.com/foo", "Rev": "@mycommitid", "Path": "/my/file"},
		},

		// Defs
		{
			path:          "/repohost.com/foo@mycommitid/-/def/t/u/p",
			wantRouteName: Def,
			wantVars:      map[string]string{"Repo": "repohost.com/foo", "UnitType": "t", "Unit": "u", "Path": "p", "Rev": "@mycommitid"},
		},
		{
			path:          "/repohost.com/foo@myrev/subrev/-/def/t/u/p",
			wantRouteName: Def,
			wantVars:      map[string]string{"Repo": "repohost.com/foo", "UnitType": "t", "Unit": "u", "Path": "p", "Rev": "@myrev/subrev"},
		},

		// Def sub-routes
		{
			path:          "/repohost.com/foo@mycommitid/-/def/t/u/p/refs",
			wantRouteName: DefRefs,
			wantVars:      map[string]string{"Repo": "repohost.com/foo", "Rev": "@mycommitid", "UnitType": "t", "Unit": "u", "Path": "p"},
		},
	}
	for _, test := range tests {
		var routeMatch mux.RouteMatch
		match := Rel.Match(&http.Request{Method: "GET", URL: &url.URL{Path: test.path}}, &routeMatch)

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
			t.Errorf("got generated path %q, want %q", path.Path, wantPath)
		}
	}
}
