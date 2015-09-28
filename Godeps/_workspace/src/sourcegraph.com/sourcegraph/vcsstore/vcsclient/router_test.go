package vcsclient

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/kr/pretty"
	muxpkg "github.com/sourcegraph/mux"
)

func TestMatch(t *testing.T) {
	router := NewRouter(nil)

	const (
		repoPath        = "example.com/my/repo"
		encodedRepoPath = "example.com/my/repo"
	)

	tests := []struct {
		path          string
		wantNoMatch   bool
		wantRouteName string
		wantVars      map[string]string
		wantPath      string
	}{
		// Root
		{
			path:          "/",
			wantRouteName: RouteRoot,
		},

		// Repo
		{
			path:          "/" + encodedRepoPath,
			wantRouteName: RouteRepo,
			wantVars:      map[string]string{"RepoPath": repoPath},
		},
		{
			path:          "/myrepo",
			wantRouteName: RouteRepo,
			wantVars:      map[string]string{"RepoPath": "myrepo"},
		},

		// Repo revisions
		{
			path:          "/" + encodedRepoPath + "/.branches/mybranch",
			wantRouteName: RouteRepoBranch,
			wantVars:      map[string]string{"RepoPath": repoPath, "Branch": "mybranch"},
		},
		{
			path:          "/" + encodedRepoPath + "/.branches/mybranch/subbranch",
			wantRouteName: RouteRepoBranch,
			wantVars:      map[string]string{"RepoPath": repoPath, "Branch": "mybranch/subbranch"},
		},
		{
			path:          "/" + encodedRepoPath + "/.tags/mytag",
			wantRouteName: RouteRepoTag,
			wantVars:      map[string]string{"RepoPath": repoPath, "Tag": "mytag"},
		},
		{
			path:          "/" + encodedRepoPath + "/.tags/mytag/subtag",
			wantRouteName: RouteRepoTag,
			wantVars:      map[string]string{"RepoPath": repoPath, "Tag": "mytag/subtag"},
		},
		{
			path:          "/" + encodedRepoPath + "/.revs/myrevspec",
			wantRouteName: RouteRepoRevision,
			wantVars:      map[string]string{"RepoPath": repoPath, "RevSpec": "myrevspec"},
		},
		{
			path:          "/" + encodedRepoPath + "/.revs/myrevspec1/mysubdir2",
			wantRouteName: RouteRepoRevision,
			wantVars:      map[string]string{"RepoPath": repoPath, "RevSpec": "myrevspec1/mysubdir2"},
		},
		{
			path:          "/" + encodedRepoPath + "/.commits/mycommitid",
			wantRouteName: RouteRepoCommit,
			wantVars:      map[string]string{"RepoPath": repoPath, "CommitID": "mycommitid"},
		},

		// Repo commit log
		{
			path:          "/" + encodedRepoPath + "/.commits",
			wantRouteName: RouteRepoCommits,
			wantVars:      map[string]string{"RepoPath": repoPath},
		},

		// Repo tree
		{
			path:          "/" + encodedRepoPath + "/.commits/mycommitid/tree",
			wantRouteName: RouteRepoTreeEntry,
			wantVars:      map[string]string{"RepoPath": repoPath, "CommitID": "mycommitid", "Path": "."},
		},
		{
			path:          "/" + encodedRepoPath + "/.commits/mycommitid/tree/",
			wantRouteName: RouteRepoTreeEntry,
			wantVars:      map[string]string{"RepoPath": repoPath, "CommitID": "mycommitid", "Path": "."},
			wantPath:      "/" + encodedRepoPath + "/.commits/mycommitid/tree",
		},
		{
			path:          "/" + encodedRepoPath + "/.commits/mycommitid/tree/a/b",
			wantRouteName: RouteRepoTreeEntry,
			wantVars:      map[string]string{"RepoPath": repoPath, "CommitID": "mycommitid", "Path": "a/b"},
		},
		{
			path:          "/" + encodedRepoPath + "/.commits/mycommitid/tree/a/b/",
			wantRouteName: RouteRepoTreeEntry,
			wantVars:      map[string]string{"RepoPath": repoPath, "CommitID": "mycommitid", "Path": "a/b"},
			wantPath:      "/" + encodedRepoPath + "/.commits/mycommitid/tree/a/b",
		},

		// Diff
		{
			path:          "/" + encodedRepoPath + "/.diff/a..b",
			wantRouteName: RouteRepoDiff,
			wantVars:      map[string]string{"RepoPath": repoPath, "Base": "a", "Head": "b"},
		},

		// Cross-repo diff
		{
			path:          "/" + encodedRepoPath + "/.cross-repo-diff/a..x.com/y/z:b",
			wantRouteName: RouteRepoCrossRepoDiff,
			wantVars:      map[string]string{"RepoPath": repoPath, "Base": "a", "HeadRepoPath": "x.com/y/z", "Head": "b"},
		},

		// Merge Base
		{
			path:          "/" + encodedRepoPath + "/.merge-base/a/b",
			wantRouteName: RouteRepoMergeBase,
			wantVars:      map[string]string{"RepoPath": repoPath, "CommitIDA": "a", "CommitIDB": "b"},
		},

		// Cross-repo merge base
		{
			path:          "/" + encodedRepoPath + "/.cross-repo-merge-base/a/x.com/y/z/b",
			wantRouteName: RouteRepoCrossRepoMergeBase,
			wantVars:      map[string]string{"RepoPath": repoPath, "CommitIDA": "a", "BRepoPath": "x.com/y/z", "CommitIDB": "b"},
		},
	}

	for _, test := range tests {
		var routeMatch muxpkg.RouteMatch
		match := (*muxpkg.Router)(router).Match(&http.Request{Method: "GET", URL: &url.URL{Path: test.path}}, &routeMatch)

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
