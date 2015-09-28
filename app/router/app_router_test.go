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
			wantVars:      map[string]string{"Repo": "repohost.com/foo"},
		},
		{
			path:          "/a/b/c/d",
			wantRouteName: Repo,
			wantVars:      map[string]string{"Repo": "a/b/c/d"},
		},
		{
			path:          "/a/b@mycommitid",
			wantRouteName: Repo,
			wantVars:      map[string]string{"Repo": "a/b", "Rev": "mycommitid"},
		},
		{
			path:          "/a/b@myrev/subrev",
			wantRouteName: Repo,
			wantVars:      map[string]string{"Repo": "a/b", "Rev": "myrev/subrev"},
		},
		{
			path:          "/a/b@myrev/subrev1/subrev2",
			wantRouteName: Repo,
			wantVars:      map[string]string{"Repo": "a/b", "Rev": "myrev/subrev1/subrev2"},
		},

		// Repo sub-routes
		{
			path:          "/repohost.com/foo/.search",
			wantRouteName: RepoSearch,
			wantVars:      map[string]string{"Repo": "repohost.com/foo"},
		},
		{
			path:          "/repohost.com/foo@myrev/.search",
			wantRouteName: RepoSearch,
			wantVars:      map[string]string{"Repo": "repohost.com/foo", "Rev": "myrev"},
		},
		{
			path:          "/repohost.com/foo@a/b/.compare/c/d",
			wantRouteName: RepoCompare,
			wantVars:      map[string]string{"Repo": "repohost.com/foo", "Rev": "a/b", "Head": "c/d"},
		},
		{
			path:          "/repohost.com/foo@a/.compare/b",
			wantRouteName: RepoCompare,
			wantVars:      map[string]string{"Repo": "repohost.com/foo", "Rev": "a", "Head": "b"},
		},
		{
			path:          "/repohost.com/foo@a/.compare/b/.all",
			wantRouteName: RepoCompareAll,
			wantVars:      map[string]string{"Repo": "repohost.com/foo", "Rev": "a", "Head": "b"},
		},
		{
			path:          "/repohost.com/foo/.commits",
			wantRouteName: RepoRevCommits,
			wantVars:      map[string]string{"Repo": "repohost.com/foo"},
		},
		{
			path:          "/repohost.com/foo/.commits/123abc",
			wantRouteName: RepoCommit,
			wantVars:      map[string]string{"Repo": "repohost.com/foo", "Rev": "123abc"},
		},
		{
			path:          "/repohost.com/foo/.commits/branch/with/slash",
			wantRouteName: RepoCommit,
			wantVars:      map[string]string{"Repo": "repohost.com/foo", "Rev": "branch/with/slash"},
		},

		// Repo app sub-routes
		{
			path:          "/repohost.com/foo@myrev/.myapp",
			wantRouteName: RepoAppFrame,
			wantVars:      map[string]string{"Repo": "repohost.com/foo", "Rev": "myrev", "App": "myapp", "AppPath": ""},
		},
		{
			path:          "/repohost.com/foo@myrev/.myapp/foo/.bar/baz",
			wantRouteName: RepoAppFrame,
			wantVars:      map[string]string{"Repo": "repohost.com/foo", "Rev": "myrev", "App": "myapp", "AppPath": "/foo/.bar/baz"},
		},

		// Repo sub-routes that don't allow an "@REVSPEC" revision.
		{
			path:          "/repohost.com/foo/.tags",
			wantRouteName: RepoTags,
			wantVars:      map[string]string{"Repo": "repohost.com/foo"},
		},

		// Repo tree
		{
			path:          "/repohost.com/foo@mycommitid/.tree",
			wantRouteName: RepoTree,
			wantVars:      map[string]string{"Repo": "repohost.com/foo", "Rev": "mycommitid", "Path": "."},
		},
		{
			path:          "/repohost.com/foo@my-commit.id_2/.tree",
			wantRouteName: RepoTree,
			wantVars:      map[string]string{"Repo": "repohost.com/foo", "Rev": "my-commit.id_2", "Path": "."},
		},
		{
			path:          "/repohost.com/foo@mycommitid/.tree/",
			wantRouteName: RepoTree,
			wantVars:      map[string]string{"Repo": "repohost.com/foo", "Rev": "mycommitid", "Path": "."},
			wantPath:      "/repohost.com/foo@mycommitid/.tree",
		},
		{
			path:          "/repohost.com/foo@mycommitid/.tree/my/file",
			wantRouteName: RepoTree,
			wantVars:      map[string]string{"Repo": "repohost.com/foo", "Rev": "mycommitid", "Path": "my/file"},
		},

		// Defs
		{
			path:          "/repohost.com/foo@mycommitid/.t/.def/p",
			wantRouteName: Def,
			wantVars:      map[string]string{"Repo": "repohost.com/foo", "UnitType": "t", "Unit": ".", "Path": "p", "Rev": "mycommitid"},
		},
		{
			path:          "/repohost.com/foo@myrev/subrev/.t/.def/p",
			wantRouteName: Def,
			wantVars:      map[string]string{"Repo": "repohost.com/foo", "UnitType": "t", "Unit": ".", "Path": "p", "Rev": "myrev/subrev"},
		},
		{
			path:          "/repohost.com/foo/.t/.def/p",
			wantRouteName: Def,
			wantVars:      map[string]string{"Repo": "repohost.com/foo", "UnitType": "t", "Unit": ".", "Path": "p"},
		},
		{
			path:          "/repohost.com/foo/.t/.def/p%3F", // Ruby-like def that ends in '?', like `directory?`
			wantRouteName: Def,
			wantVars:      map[string]string{"Repo": "repohost.com/foo", "UnitType": "t", "Unit": ".", "Path": "p?"},
		},
		{
			path:          "/repohost.com/foo/.t/.def", // empty path
			wantRouteName: Def,
			wantVars:      map[string]string{"Repo": "repohost.com/foo", "UnitType": "t", "Unit": ".", "Path": "."},
		},
		{
			path:          "/repohost.com/foo/.t/u1/.def/p",
			wantRouteName: Def,
			wantVars:      map[string]string{"Repo": "repohost.com/foo", "UnitType": "t", "Unit": "u1", "Path": "p"},
		},
		{
			path:          "/repohost.com/foo/.t/u1/u2/.def/p1/p2",
			wantRouteName: Def,
			wantVars:      map[string]string{"Repo": "repohost.com/foo", "UnitType": "t", "Unit": "u1/u2", "Path": "p1/p2"},
		},

		// Def sub-routes
		{
			path:          "/repohost.com/foo/.t/.def/p/.examples",
			wantRouteName: DefExamples,
			wantVars:      map[string]string{"Repo": "repohost.com/foo", "UnitType": "t", "Unit": ".", "Path": "p"},
		},
		{
			path:          "/repohost.com/foo/.t/.def/.examples", // empty path
			wantRouteName: DefExamples,
			wantVars:      map[string]string{"Repo": "repohost.com/foo", "UnitType": "t", "Unit": ".", "Path": "."},
		},
		{
			path:          "/repohost.com/foo/.t/u1/u2/.def/p1/p2/.examples",
			wantRouteName: DefExamples,
			wantVars:      map[string]string{"Repo": "repohost.com/foo", "UnitType": "t", "Unit": "u1/u2", "Path": "p1/p2"},
		},

		// Sourceboxes
		{
			path:          "/repohost.com/foo@mycommitid/.t/.def/p/.sourcebox.js",
			wantRouteName: SourceboxDef,
			wantVars:      map[string]string{"Repo": "repohost.com/foo", "UnitType": "t", "Unit": ".", "Path": "p", "Rev": "mycommitid", "Format": "js"},
		},
		{
			path:          "/repohost.com/foo@mycommitid/.t/.def/p/.sourcebox.json",
			wantRouteName: SourceboxDef,
			wantVars:      map[string]string{"Repo": "repohost.com/foo", "UnitType": "t", "Unit": ".", "Path": "p", "Rev": "mycommitid", "Format": "json"},
		},
		{
			path:          "/repohost.com/foo@mycommitid/.tree/my/file/.sourcebox.js",
			wantRouteName: SourceboxFile,
			wantVars:      map[string]string{"Repo": "repohost.com/foo", "Rev": "mycommitid", "Path": "my/file", "Format": "js"},
		},

		{
			path:          "/repohost.com/foo@mycommitid/.tree/my/file/.sourcebox.json",
			wantRouteName: SourceboxFile,
			wantVars:      map[string]string{"Repo": "repohost.com/foo", "Rev": "mycommitid", "Path": "my/file", "Format": "json"},
		},

		// User
		{
			path:          "/~alice",
			wantRouteName: User,
			wantVars:      map[string]string{"User": "alice"},
		},
		{
			path:          "/~alice@example.com",
			wantRouteName: User,
			wantVars:      map[string]string{"User": "alice@example.com"},
		},
		{
			path:          "/~123$",
			wantRouteName: User,
			wantVars:      map[string]string{"User": "123$"},
		},
		{
			path:          "/~123$@example.com",
			wantRouteName: User,
			wantVars:      map[string]string{"User": "123$@example.com"},
		},
		{
			path:          "/~alice@-x-yJAANTud-iAVVw",
			wantRouteName: User,
			wantVars:      map[string]string{"User": "alice@-x-yJAANTud-iAVVw"},
		},

		// Blog
		{
			path:          "/blog",
			wantRouteName: BlogIndex,
		},
		{
			path:          "/blog.atom",
			wantRouteName: BlogIndexAtom,
			wantVars:      map[string]string{"Format": ".atom"},
		},
		{
			path:          "/blog/foo-bar",
			wantRouteName: BlogPost,
			wantVars:      map[string]string{"Slug": "foo-bar"},
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
