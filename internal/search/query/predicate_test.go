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
			{`literal`, `repo`, &RepoDependenciesPredicate{RepoRev: "repo"}},
			{`regex with revs`, `^github\.com/org/repo$@v3.0:v4.0`, &RepoDependenciesPredicate{RepoRev: `^github\.com/org/repo$@v3.0:v4.0`}},
			{`regex with transitive:yes`, `^github\.com/org/repo$ transitive:yes`, &RepoDependenciesPredicate{RepoRev: `^github\.com/org/repo$`, Transitive: true}},
			{`transitive:no`, `repo transitive:no`, &RepoDependenciesPredicate{RepoRev: "repo", Transitive: false}},
			// transitive:only is ignored for now
			{`transitive:only`, `repo transitive:only`, &RepoDependenciesPredicate{RepoRev: "repo", Transitive: false}},
			{`transitive:horse`, `repo transitive:horse`, &RepoDependenciesPredicate{RepoRev: "repo", Transitive: false}},
		}

		for _, tc := range valid {
			t.Run(tc.name, func(t *testing.T) {
				p := &RepoDependenciesPredicate{}
				err := p.ParseParams(tc.params)
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}

				if !reflect.DeepEqual(tc.expected, p) {
					t.Fatalf("expected %+v, got %+v", tc.expected, p)
				}
			})
		}

		invalid := []test{
			{`empty`, ``, nil},
			{`catch invalid regexp`, `([)`, nil},
			{`only transitive`, `transitive:yes`, nil},
		}

		for _, tc := range invalid {
			t.Run(tc.name, func(t *testing.T) {
				p := &RepoDependenciesPredicate{}
				err := p.ParseParams(tc.params)
				if err == nil {
					t.Fatal("expected error but got none")
				}
			})
		}
	})
}

func TestRepoDependentsPredicate(t *testing.T) {
	t.Run("ParseParams", func(t *testing.T) {
		type test struct {
			name     string
			params   string
			expected *RepoDependentsPredicate
		}

		valid := []test{
			{`literal`, `test`, &RepoDependentsPredicate{}},
			{`regex with revs`, `^npm/@bar:baz`, &RepoDependentsPredicate{}},
			{`regex with single rev`, `^npm/foobar$@2.3.4`, &RepoDependentsPredicate{}},
		}

		for _, tc := range valid {
			t.Run(tc.name, func(t *testing.T) {
				p := &RepoDependentsPredicate{}
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
			{`catch invalid regexp`, `([)`, nil},
		}

		for _, tc := range invalid {
			t.Run(tc.name, func(t *testing.T) {
				p := &RepoDependentsPredicate{}
				err := p.ParseParams(tc.params)
				if err == nil {
					t.Fatal("expected error but got none")
				}
			})
		}
	})
}

func TestRepoHasDescriptionPredicate(t *testing.T) {
	t.Run("ParseParams", func(t *testing.T) {
		type test struct {
			name     string
			params   string
			expected *RepoHasDescriptionPredicate
		}

		valid := []test{
			{`literal`, `test`, &RepoHasDescriptionPredicate{Pattern: "test"}},
			{`regexp`, `test(.*)package`, &RepoHasDescriptionPredicate{Pattern: "test(.*)package"}},
		}

		for _, tc := range valid {
			t.Run(tc.name, func(t *testing.T) {
				p := &RepoHasDescriptionPredicate{}
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
			{`catch invalid regexp`, `([)`, nil},
		}

		for _, tc := range invalid {
			t.Run(tc.name, func(t *testing.T) {
				p := &RepoHasDescriptionPredicate{}
				err := p.ParseParams(tc.params)
				if err == nil {
					t.Fatal("expected error but got none")
				}
			})
		}
	})
}

func TestFileHasOwnerPredicate(t *testing.T) {
	t.Run("ParseParams", func(t *testing.T) {
		type test struct {
			name     string
			params   string
			expected *FileHasOwnerPredicate
		}

		valid := []test{
			{`literal`, `test`, &FileHasOwnerPredicate{Owner: "test"}},
			{`regexp`, `@octo-org/octocats`, &FileHasOwnerPredicate{Owner: "@octo-org/octocats"}},
			{`regexp`, `test@example.com`, &FileHasOwnerPredicate{Owner: "test@example.com"}},
		}

		for _, tc := range valid {
			t.Run(tc.name, func(t *testing.T) {
				p := &FileHasOwnerPredicate{}
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
		}

		for _, tc := range invalid {
			t.Run(tc.name, func(t *testing.T) {
				p := &FileHasOwnerPredicate{}
				err := p.ParseParams(tc.params)
				if err == nil {
					t.Fatal("expected error but got none")
				}
			})
		}
	})
}
