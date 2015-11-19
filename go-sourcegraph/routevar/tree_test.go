package routevar

import (
	"net/http"
	"net/url"
	"path"
	"reflect"
	"testing"

	"github.com/sourcegraph/mux"
)

func TestTreeEntry(t *testing.T) {
	r := mux.NewRouter()
	r.Path("/x" + TreeEntryPath).PostMatchFunc(FixTreeEntryVars).BuildVarsFunc(PrepareTreeEntryRouteVars)

	tests := []struct {
		path        string
		wantNoMatch bool
		wantVars    map[string]string
	}{
		{path: "/x", wantVars: map[string]string{"Path": "."}},
		{path: "/x/", wantVars: map[string]string{"Path": "."}},
		{path: "/x/.", wantVars: map[string]string{"Path": "."}},
		{path: "/x/foo", wantVars: map[string]string{"Path": "foo"}},
		{path: "/x/foo/bar", wantVars: map[string]string{"Path": "foo/bar"}},
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
