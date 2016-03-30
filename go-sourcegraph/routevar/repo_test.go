package routevar

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/gorilla/mux"
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

		{path: "/foo.com/bar/-/abc/def", wantNoMatch: true},
		{path: "/foo.com/bar@a", wantNoMatch: true},
		{path: "/foo.com/bar@a/b", wantNoMatch: true},
		{path: "/foo.com/bar/@a", wantNoMatch: true},
		{path: "/-/foo.com/bar", wantNoMatch: true},
		{path: "/", wantNoMatch: true},
		{path: "/-/", wantNoMatch: true},
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

func TestRev(t *testing.T) {
	r := mux.NewRouter()
	r.Path("/" + Rev)

	tests := []struct {
		path        string
		wantNoMatch bool
		wantVars    map[string]string
	}{
		{path: "/v", wantVars: map[string]string{"Rev": "v"}},
		{path: "/v/v/v", wantVars: map[string]string{"Rev": "v/v/v"}},
		{path: "/v===" + commitID, wantVars: map[string]string{"Rev": "v===" + commitID}},

		{path: "", wantNoMatch: true},
		{path: "/", wantNoMatch: true},
		{path: "/===", wantNoMatch: true},
		{path: "/v===", wantNoMatch: true},
		{path: "/===" + commitID, wantNoMatch: true},
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
