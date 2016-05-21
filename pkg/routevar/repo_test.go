package routevar

import (
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"

	"github.com/gorilla/mux"
)

func TestRepoPattern(t *testing.T) {
	pat, err := regexp.Compile("^" + RepoPattern + "$")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		input     string
		wantMatch bool
	}{
		{"foo", true},
		{"foo/bar", true},
		{"foo.com/bar", true},
		{"foo.com/-bar", true},
		{"foo.com/-bar-", true},
		{"foo.com/bar-", true},
		{"foo.com/.bar", true},
		{"foo.com/bar.baz", true},
		{"fo_o.com/bar", true},
		{".foo", true},
		{"./foo", true},

		{"", false},
		{"/foo", false},
		{"foo/", false},
		{"/foo/", false},
		{"foo.com/-", false},
		{"foo.com/-/bar", false},
		{"-/bar", false},
		{"/-/bar", false},
		{"bar@a", false},
		{"bar@a/b", false},
	}
	for _, test := range tests {
		match := pat.MatchString(test.input)
		if match != test.wantMatch {
			t.Errorf("%q: got match == %v, want %v", test.input, match, test.wantMatch)
		}

		repo, err := ParseRepo(test.input)
		if gotErr, wantErr := err != nil, !test.wantMatch; gotErr != wantErr {
			t.Errorf("%q: got err == %v, want error? == %v", test.input, err, wantErr)
		}
		if err == nil {
			if repo != test.input {
				t.Errorf("%q: got repo == %q, want %q", test.input, repo, test.input)
			}

			str := RepoString(repo)
			if str != test.input {
				t.Errorf("%q: got string %q, want %q", test.input, str, test.input)
			}
		}
	}
}

const commitID = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func TestResolvedRevPattern(t *testing.T) {
	pat, err := regexp.Compile("^" + RevPattern + "$")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		input                 string
		wantMatch             bool
		wantRev, wantCommitID string
	}{
		{"v", true, "v", ""},
		{"v/v", true, "v/v", ""},
		{"my/branch/name", true, "my/branch/name", ""},
		{"xx===" + commitID, true, "xx", commitID},
		{"bar~10", true, "bar~10", ""},
		{"bar^10", true, "bar^10", ""},

		{input: "===" + commitID, wantMatch: false},
		{input: "xx===aa", wantMatch: false},
		{input: "xx===", wantMatch: false},
		{input: "-", wantMatch: false},
		{input: "v/-", wantMatch: false},
		{input: "v/-/v", wantMatch: false},
		{input: "-/v", wantMatch: false},
	}
	for _, test := range tests {
		match := pat.MatchString(test.input)
		if match != test.wantMatch {
			t.Errorf("%q: got match == %v, want %v", test.input, match, test.wantMatch)
		}

		if test.wantMatch {
			rev, commitID := ParseResolvedRev(test.input)
			if rev != test.wantRev {
				t.Errorf("%q: got rev == %q, want %q", test.input, rev, test.wantRev)
			}
			if commitID != test.wantCommitID {
				t.Errorf("%q: got commitID == %q, want %q", test.input, commitID, test.wantCommitID)
			}

			str := resolvedRevString(rev, commitID)
			if str != test.input {
				t.Errorf("%q: got string %q, want %q", test.input, str, test.input)
			}
		}
	}
}

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

func TestRepoSpec(t *testing.T) {
	tests := []struct {
		str  string
		spec sourcegraph.RepoSpec
	}{
		{"a.com/x", sourcegraph.RepoSpec{URI: "a.com/x"}},
		{"x", sourcegraph.RepoSpec{URI: "x"}},
	}

	for _, test := range tests {
		spec, err := ParseRepoSpec(test.str)
		if err != nil {
			t.Errorf("%q: ParseRepoSpec failed: %s", test.str, err)
			continue
		}
		if spec != test.spec {
			t.Errorf("%q: got spec %+v, want %+v", test.str, spec, test.spec)
			continue
		}

		str := RepoSpecString(test.spec)
		if str != test.str {
			t.Errorf("%+v: got str %q, want %q", test.spec, str, test.str)
			continue
		}

		spec2, err := ToRepoSpec(RepoRouteVars(test.spec))
		if err != nil {
			t.Errorf("%+v: ToRepoSpec: %s", test.spec, err)
			continue
		}
		if spec2 != test.spec {
			t.Errorf("%q: got spec %+v, want %+v", test.str, spec, test.spec)
			continue
		}
	}
}

func TestRepoRevSpec(t *testing.T) {
	tests := []struct {
		spec      sourcegraph.RepoRevSpec
		routeVars map[string]string
	}{
		{sourcegraph.RepoRevSpec{RepoSpec: sourcegraph.RepoSpec{URI: "a.com/x"}, Rev: "r"}, map[string]string{"Repo": "a.com/x", "Rev": "@r"}},
		{sourcegraph.RepoRevSpec{RepoSpec: sourcegraph.RepoSpec{URI: "x"}, Rev: "r"}, map[string]string{"Repo": "x", "Rev": "@r"}},
		{sourcegraph.RepoRevSpec{RepoSpec: sourcegraph.RepoSpec{URI: "a.com/x"}, Rev: "r", CommitID: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}, map[string]string{"Repo": "a.com/x", "Rev": "@r===aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}},
	}

	for _, test := range tests {
		routeVars := RepoRevRouteVars(test.spec)
		if !reflect.DeepEqual(routeVars, test.routeVars) {
			t.Errorf("got route vars %+v, want %+v", routeVars, test.routeVars)
		}
		spec, err := ToRepoRevSpec(routeVars)
		if err != nil {
			t.Errorf("ToRepoRevSpec(%+v): %s", routeVars, err)
			continue
		}
		if spec != test.spec {
			t.Errorf("got spec %+v from route vars %+v, want %+v", spec, routeVars, test.spec)
		}
	}
}
