package query

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func prettyPrint(nodes []Node) string {
	var resultStr []string
	for _, node := range nodes {
		resultStr = append(resultStr, node.String())
	}
	return strings.Join(resultStr, " ")
}

func TestSubstituteAliases(t *testing.T) {
	input := "r:repo g:repogroup f:file"
	want := `(and "repo:repo" "repogroup:repogroup" "file:file")`
	query, _ := ParseAndOr(input)
	got := prettyPrint(SubstituteAliases(query))
	if diff := cmp.Diff(got, want); diff != "" {
		t.Fatal(diff)
	}
}

func TestLowercaseFieldNames(t *testing.T) {
	input := "rEpO:foo PATTERN"
	want := `(and "repo:foo" "PATTERN")`
	query, _ := ParseAndOr(input)
	got := prettyPrint(LowercaseFieldNames(query))
	if diff := cmp.Diff(got, want); diff != "" {
		t.Fatal(diff)
	}
}

func TestHoist(t *testing.T) {
	cases := []struct {
		input      string
		want       string
		wantErrMsg string
	}{
		{
			input: `repo:foo a or b`,
			want:  `"repo:foo" (or "a" "b")`,
		},
		{
			input: `repo:foo a or b file:bar`,
			want:  `"repo:foo" "file:bar" (or "a" "b")`,
		},
		{
			input: `repo:foo a or b or c file:bar`,
			want:  `"repo:foo" "file:bar" (or "a" "b" "c")`,
		},
		{
			input: "repo:foo bar { and baz {",
			want:  `"repo:foo" (and (concat "bar" "{") (concat "baz" "{"))`,
		},
		{
			input: "repo:foo bar { and baz { and qux {",
			want:  `"repo:foo" (and (concat "bar" "{") (concat "baz" "{") (concat "qux" "{"))`,
		},
		{
			input: `repo:foo a and b or c and d or e file:bar`,
			want:  `"repo:foo" "file:bar" (or (and "a" "b") (and "c" "d") "e")`,
		},
		// This next pattern is valid for the heuristic, even though the ordering of the
		// patterns 'a' and 'c' in the first and last position are not ordered next to the
		// 'or' keyword. This because no ordering is assumed for patterns vs. field:value
		// parameters in the grammar. To preserve relative ordering and check this would
		// impose significant complexity to PartitionParameters function during parsing, and
		// the PartitionSearchPattern helper function that the heurstic relies on. So: we
		// accept this heuristic behavior here.
		{
			input: `a repo:foo or b or file:bar c`,
			want:  `"repo:foo" "file:bar" (or "a" "b" "c")`,
		},
		// Errors.
		{
			input:      "repo:foo or a",
			wantErrMsg: "could not partition first or last expression",
		},
		{
			input:      "a or repo:foo",
			wantErrMsg: "could not partition first or last expression",
		},
		{
			input:      "repo:foo or repo:bar",
			wantErrMsg: "could not partition first or last expression",
		},
		{
			input:      "a b",
			wantErrMsg: "heuristic requires top-level and- or or-expression",
		},
		{
			input:      "repo:foo a or repo:foobar b or c file:bar",
			wantErrMsg: `inner expression (and "repo:foobar" "b") is not a pure pattern expression`,
		},
	}
	for _, c := range cases {
		t.Run("hoist", func(t *testing.T) {
			// To test Hoist, Use a simplified parse function that
			// does not perform the heuristic.
			parse := func(in string) []Node {
				parser := &parser{
					buf:        []byte(in),
					heuristics: parensAsPatterns,
				}
				nodes, _ := parser.parseOr()
				return newOperator(nodes, And)
			}
			query := parse(c.input)
			hoistedQuery, err := Hoist(query)
			if err != nil {
				if diff := cmp.Diff(c.wantErrMsg, err.Error()); diff != "" {
					t.Error(diff)
				}
				return
			}
			got := prettyPrint(hoistedQuery)
			if diff := cmp.Diff(c.want, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestSearchUppercase(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{
			input: `TeSt`,
			want:  `(and "TeSt" "case:yes")`,
		},
		{
			input: `test`,
			want:  `"test"`,
		},
		{
			input: `content:TeSt`,
			want:  `(and "TeSt" "case:yes")`,
		},
		{
			input: `content:test`,
			want:  `"test"`,
		},
		{
			input: `repo:foo TeSt`,
			want:  `(and "repo:foo" "TeSt" "case:yes")`,
		},
		{
			input: `repo:foo test`,
			want:  `(and "repo:foo" "test")`,
		},
		{
			input: `repo:foo content:TeSt`,
			want:  `(and "repo:foo" "TeSt" "case:yes")`,
		},
		{
			input: `repo:foo content:test`,
			want:  `(and "repo:foo" "test")`,
		},
		{
			input: `TeSt1 TesT2`,
			want:  `(and (concat "TeSt1" "TesT2") "case:yes")`,
		},
		{
			input: `TeSt1 test2`,
			want:  `(and (concat "TeSt1" "test2") "case:yes")`,
		},
	}
	for _, c := range cases {
		t.Run("searchUppercase", func(t *testing.T) {
			query, _ := ParseAndOr(c.input)
			got := prettyPrint(SearchUppercase(SubstituteAliases(query)))
			if diff := cmp.Diff(c.want, got); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestSubstituteOrForRegexp(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{
			input: "foo or bar",
			want:  `"(foo)|(bar)"`,
		},
		{
			input: "(foo or (bar or baz))",
			want:  `"(foo)|(bar)|(baz)"`,
		},
		{
			input: "repo:foobar foo or (bar or baz)",
			want:  `(or "(bar)|(baz)" (and "repo:foobar" "foo"))`,
		},
		{
			input: "(foo or (bar or baz)) and foobar",
			want:  `(and "(foo)|(bar)|(baz)" "foobar")`,
		},
		{
			input: "(foo or (bar and baz))",
			want:  `(or "(foo)" (and "bar" "baz"))`,
		},
		{
			input: "foo or (bar and baz) or foobar",
			want:  `(or "(foo)|(foobar)" (and "bar" "baz"))`,
		},
		{
			input: "repo:foo a or b",
			want:  `(and "repo:foo" "(a)|(b)")`,
		},
	}
	for _, c := range cases {
		t.Run("Map query", func(t *testing.T) {
			query, _ := ParseAndOr(c.input)
			got := prettyPrint(substituteOrForRegexp(query))
			if diff := cmp.Diff(c.want, got); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestSubstituteConcat(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{
			input: "a b c d e f",
			want:  `"a b c d e f"`,
		},
		{
			input: "a (b and c) d",
			want:  `"a" (and "b" "c") "d"`,
		},
		{
			input: "a b (c and d) e f (g or h) (i j k)",
			want:  `"a b" (and "c" "d") "e f" (or "g" "h") "(i j k)"`,
		},
		{
			input: "(((a b c))) and d",
			want:  `(and "(((a b c)))" "d")`,
		},
	}
	for _, c := range cases {
		t.Run("Map query", func(t *testing.T) {
			query, _ := ParseAndOr(c.input)
			got := prettyPrint(substituteConcat(query, " "))
			if diff := cmp.Diff(c.want, got); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestConvertEmptyGroupsToLiteral(t *testing.T) {
	cases := []struct {
		input      string
		wantLabels labels
	}{
		{
			input:      "func()",
			wantLabels: HeuristicParensAsPatterns | Literal,
		},
		{
			input:      "func(.*)",
			wantLabels: HeuristicParensAsPatterns | Regexp,
		},
	}
	for _, c := range cases {
		t.Run("Map query", func(t *testing.T) {
			query, _ := ParseAndOr(c.input)
			got := EmptyGroupsToLiteral(query)[0].(Pattern)
			if diff := cmp.Diff(c.wantLabels, got.Annotation.Labels); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestMap(t *testing.T) {
	cases := []struct {
		input string
		fns   []func(_ []Node) []Node
		want  string
	}{
		{
			input: "RePo:foo",
			fns:   []func(_ []Node) []Node{LowercaseFieldNames},
			want:  `"repo:foo"`,
		},
		{
			input: "RePo:foo r:bar",
			fns:   []func(_ []Node) []Node{LowercaseFieldNames, SubstituteAliases},
			want:  `(and "repo:foo" "repo:bar")`,
		},
	}
	for _, c := range cases {
		t.Run("Map query", func(t *testing.T) {
			query, _ := ParseAndOr(c.input)
			got := prettyPrint(Map(query, c.fns...))
			if diff := cmp.Diff(c.want, got); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestTranslateGlobToRegex(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{
			input: "*",
			want:  "[^/]*?",
		},
		{
			input: "*repo",
			want:  "[^/]*?repo$",
		},
		{
			input: "re*o",
			want:  "^re[^/]*?o$",
		},
		{
			input: "repo*",
			want:  "^repo[^/]*?",
		},
		{
			input: "?",
			want:  "^.$",
		},
		{
			input: "?repo",
			want:  "^.repo$",
		},
		{
			input: "re?o",
			want:  "^re.o$",
		},
		{
			input: "repo?",
			want:  "^repo.$",
		},
		{
			input: "123",
			want:  "^123$",
		},
		{
			input: ".123",
			want:  "^\\.123$",
		},
		{
			input: "*.go",
			want:  "[^/]*?\\.go$",
		},
		{
			input: "h[a-z]llo",
			want:  "^h[a-z]llo$",
		},
		{
			input: "h[!a-z]llo",
			want:  "^h[^a-z]llo$",
		},
		{
			input: "h[!abcde]llo",
			want:  "^h[^abcde]llo$",
		},
		{
			input: "h[]-]llo",
			want:  "^h[]-]llo$",
		},
		{
			input: "h\\[llo",
			want:  "^h\\[llo$",
		},
		{
			input: "h\\*llo",
			want:  "^h\\*llo$",
		},
		{
			input: "h\\?llo",
			want:  "^h\\?llo$",
		},
		{
			input: "fo[a-z]baz",
			want:  "^fo[a-z]baz$",
		},
		{
			input: "foo/**",
			want:  "^foo/.*?",
		},
		{
			input: "[a-z0-9]",
			want:  "^[a-z0-9]$",
		},
		{
			input: "[abc-]",
			want:  "^[abc-]$",
		},
		{
			input: "[--0]",
			want:  "^[--0]$",
		},
		{
			input: "",
			want:  "",
		},
	}

	for _, c := range cases {
		t.Run(c.want, func(t *testing.T) {
			got, err := globToRegex(c.input)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(c.want, got); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestTranslateBadGlobPattern(t *testing.T) {
	cases := []struct {
		input string
	}{
		{input: "fo[a-b-c]"},
		{input: "fo\\o"},
		{input: "fo[o"},
		{input: "[z-a]"},
		{input: "[a-z--0]"},
	}
	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			_, err := globToRegex(c.input)
			if diff := cmp.Diff(ErrBadGlobPattern.Error(), err.Error()); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
