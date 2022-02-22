package query

import (
	"reflect"
	"testing"
)

func TestRepoContainsPredicate(t *testing.T) {
	t.Run("ParseParams", func(t *testing.T) {
		type test struct {
			name     string
			params   string
			expected *RepoContainsPredicate
		}

		valid := []test{
			{`file`, `file:test`, &RepoContainsPredicate{File: "test"}},
			{`file regex`, `file:test(a|b)*.go`, &RepoContainsPredicate{File: "test(a|b)*.go"}},
			{`content`, `content:test`, &RepoContainsPredicate{Content: "test"}},
			{`file and content`, `file:test.go content:abc`, &RepoContainsPredicate{File: "test.go", Content: "abc"}},
			{`content and file`, `content:abc file:test.go`, &RepoContainsPredicate{File: "test.go", Content: "abc"}},
		}

		for _, tc := range valid {
			t.Run(tc.name, func(t *testing.T) {
				p := &RepoContainsPredicate{}
				err := p.ParseParams(tc.params)
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}

				if !reflect.DeepEqual(tc.expected, p) {
					t.Fatalf("expected %#v, got %#v", tc.expected, p)
				}
			})
		}

		invalid := []test{
			{`empty`, ``, nil},
			{`negated file`, `-file:test`, nil},
			{`negated content`, `-content:test`, nil},
			{`unsupported syntax`, `abc:test`, nil},
			{`unnamed content`, `test`, nil},
			{`catch invalid content regexp`, `file:foo content:([)`, nil},
		}

		for _, tc := range invalid {
			t.Run(tc.name, func(t *testing.T) {
				p := &RepoContainsPredicate{}
				err := p.ParseParams(tc.params)
				if err == nil {
					t.Fatal("expected error but got none")
				}
			})
		}
	})
}

func TestParseAsPredicate(t *testing.T) {
	tests := []struct {
		input  string
		name   string
		params string
	}{
		{`a()`, "a", ""},
		{`a(b)`, "a", "b"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			name, params := ParseAsPredicate(tc.input)
			if name != tc.name {
				t.Fatalf("expected name %s, got %s", tc.name, name)
			}

			if params != tc.params {
				t.Fatalf("expected params %s, got %s", tc.params, params)
			}
		})
	}

}

func TestRepoDependenciesPredicate(t *testing.T) {
	t.Run("ParseParams", func(t *testing.T) {
		type test struct {
			name     string
			params   string
			expected *RepoDependenciesPredicate
		}

		valid := []test{
			{`of`, `of:test`, &RepoDependenciesPredicate{Of: "test"}},
			{`of with revs`, `of:^npm/@bar:baz`, &RepoDependenciesPredicate{Of: "^npm/@bar:baz"}},
			{`of with quotes`, `of:"^npm/"`, &RepoDependenciesPredicate{Of: "^npm/"}},
			{`case insensitive`, `Of:"^npm/"`, &RepoDependenciesPredicate{Of: "^npm/"}},
			{`of with spaces`, `Of:npm/ `, &RepoDependenciesPredicate{Of: "npm/"}},
		}

		for _, tc := range valid {
			t.Run(tc.name, func(t *testing.T) {
				p := &RepoDependenciesPredicate{}
				err := p.ParseParams(tc.params)
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}

				if !reflect.DeepEqual(tc.expected, p) {
					t.Fatalf("expected %#v, got %#v", tc.expected, p)
				}
			})
		}

		invalid := []test{
			{`empty`, ``, nil},
			{`negated of`, `-of:test`, nil},
			{`unsupported syntax`, `abc:test`, nil},
			{`unnamed param`, `test`, nil},
			{`unbalanced quotes`, `of:"test`, nil},
			{`catch invalid regexp`, `of:([)`, nil},
			{`duplicate of`, `of:foo of:bar`, nil},
			{`empty of`, `of:`, nil},
			{`empty of with quotes`, `of:""`, nil},
			// TODO: Reuse existing parser facilities which would allow
			// spaces inside a quote delimited value.
			{`of with spaces inside quotes`, `Of:"npm/ "`, nil},
		}

		for _, tc := range invalid {
			t.Run(tc.name, func(t *testing.T) {
				p := &RepoContainsPredicate{}
				err := p.ParseParams(tc.params)
				if err == nil {
					t.Fatal("expected error but got none")
				}
			})
		}
	})
}
