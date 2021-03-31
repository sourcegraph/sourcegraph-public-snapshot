package predicate

import (
	"reflect"
	"testing"
)

func TestRepoContainsPredicate(t *testing.T) {
	t.Run("ParseParams", func(t *testing.T) {
		type test struct {
			name     string
			params   string
			expected *RepoContains
		}

		valid := []test{
			{`file`, `file:test`, &RepoContains{File: "test"}},
			{`file regex`, `file:test(a|b)*.go`, &RepoContains{File: "test(a|b)*.go"}},
			{`content`, `content:test`, &RepoContains{Content: "test"}},
			{`unnamed content`, `test`, &RepoContains{Content: "test"}},

			// TODO (@camdencheek) Query parsing currently checks parameter names against an allowlist.
			// This will be a problem as soon as we add more fields. Might make sense to do
			// as part of #19075
			{`unrecognized field`, `abc:test`, &RepoContains{Content: "abc:test"}},
		}

		for _, tc := range valid {
			t.Run(tc.name, func(t *testing.T) {
				p := &RepoContains{}
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
				p := &RepoContains{}
				err := p.ParseParams(tc.params)
				if err == nil {
					t.Fatal("expected error but got none")
				}
			})
		}
	})
}
