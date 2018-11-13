package search

import (
	"reflect"
	"testing"

	dbquery "github.com/sourcegraph/sourcegraph/cmd/frontend/db/query"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/search/query"
)

func TestParseRepositoryRevisions(t *testing.T) {
	tests := map[string]struct {
		repo api.RepoName
		revs []RevisionSpecifier
	}{
		"repo":           {repo: "repo", revs: []RevisionSpecifier{}},
		"repo@":          {repo: "repo", revs: []RevisionSpecifier{{RevSpec: ""}}},
		"repo@rev":       {repo: "repo", revs: []RevisionSpecifier{{RevSpec: "rev"}}},
		"repo@rev1:rev2": {repo: "repo", revs: []RevisionSpecifier{{RevSpec: "rev1"}, {RevSpec: "rev2"}}},
		"repo@:rev1:":    {repo: "repo", revs: []RevisionSpecifier{{RevSpec: "rev1"}}},
		"repo@*glob":     {repo: "repo", revs: []RevisionSpecifier{{RefGlob: "glob"}}},
		"repo@rev1:*glob1:^rev2": {
			repo: "repo",
			revs: []RevisionSpecifier{{RevSpec: "rev1"}, {RefGlob: "glob1"}, {RevSpec: "^rev2"}},
		},
		"repo@rev1:*glob1:*!glob2:rev2:*glob3": {
			repo: "repo",
			revs: []RevisionSpecifier{
				{RevSpec: "rev1"},
				{RefGlob: "glob1"},
				{ExcludeRefGlob: "glob2"},
				{RevSpec: "rev2"},
				{RefGlob: "glob3"},
			},
		},
	}
	for input, want := range tests {
		t.Run(input, func(t *testing.T) {
			repo, revs := ParseRepositoryRevisions(input)
			if repo != want.repo {
				t.Fatalf("got %+v, want %+v", repo, want.repo)
			}
			if !reflect.DeepEqual(revs, want.revs) {
				t.Fatalf("got %+v, want %+v", revs, want.revs)
			}
		})
	}
}

func TestRepoQuery(t *testing.T) {
	cases := map[string]string{
		"r:":         `""`,
		"r:b":        `"b"`,
		"r:b r:a":    `("b" AND "a")`,
		"r:b -r:baz": `("b" AND NOT("baz"))`,
		"-r:f":       `NOT("f")`,

		"foo -(r:foo r:baz)":                  `NOT(("foo" AND "baz"))`,
		"foo r:ba -(r:foo or r:baz)":          `("ba" AND NOT(("foo" OR "baz")))`,
		"foo r:ba -(r:foo or r:baz or hello)": `("ba" AND NOT(("foo" OR "baz")))`,
		"foo r:ba -(r:foo or hello)":          `("ba" AND NOT("foo"))`,

		// because of the hello, we could match baz. So we actually want to
		// look at all repos.
		"foo -(r:baz hello)": `TRUE`,

		// The world makes the subquery irrelevant
		"foo r:ba -((r:foo or hello) world)": `"ba"`,
		"foo r:ba -((r:foo or hello) r:foo)": `("ba" AND NOT(("foo" AND "foo")))`,

		"foo -(-(r:bar hello))": `"bar"`,
		"foo -(-(hello))":       `TRUE`,
		"foo -(r:bar hello)":    `TRUE`,
		"foo -(hello)":          `TRUE`,

		"foo -(type:repo -(r:bar hello))": `"bar"`,

		"foo r:b -(-(r:bar hello))": `("b" AND "bar")`,
		"foo r:b -(r:bar hello)":    `"b"`,
		"foo r:b -(-(hello))":       `"b"`,
		"foo r:b -(hello)":          `"b"`,
		"foo r:b":                   `"b"`,

		"r:foo test":                 `"foo"`,
		"r:foo test -hello":          `"foo"`,
		"(r:ba test (r:b r:a -r:z))": `("ba" AND "b" AND "a" AND NOT("z"))`,

		"bar -(r:foo test)": `TRUE`,
	}
	for qStr, want := range cases {
		q, err := query.Parse(qStr)
		if err != nil {
			t.Fatal(err)
		}
		rq, err := RepoQuery(q)
		if err != nil {
			t.Fatal(err)
		}
		got := dbquery.Print(rq)
		if got != want {
			t.Errorf("RepoQuery(%q):\ngot  %s\nwant %s", qStr, got, want)
		}
	}
}
