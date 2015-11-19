package routevar

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/sourcegraph/mux"
)

const commitID = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func TestRepo(t *testing.T) {
	r := mux.NewRouter()
	r.Path("/" + Repo)

	tests := []struct {
		path        string
		wantNoMatch bool
		wantVars    map[string]string
	}{
		{path: "/foo", wantVars: map[string]string{"Repo": "foo"}},
		{path: "/foo.com/bar", wantVars: map[string]string{"Repo": "foo.com/bar"}},

		{path: "/", wantNoMatch: true},
		{path: "/.foo", wantNoMatch: true},
		{path: "/foo/.bar", wantNoMatch: true},
	}
	for _, test := range tests {
		var m mux.RouteMatch
		ok := r.Match(&http.Request{Method: "GET", URL: &url.URL{Path: test.path}}, &m)
		if ok == test.wantNoMatch {
			t.Errorf("%q: got match == %v, want %v", test.path, ok, !test.wantNoMatch)
		}
		if ok {
			if !reflect.DeepEqual(m.Vars, test.wantVars) {
				t.Errorf("%q: got vars == %v, want %v", test.path, m.Vars, test.wantVars)
			}

			url, err := m.Route.URLPath(pairs(m.Vars)...)
			if err != nil {
				t.Errorf("%q: URLPath: %s", test.path, err)
				continue
			}
			if url.Path != test.path {
				t.Errorf("%q: got path == %q, want %q", test.path, url.Path, test.path)
			}
		}
	}
}

func TestRepoRev(t *testing.T) {
	r := mux.NewRouter()
	r.Path("/" + RepoRev).PostMatchFunc(FixRepoRevVars).BuildVarsFunc(PrepareRepoRevRouteVars)

	tests := []struct {
		path        string
		wantNoMatch bool
		wantVars    map[string]string
	}{
		{path: "/foo", wantVars: map[string]string{"Repo": "foo"}},
		{path: "/foo@v", wantVars: map[string]string{"Repo": "foo", "Rev": "v"}},
		{path: "/foo@v===" + commitID, wantVars: map[string]string{"Repo": "foo", "Rev": "v", "CommitID": commitID}},
		{path: "/foo.com/bar", wantVars: map[string]string{"Repo": "foo.com/bar"}},
		{path: "/foo.com/bar@v", wantVars: map[string]string{"Repo": "foo.com/bar", "Rev": "v"}},
		{path: "/foo.com/bar@v===" + commitID, wantVars: map[string]string{"Repo": "foo.com/bar", "Rev": "v", "CommitID": commitID}},

		{path: "/", wantNoMatch: true},
		{path: "/.foo", wantNoMatch: true},
		{path: "/foo/.bar", wantNoMatch: true},
	}
	for _, test := range tests {
		var m mux.RouteMatch
		ok := r.Match(&http.Request{Method: "GET", URL: &url.URL{Path: test.path}}, &m)
		if ok == test.wantNoMatch {
			t.Errorf("%q: got match == %v, want %v", test.path, ok, !test.wantNoMatch)
		}
		if ok {
			if !reflect.DeepEqual(m.Vars, test.wantVars) {
				t.Errorf("%q: got vars == %v, want %v", test.path, m.Vars, test.wantVars)
			}

			url, err := m.Route.URLPath(pairs(m.Vars)...)
			if err != nil {
				t.Errorf("%q: URLPath: %s", test.path, err)
				continue
			}
			if url.Path != test.path {
				t.Errorf("%q: got path == %q, want %q", test.path, url.Path, test.path)
			}
		}
	}
}
