package search

import (
	"reflect"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
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

func TestRepoRevisionsQuery(t *testing.T) {
	repos := []*db.MinimalRepo{{Name: "foo"}, {Name: "bar"}, {Name: "baz"}}
	cases := map[string]string{
		// Short circuit (no ref specifier) which doesn't filter input
		"r:":         "foo@ bar@ baz@",
		"r:b":        "foo@ bar@ baz@",
		"r:b -r:baz": "foo@ bar@ baz@",
		"x":          "foo@ bar@ baz@",

		"x branch:dev": "foo@dev bar@dev baz@dev",

		// Zero or more branch specifiers, including default branch
		"x r:foo branch:dev":                              "foo@dev",
		"x r:foo (branch:dev or branch:insiders)":         "foo@dev:insiders",
		"x r:foo (branch:dev or branch:insiders or TRUE)": "foo@dev:insiders:",

		// Search insiders for foo and dev for b*
		"x ((r:foo branch:insiders) or (r:b branch:dev))": "foo@insiders bar@dev baz@dev",

		// Our older funky branch specifiers. Not exactly sure how this will
		// evolve yet.
		"x r:foo branch:*refs/heads/:^refs/heads/master": "foo@*refs/heads/:^refs/heads/master",

		// We don't know how to translate branch specifiers across nots, so we
		// disallow it.
		"x -branch:dev":      "error: search clauses that filter git refs cannot be negated",
		"x -(y branch:dev)":  "error: search clauses that filter git refs cannot be negated",
		"x -(y -branch:dev)": "error: search clauses that filter git refs cannot be negated",

		// Temp: Check we correctly handle regex
		"r:[f] branch:dev":       "foo@dev",
		"r:(foo|bar) branch:dev": "foo@dev bar@dev",
		"r:f$ branch:dev":        "",
		"r:f** branch:dev":       "error: invalid nested repetition operator",
	}

	for qStr, want := range cases {
		q, err := query.Parse(qStr)
		if err != nil {
			t.Fatal(err)
		}
		rr, err := RepoRevisionsQuery(q, repos)
		if err != nil {
			if strings.HasPrefix(want, "error: ") {
				got := err.Error()
				want = want[len("error: "):]
				if !strings.Contains(got, want) {
					t.Errorf("RepoRevisionsQuery(%q) error:\ngot  %s\nwant %s", qStr, got, want)
				}
				continue
			}
			t.Fatal(err)
		}
		var parts []string
		for _, r := range rr {
			parts = append(parts, r.String())
		}
		got := strings.Join(parts, " ")
		if got != want {
			t.Errorf("RepoRevisionsQuery(%q):\ngot  %s\nwant %s", qStr, got, want)
		}
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
