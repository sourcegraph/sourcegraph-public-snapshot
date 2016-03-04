package routevar

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/gorilla/mux"
)

func TestUser(t *testing.T) {
	r := mux.NewRouter()
	r.Path("/" + User)

	tests := []struct {
		path        string
		wantNoMatch bool
		wantVars    map[string]string
	}{
		{path: "/foo", wantVars: map[string]string{"User": "foo"}},
		{path: "/foo@bar", wantVars: map[string]string{"User": "foo@bar"}},
		{path: "/foo@bar.com", wantVars: map[string]string{"User": "foo@bar.com"}},

		{path: "/", wantNoMatch: true},
		{path: "/.foo", wantNoMatch: true},
		{path: "/foo@", wantNoMatch: true},
		{path: "/foo/bar", wantNoMatch: true},
		{path: "/foo/bar@", wantNoMatch: true},
		{path: "/foo@bar/baz", wantNoMatch: true},
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
