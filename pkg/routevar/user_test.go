package routevar

import (
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

func TestUserPattern(t *testing.T) {
	pat, err := regexp.Compile("^" + UserPattern + "$")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		input     string
		wantMatch bool
		wantError string
		wantUID   uint32
		wantLogin string
	}{
		{"alice", true, "", 0, "alice"},
		{"alice-x", true, "", 0, "alice-x"},
		{"alice.x", true, "", 0, "alice.x"},
		{"alice_x", true, "", 0, "alice_x"},
		{"123$", true, "", 123, ""},

		{input: "", wantMatch: false},
		{input: ".", wantMatch: false},
		{input: "~", wantMatch: false},
		{input: "$1", wantMatch: false},
		{input: "~@", wantMatch: false},
		{input: "1$@", wantMatch: false},
		{input: "999999999999999999999$", wantMatch: true, wantError: "value out of range"},
		{input: "alice@foo.com", wantMatch: false},
		{input: "alice@", wantMatch: false},
		{input: "alice@~", wantMatch: false},
		{input: "alice@.", wantMatch: false},
		{input: "alice@.com", wantMatch: false},
		{input: "alice@com.", wantMatch: false},
	}
	for _, test := range tests {
		match := pat.MatchString(test.input)
		if match != test.wantMatch {
			t.Errorf("%q: got match == %v, want %v", test.input, match, test.wantMatch)
		}

		uid, login, err := parseUser(test.input)
		if test.wantError != "" {
			if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Errorf("%q: got err == %v, want error to contain %q", test.input, err, test.wantError)
			}
			continue
		}
		if gotErr, wantErr := err != nil, !test.wantMatch; gotErr != wantErr {
			t.Errorf("%q: got err == %v, want error? == %v", test.input, err, wantErr)
		}
		if err == nil {
			if uid != test.wantUID {
				t.Errorf("%q: got uid == %d, want %d", test.input, uid, test.wantUID)
			}
			if login != test.wantLogin {
				t.Errorf("%q: got login == %q, want %q", test.input, login, test.wantLogin)
			}

			str := userString(uid, login)
			if str != test.input {
				t.Errorf("%q: got string %q, want %q", test.input, str, test.input)
			}
		}
	}
}

func TestUser(t *testing.T) {
	r := mux.NewRouter()
	r.Path("/" + User)

	tests := []struct {
		path        string
		wantNoMatch bool
		wantVars    map[string]string
	}{
		{path: "/foo", wantVars: map[string]string{"User": "foo"}},
		{path: "/1$", wantVars: map[string]string{"User": "1$"}},

		{path: "/foo@bar.com", wantNoMatch: true},
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

func TestUserSpec(t *testing.T) {
	tests := []struct {
		str       string
		spec      sourcegraph.UserSpec
		wantError bool
	}{
		{"a", sourcegraph.UserSpec{Login: "a"}, false},
		{"1$", sourcegraph.UserSpec{UID: 1}, false},
	}

	for _, test := range tests {
		spec, err := ParseUserSpec(test.str)
		if err != nil && !test.wantError {
			t.Errorf("%q: ParseUserSpec failed: %s", test.str, err)
		}
		if test.wantError && err == nil {
			t.Errorf("%q: ParseUserSpec returned nil error, want non-nil error", test.str)
			continue
		}
		if err != nil {
			continue
		}
		if spec != test.spec {
			t.Errorf("%q: got spec %+v, want %+v", test.str, spec, test.spec)
			continue
		}

		str := UserString(test.spec)
		if str != test.str {
			t.Errorf("%+v: got str %q, want %q", test.spec, str, test.str)
			continue
		}
	}
}
