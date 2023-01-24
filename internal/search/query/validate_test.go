package query

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestValidation(t *testing.T) {
	cases := []struct {
		input      string
		searchType SearchType // nil value is regexp
		want       string
	}{
		{
			input: "index:foo",
			want:  `invalid value "foo" for field "index". Valid values are: yes, only, no`,
		},
		{
			input: "case:yes case:no",
			want:  `field "case" may not be used more than once`,
		},
		{
			input: "repo:[",
			want:  "error parsing regexp: missing closing ]: `[`",
		},
		{
			input: "repo:[@rev]",
			want:  "error parsing regexp: missing closing ]: `[`",
		},
		{
			input: "repo:\\@Query\\(\"SELECT",
			want:  "error parsing regexp: trailing backslash at end of expression: ``",
		},
		{
			input: "file:filename[2.txt",
			want:  "error parsing regexp: missing closing ]: `[2.txt`",
		},
		{
			input: "-index:yes",
			want:  `field "index" does not support negation`,
		},
		{
			input: "lang:c lang:go lang:stephenhas9cats",
			want:  `unknown language: "stephenhas9cats"`,
		},
		{
			input: "count:sedonuts",
			want:  "field count has value sedonuts, sedonuts is not a number",
		},
		{
			input: "count:10000000000000000",
			want:  "field count has a value that is out of range, try making it smaller",
		},
		{
			input: "count:-1",
			want:  "field count requires a positive number",
		},
		{
			input: "+",
			want:  "error parsing regexp: missing argument to repetition operator: `+`",
		},
		{
			input: `\\\`,
			want:  "error parsing regexp: trailing backslash at end of expression: ``",
		},
		{
			input:      `-content:"foo"`,
			want:       "the query contains a negated search pattern. Structural search does not support negated search patterns at the moment",
			searchType: SearchTypeStructural,
		},
		{
			input:      `NOT foo`,
			want:       "the query contains a negated search pattern. Structural search does not support negated search patterns at the moment",
			searchType: SearchTypeStructural,
		},
		{
			input: "repo:foo rev:a rev:b",
			want:  `field "rev" may not be used more than once`,
		},
		{
			input: "repo:foo@a rev:b",
			want:  "invalid syntax. You specified both @ and rev: for a repo: filter and I don't know how to interpret this. Remove either @ or rev: and try again",
		},
		{
			input: "rev:this is a good channel",
			want:  "invalid syntax. The query contains `rev:` without `repo:`. Add a `repo:` filter and try again",
		},
		{
			input: `repo:'' rev:bedge`,
			want:  "invalid syntax. The query contains `rev:` without `repo:`. Add a `repo:` filter and try again",
		},
		{
			input: "repo:foo author:rob@saucegraph.com",
			want:  `your query contains the field 'author', which requires type:commit or type:diff in the query`,
		},
		{
			input: "repohasfile:README type:symbol yolo",
			want:  "repohasfile is not compatible for type:symbol. Subscribe to https://github.com/sourcegraph/sourcegraph/issues/4610 for updates",
		},
		{
			input: "foo context:a context:b",
			want:  `field "context" may not be used more than once`,
		},
		{
			input: "-context:a",
			want:  `field "context" does not support negation`,
		},
		{
			input: "type:symbol select:symbol.timelime",
			want:  `invalid field "timelime" on select path "symbol.timelime"`,
		},
		{
			input:      "nice try type:repo",
			want:       "this structural search query specifies `type:` and is not supported. Structural search syntax only applies to searching file contents",
			searchType: SearchTypeStructural,
		},
		{
			input:      "type:diff nice try",
			want:       "this structural search query specifies `type:` and is not supported. Structural search syntax only applies to searching file contents and is not currently supported for diff searches",
			searchType: SearchTypeStructural,
		},
	}
	for _, c := range cases {
		t.Run("validate and/or query", func(t *testing.T) {
			_, err := Pipeline(Init(c.input, c.searchType))
			if err == nil {
				t.Fatal(fmt.Sprintf("expected test for %s to fail", c.input))
			}
			if diff := cmp.Diff(c.want, err.Error()); diff != "" {
				t.Fatal(diff)
			}

		})

	}
}

func TestIsCaseSensitive(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "yes",
			input: "case:yes",
			want:  true,
		},
		{
			name:  "no (explicit)",
			input: "case:no",
			want:  false,
		},
		{
			name:  "no (default)",
			input: "case:no",
			want:  false,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			query, err := ParseRegexp(c.input)
			if err != nil {
				t.Fatal(err)
			}
			got := query.IsCaseSensitive()
			if got != c.want {
				t.Errorf("got %v, want %v", got, c.want)
			}
		})
	}
}

func TestPartitionSearchPattern(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{
			input: "x",
			want:  `"x"`,
		},
		{
			input: "file:foo",
			want:  `"file:foo"`,
		},
		{
			input: "x y",
			want:  `(concat "x" "y")`,
		},
		{
			input: "x or y",
			want:  `(or "x" "y")`,
		},
		{
			input: "x and y",
			want:  `(and "x" "y")`,
		},
		{
			input: "file:foo x y",
			want:  `"file:foo" (concat "x" "y")`,
		},
		{
			input: "file:foo (x y)",
			want:  `"file:foo" "(x y)"`,
		},
		{
			input: "(file:foo x) y",
			want:  "cannot evaluate: unable to partition pure search pattern",
		},
		{
			input: "file:foo (x and y)",
			want:  `"file:foo" (and "x" "y")`,
		},
		{
			input: "file:foo x and y",
			want:  `"file:foo" (and "x" "y")`,
		},
		{
			input: "file:foo (x or y)",
			want:  `"file:foo" (or "x" "y")`,
		},
		{
			input: "file:foo x or y",
			want:  `"file:foo" (or "x" "y")`,
		},
		{
			input: "(file:foo x) or y",
			want:  "cannot evaluate: unable to partition pure search pattern",
		},
		{
			input: "file:foo and content:x",
			want:  `"file:foo" "content:x"`,
		},
		{
			input: "repo:foo and file:bar and x",
			want:  `"repo:foo" "file:bar" "x"`,
		},
		{
			input: "repo:foo and (file:bar or file:baz) and x",
			want:  "cannot evaluate: unable to partition pure search pattern",
		},
	}
	for _, tt := range cases {
		t.Run("partition search pattern", func(t *testing.T) {
			q, _ := Parse(tt.input, SearchTypeRegex)
			scopeParameters, pattern, err := PartitionSearchPattern(q)
			if err != nil {
				if diff := cmp.Diff(tt.want, err.Error()); diff != "" {
					t.Fatal(diff)
				}
				return
			}
			result := toNodes(scopeParameters)
			if pattern != nil {
				result = append(result, pattern)
			}
			got := toString(result)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestForAll(t *testing.T) {
	nodes := []Node{
		Parameter{Field: "repo", Value: "foo"},
		Parameter{Field: "repo", Value: "bar"},
	}
	result := ForAll(nodes, func(node Node) bool {
		_, ok := node.(Parameter)
		return ok
	})
	if !result {
		t.Errorf("Expected all nodes to be parameters.")
	}
}

func TestContainsRefGlobs(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		{
			input: "repo:foo",
			want:  false,
		},
		{
			input: "repo:foo@bar",
			want:  false,
		},
		{
			input: "repo:foo@*ref/tags",
			want:  true,
		},
		{
			input: "repo:foo@*!refs/tags",
			want:  true,
		},
		{
			input: "repo:foo@bar:*refs/heads",
			want:  true,
		},
		{
			input: "repo:foo@refs/tags/v3.14.3",
			want:  false,
		},
		{
			input: "repo:foo@*refs/tags/v3.14.?",
			want:  true,
		},
		{
			input: "repo:foo@v3.14.3 repo:foo@*refs/tags/v3.14.* bar",
			want:  true,
		},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			query, err := Run(Sequence(
				Init(c.input, SearchTypeLiteral),
			))
			if err != nil {
				t.Error(err)
			}
			got := ContainsRefGlobs(query)
			if got != c.want {
				t.Errorf("got %t, expected %t", got, c.want)
			}
		})
	}
}
