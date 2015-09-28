package routevar

import (
	"net/http"
	"net/url"
	"path"
	"reflect"
	"testing"

	"github.com/sourcegraph/mux"
)

func TestDef(t *testing.T) {
	r := mux.NewRouter()
	r.Path("/" + Def).PostMatchFunc(FixDefUnitVars).BuildVarsFunc(PrepareDefRouteVars)

	tests := []struct {
		path        string
		wantNoMatch bool
		wantVars    map[string]string
	}{
		{path: "/.t/.def/p", wantVars: map[string]string{"UnitType": "t", "Unit": ".", "Path": "p"}},
		{path: "/.t/.def", wantVars: map[string]string{"UnitType": "t", "Unit": ".", "Path": "."}},
		{path: "/.t/u/.def/p", wantVars: map[string]string{"UnitType": "t", "Unit": "u", "Path": "p"}},
		{path: "/.t/u/v/.def/p", wantVars: map[string]string{"UnitType": "t", "Unit": "u/v", "Path": "p"}},
		{path: "/.t/u/.def/p/q", wantVars: map[string]string{"UnitType": "t", "Unit": "u", "Path": "p/q"}},

		{path: "/", wantNoMatch: true},
		{path: "/foo", wantNoMatch: true},
		{path: "/.t/foo", wantNoMatch: true},
		{path: "/.t/foo", wantNoMatch: true},
		{path: "/.t/.def/p/.p2", wantNoMatch: true},
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
			test.path = path.Clean(test.path)
			if url.Path != test.path {
				t.Errorf("%q: got path == %q, want %q", test.path, url.Path, test.path)
			}
		}
	}
}
