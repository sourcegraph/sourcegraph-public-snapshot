package query

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hexops/autogold"
	"github.com/pkg/errors"
)

func collectLabels(nodes []Node) (result labels) {
	for _, node := range nodes {
		switch v := node.(type) {
		case Operator:
			result |= v.Annotation.Labels
			result |= collectLabels(v.Operands)
		case Pattern:
			result |= v.Annotation.Labels
		}
	}
	return result
}

func heuristicLabels(nodes []Node) string {
	labels := collectLabels(nodes)
	return strings.Join(labels.String(), ",")
}

func TestParseParameterList(t *testing.T) {
	type value struct {
		Want       string
		WantLabels string
		WantRange  string
	}

	test := func(input string) value {
		parser := &parser{buf: []byte(input), heuristics: parensAsPatterns | allowDanglingParens}
		result, err := parser.parseLeaves(Regexp)
		if err != nil {
			t.Fatal(fmt.Sprintf("Unexpected error: %s", err))
		}
		resultNode := result[0]
		got, _ := json.Marshal(resultNode)

		var gotRange string
		switch n := resultNode.(type) {
		case Pattern:
			gotRange = n.Annotation.Range.String()
		case Parameter:
			gotRange = n.Annotation.Range.String()
		}

		var gotLabels string
		if _, ok := resultNode.(Pattern); ok {
			gotLabels = heuristicLabels([]Node{resultNode})
		}

		return value{
			Want:       string(got),
			WantLabels: gotLabels,
			WantRange:  gotRange,
		}
	}

	autogold.Want("Normal field:value", value{
		Want:      `{"field":"file","value":"README.md","negated":false}`,
		WantRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":14}}`,
	}).Equal(t, test(`file:README.md`))

	autogold.Want("Normal field:value with trailing space", value{
		Want:      `{"field":"file","value":"README.md","negated":false}`,
		WantRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":14}}`,
	}).Equal(t, test(`file:README.md    `))

	autogold.Want("First char is colon", value{
		Want: `{"value":":foo","negated":false}`, WantLabels: "Regexp",
		WantRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":4}}`,
	}).Equal(t, test(`:foo`))

	autogold.Want("Last char is colon", value{
		Want: `{"value":"foo:","negated":false}`, WantLabels: "Regexp",
		WantRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":4}}`,
	}).Equal(t, test(`foo:`))

	autogold.Want("Match first colon", value{
		Want:      `{"field":"file","value":"bar:baz","negated":false}`,
		WantRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":12}}`,
	}).Equal(t, test(`file:bar:baz`))

	autogold.Want("No field, start with minus", value{
		Want: `{"value":"-:foo","negated":false}`, WantLabels: "Regexp",
		WantRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":5}}`,
	}).Equal(t, test(`-:foo`))

	autogold.Want("Minus prefix on field", value{
		Want:      `{"field":"file","value":"README.md","negated":true}`,
		WantRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":15}}`,
	}).Equal(t, test(`-file:README.md`))

	autogold.Want("NOT prefix on file", value{
		Want:      `{"field":"file","value":"README.md","negated":true}`,
		WantRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":18}}`,
	}).Equal(t, test(`NOT file:README.md`))

	autogold.Want("NOT prefix on unsupported key-value pair", value{
		Want: `{"value":"foo:bar","negated":true}`, WantLabels: "Regexp",
		WantRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":11}}`,
	}).Equal(t, test(`NOT foo:bar`))
	autogold.Want("NOT prefix on content", value{
		Want:      `{"field":"content","value":"bar","negated":true}`,
		WantRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":15}}`,
	}).Equal(t, test(`NOT content:bar`))

	autogold.Want("Double NOT", value{
		Want: `{"value":"NOT","negated":true}`, WantLabels: "Regexp",
		WantRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":7}}`,
	}).Equal(t, test(`NOT NOT`))

	autogold.Want("Double minus prefix on field", value{
		Want:       `{"value":"--foo:bar","negated":false}`,
		WantLabels: "Regexp",
		WantRange:  `{"start":{"line":0,"column":0},"end":{"line":0,"column":9}}`,
	}).Equal(t, test(`--foo:bar`))

	autogold.Want("Minus in the middle is not a valid field", value{
		Want:       `{"value":"fie-ld:bar","negated":false}`,
		WantLabels: "Regexp",
		WantRange:  `{"start":{"line":0,"column":0},"end":{"line":0,"column":10}}`,
	}).Equal(t, test(`fie-ld:bar`))

	autogold.Want("Preserve escaped whitespace", value{
		Want:       `{"value":"a\\ pattern","negated":false}`,
		WantLabels: "Regexp",
		WantRange:  `{"start":{"line":0,"column":0},"end":{"line":0,"column":10}}`,
	}).Equal(t, test(`a\ pattern`))

	autogold.Want("Quoted", value{
		Want: `{"value":"quoted","negated":false}`, WantLabels: "Literal,Quoted",
		WantRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":8}}`,
	}).Equal(t, test(`"quoted"`))

	autogold.Want("Escaped quote", value{
		Want: `{"value":"'","negated":false}`, WantLabels: "Literal,Quoted",
		WantRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":4}}`,
	}).Equal(t, test(`'\''`))

	autogold.Want("Regexp syntax with unbalanced paren", value{
		Want:       `{"value":"foo.*bar(","negated":false}`,
		WantLabels: "HeuristicDanglingParens,Regexp",
		WantRange:  `{"start":{"line":0,"column":0},"end":{"line":0,"column":9}}`,
	}).Equal(t, test(`foo.*bar(`))

	autogold.Want("Regexp delimiters", value{
		Want:       `{"value":"a regex pattern","negated":false}`,
		WantLabels: "Regexp",
		WantRange:  `{"start":{"line":0,"column":0},"end":{"line":0,"column":17}}`,
	}).Equal(t, test(`/a regex pattern/`))

	autogold.Want("Regexp group", value{
		Want:       `{"value":"Search()\\(","negated":false}`,
		WantLabels: "Regexp",
		WantRange:  `{"start":{"line":0,"column":0},"end":{"line":0,"column":10}}`,
	}).Equal(t, test(`Search()\(`))

	autogold.Want("Regexp non-empty group", value{
		Want:       `{"value":"Search(xxx)\\\\(","negated":false}`,
		WantLabels: "HeuristicDanglingParens,Regexp",
		WantRange:  `{"start":{"line":0,"column":0},"end":{"line":0,"column":14}}`,
	}).Equal(t, test(`Search(xxx)\\(`))
}

func TestScanField(t *testing.T) {
	type value struct {
		Field   string
		Negated bool
		Advance int
	}

	test := func(input string) string {
		gotField, gotNegated, gotAdvance := ScanField([]byte(input))
		v, _ := json.Marshal(value{gotField, gotNegated, gotAdvance})
		return string(v)
	}

	autogold.Want("repo:foo", `{"Field":"repo","Negated":false,"Advance":5}`).Equal(t, test("repo:foo"))
	autogold.Want("RepO:foo", `{"Field":"RepO","Negated":false,"Advance":5}`).Equal(t, test("RepO:foo"))
	autogold.Want("after:", `{"Field":"after","Negated":false,"Advance":6}`).Equal(t, test("after:"))
	autogold.Want("-repo:", `{"Field":"repo","Negated":true,"Advance":6}`).Equal(t, test("-repo:"))
	autogold.Want("", `{"Field":"","Negated":false,"Advance":0}`).Equal(t, test(""))
	autogold.Want("-", `{"Field":"","Negated":false,"Advance":0}`).Equal(t, test("-"))
	autogold.Want("-:", `{"Field":"","Negated":false,"Advance":0}`).Equal(t, test("-:"))
	autogold.Want(":", `{"Field":"","Negated":false,"Advance":0}`).Equal(t, test(":"))
	autogold.Want("??:foo", `{"Field":"","Negated":false,"Advance":0}`).Equal(t, test("??:foo"))
	autogold.Want("repo", `{"Field":"","Negated":false,"Advance":0}`).Equal(t, test("repo"))
	autogold.Want("-repo", `{"Field":"","Negated":false,"Advance":0}`).Equal(t, test("-repo"))
	autogold.Want("--repo:", `{"Field":"","Negated":false,"Advance":0}`).Equal(t, test("--repo:"))
	autogold.Want(":foo", `{"Field":"","Negated":false,"Advance":0}`).Equal(t, test(":foo"))

}

func parseAndOrGrammar(in string) ([]Node, error) {
	if strings.TrimSpace(in) == "" {
		return nil, nil
	}
	parser := &parser{
		buf:        []byte(in),
		leafParser: SearchTypeRegex,
	}
	nodes, err := parser.parseOr()
	if err != nil {
		return nil, err
	}
	if parser.balanced != 0 {
		return nil, errors.New("unbalanced expression: unmatched closing parenthesis )")
	}
	return newOperator(nodes, And), nil
}

func TestParse(t *testing.T) {
	type relation string         // a relation for comparing test outputs of queries parsed according to grammar and heuristics.
	const Same relation = "Same" // a constant that says heuristic output is interpreted the same as the grammar spec.
	type Spec = relation         // constructor for expected output of the grammar spec without heuristics.
	type Diff = relation         // constructor for expected heuristic output when different to the grammar spec.

	cases := []struct {
		Name          string
		Input         string
		WantGrammar   relation
		WantHeuristic relation
	}{
		{
			Name:          "Empty string",
			Input:         "",
			WantGrammar:   "",
			WantHeuristic: Same,
		},
		{
			Name:          "Whitespace",
			Input:         "             ",
			WantGrammar:   "",
			WantHeuristic: Same,
		},
		{
			Name:          "Single",
			Input:         "a",
			WantGrammar:   `"a"`,
			WantHeuristic: Same,
		},
		{
			Name:          "Whitespace basic",
			Input:         "a b",
			WantGrammar:   `(concat "a" "b")`,
			WantHeuristic: Same,
		},
		{
			Name:          "Basic",
			Input:         "a and b and c",
			WantGrammar:   `(and "a" "b" "c")`,
			WantHeuristic: Same,
		},
		{
			Input:         "(f(x)oo((a|b))bar)",
			WantGrammar:   Spec(`(concat "f" "x" "oo" "a|b" "bar")`),
			WantHeuristic: Diff(`"(f(x)oo((a|b))bar)"`),
		},
		{
			Input:         "aorb",
			WantGrammar:   `"aorb"`,
			WantHeuristic: Same,
		},
		{
			Input:         "aANDb",
			WantGrammar:   `"aANDb"`,
			WantHeuristic: Same,
		},
		{
			Input:         "a oror b",
			WantGrammar:   `(concat "a" "oror" "b")`,
			WantHeuristic: Same,
		},
		{
			Name:          "Reduced complex query mixed caps",
			Input:         "a and b AND c or d and (e OR f) g h i or j",
			WantGrammar:   `(or (and "a" "b" "c") (and "d" (concat (or "e" "f") "g" "h" "i")) "j")`,
			WantHeuristic: Same,
		},
		{
			Name:          "Basic reduced complex query",
			Input:         "a and b or c and d or e",
			WantGrammar:   `(or (and "a" "b") (and "c" "d") "e")`,
			WantHeuristic: Same,
		},
		{
			Name:          "Reduced complex query, reduction over parens",
			Input:         "(a and b or c and d) or e",
			WantGrammar:   `(or (and "a" "b") (and "c" "d") "e")`,
			WantHeuristic: Same,
		},
		{
			Name:          "Reduced complex query, nested 'or' trickles up",
			Input:         "(a and b or c) or d",
			WantGrammar:   `(or (and "a" "b") "c" "d")`,
			WantHeuristic: Same,
		},
		{
			Name:          "Reduced complex query, nested nested 'or' trickles up",
			Input:         "(a and b or (c and d or f)) or e",
			WantGrammar:   `(or (and "a" "b") (and "c" "d") "f" "e")`,
			WantHeuristic: Same,
		},
		{
			Name:          "No reduction on precedence defined by parens",
			Input:         "(a and (b or c) and d) or e",
			WantGrammar:   `(or (and "a" (or "b" "c") "d") "e")`,
			WantHeuristic: Same,
		},
		{
			Name:          "Paren reduction over operators",
			Input:         "(((a b c))) and d",
			WantGrammar:   Spec(`(and (concat "a" "b" "c") "d")`),
			WantHeuristic: Diff(`(and "(((a b c)))" "d")`),
		},
		// Partition parameters and concatenated patterns.
		{
			Input:         "a (b and c) d",
			WantGrammar:   `(concat "a" (and "b" "c") "d")`,
			WantHeuristic: Same,
		},
		{
			Input:         "(a b c) and (d e f) and (g h i)",
			WantGrammar:   Spec(`(and (concat "a" "b" "c") (concat "d" "e" "f") (concat "g" "h" "i"))`),
			WantHeuristic: `(and "(a b c)" "(d e f)" "(g h i)")`,
		},
		{
			Input:         "(a) repo:foo (b)",
			WantGrammar:   Spec(`(and "repo:foo" (concat "a" "b"))`),
			WantHeuristic: Diff(`(and "repo:foo" (concat "(a)" "(b)"))`),
		},
		{
			Input:         "repo:foo func( or func(.*)",
			WantGrammar:   Spec(`expected operand at 15`),
			WantHeuristic: Diff(`(and "repo:foo" (or "func(" "func(.*)"))`),
		},
		{
			Input:         "repo:foo main { and bar {",
			WantGrammar:   Spec(`(and (and "repo:foo" (concat "main" "{")) (concat "bar" "{"))`),
			WantHeuristic: Diff(`(and "repo:foo" (concat "main" "{") (concat "bar" "{"))`),
		},
		{
			Input:         "a b (repo:foo c d)",
			WantGrammar:   `(concat "a" "b" (and "repo:foo" (concat "c" "d")))`,
			WantHeuristic: Same,
		},
		{
			Input:         "a b (c d repo:foo)",
			WantGrammar:   `(concat "a" "b" (and "repo:foo" (concat "c" "d")))`,
			WantHeuristic: Same,
		},
		{
			Input:         "a repo:b repo:c (d repo:e repo:f)",
			WantGrammar:   `(and "repo:b" "repo:c" (concat "a" (and "repo:e" "repo:f" "d")))`,
			WantHeuristic: Same,
		},
		{
			Input:         "a repo:b repo:c (repo:e repo:f (repo:g repo:h))",
			WantGrammar:   `(and "repo:b" "repo:c" "repo:e" "repo:f" "repo:g" "repo:h" "a")`,
			WantHeuristic: Same,
		},
		{
			Input:         "a repo:b repo:c (repo:e repo:f (repo:g repo:h)) b",
			WantGrammar:   `(and "repo:b" "repo:c" "repo:e" "repo:f" "repo:g" "repo:h" (concat "a" "b"))`,
			WantHeuristic: Same,
		},
		{
			Input:         "a repo:b repo:c (repo:e repo:f (repo:g repo:h b)) ",
			WantGrammar:   `(and "repo:b" "repo:c" (concat "a" (and "repo:e" "repo:f" "repo:g" "repo:h" "b")))`,
			WantHeuristic: Same,
		},
		{
			Input:         "(repo:foo a (repo:bar b (repo:qux c)))",
			WantGrammar:   `(and "repo:foo" (concat "a" (and "repo:bar" (concat "b" (and "repo:qux" "c")))))`,
			WantHeuristic: Same,
		},
		{
			Input:         "a repo:b repo:c (d repo:e repo:f e)",
			WantGrammar:   `(and "repo:b" "repo:c" (concat "a" (and "repo:e" "repo:f" (concat "d" "e"))))`,
			WantHeuristic: Same,
		},
		// Keywords as patterns.
		{
			Input:         "a or",
			WantGrammar:   `(concat "a" "or")`,
			WantHeuristic: Same,
		},
		{
			Input:         "or",
			WantGrammar:   `"or"`,
			WantHeuristic: Same,
		},
		{
			Input:         "or or or",
			WantGrammar:   `(or "or" "or")`,
			WantHeuristic: Same,
		},
		{
			Input:         "and and andand or oror",
			WantGrammar:   `(or (and "and" "andand") "oror")`,
			WantHeuristic: Same,
		},
		// Errors.
		{
			Name:          "Unbalanced",
			Input:         "(foo) (bar",
			WantGrammar:   Spec(`unbalanced expression: unmatched closing parenthesis )`),
			WantHeuristic: Diff(`(concat "(foo)" "(bar")`),
		},
		{
			Name:          "Illegal expression on the right",
			Input:         "a or or b",
			WantGrammar:   "expected operand at 5",
			WantHeuristic: Same,
		},
		{
			Name:          "Illegal expression on the right, mixed operators",
			Input:         "a and OR",
			WantGrammar:   `(and "a" "OR")`,
			WantHeuristic: Same,
		},
		{
			Input:         "repo:foo or or or",
			WantGrammar:   "expected operand at 12",
			WantHeuristic: Same,
		},
		// Reduction.
		{
			Name:          "paren reduction with ands",
			Input:         "(a and b) and (c and d)",
			WantGrammar:   `(and "a" "b" "c" "d")`,
			WantHeuristic: Same,
		},
		{
			Name:          "paren reduction with ors",
			Input:         "(a or b) or (c or d)",
			WantGrammar:   `(or "a" "b" "c" "d")`,
			WantHeuristic: Same,
		},
		{
			Name:          "nested paren reduction with whitespace",
			Input:         "(((a b c))) d",
			WantGrammar:   Spec(`(concat "a" "b" "c" "d")`),
			WantHeuristic: Diff(`(concat "(((a b c)))" "d")`),
		},
		{
			Name:          "left paren reduction with whitespace",
			Input:         "(a b) c d",
			WantGrammar:   Spec(`(concat "a" "b" "c" "d")`),
			WantHeuristic: Diff(`(concat "(a b)" "c" "d")`),
		},
		{
			Name:          "right paren reduction with whitespace",
			Input:         "a b (c d)",
			WantGrammar:   Spec(`(concat "a" "b" "c" "d")`),
			WantHeuristic: Diff(`(concat "a" "b" "(c d)")`),
		},
		{
			Name:          "grouped paren reduction with whitespace",
			Input:         "(a b) (c d)",
			WantGrammar:   Spec(`(concat "a" "b" "c" "d")`),
			WantHeuristic: Diff(`(concat "(a b)" "(c d)")`),
		},
		{
			Name:          "multiple grouped paren reduction with whitespace",
			Input:         "(a b) (c d) (e f)",
			WantGrammar:   Spec(`(concat "a" "b" "c" "d" "e" "f")`),
			WantHeuristic: Diff(`(concat "(a b)" "(c d)" "(e f)")`),
		},
		{
			Name:          "interpolated grouped paren reduction",
			Input:         "(a b) c d (e f)",
			WantGrammar:   Spec(`(concat "a" "b" "c" "d" "e" "f")`),
			WantHeuristic: Diff(`(concat "(a b)" "c" "d" "(e f)")`),
		},
		{
			Name:          "mixed interpolated grouped paren reduction",
			Input:         "(a and b and (z or q)) and (c and d) and (e and f)",
			WantGrammar:   `(and "a" "b" (or "z" "q") "c" "d" "e" "f")`,
			WantHeuristic: Same,
		},
		// Parentheses.
		{
			Name:          "empty paren",
			Input:         "()",
			WantGrammar:   Spec(`""`),
			WantHeuristic: Diff(`"()"`),
		},
		{
			Name:          "paren inside contiguous string",
			Input:         "foo()bar",
			WantGrammar:   Spec(`(concat "foo" "bar")`),
			WantHeuristic: Diff(`"foo()bar"`),
		},
		{
			Name:          "paren inside contiguous string",
			Input:         "(x and regex(s)?)",
			WantGrammar:   Spec(`(and "x" (concat "regex" "s" "?"))`),
			WantHeuristic: Diff(`(and "x" "regex(s)?")`),
		},
		{
			Name:          "paren containing whitespace inside contiguous string",
			Input:         "foo(   )bar",
			WantGrammar:   Spec(`(concat "foo" "bar")`),
			WantHeuristic: Diff(`"foo(   )bar"`),
		},
		{
			Name:          "nested empty paren",
			Input:         "(x())",
			WantGrammar:   Spec(`"x"`),
			WantHeuristic: Diff(`"(x())"`),
		},
		{
			Name:          "interpolated nested empty paren",
			Input:         "(()x(  )(())())",
			WantGrammar:   Spec(`"x"`),
			WantHeuristic: Diff(`"(()x(  )(())())"`),
		},
		{
			Name:          "empty paren on or",
			Input:         "() or ()",
			WantGrammar:   Spec(`""`),
			WantHeuristic: Diff(`(or "()" "()")`),
		},
		{
			Name:          "empty left paren on or",
			Input:         "() or (x)",
			WantGrammar:   Spec(`"x"`),
			WantHeuristic: Diff(`(or "()" "(x)")`),
		},
		{
			Name:          "complex interpolated nested empty paren",
			Input:         "(()x(  )(y or () or (f))())",
			WantGrammar:   Spec(`(concat "x" (or "y" "f"))`),
			WantHeuristic: Diff(`(concat "()" "x" "()" (or "y" "()" "(f)") "()")`),
		},
		{
			Name:          "disable parens as patterns heuristic if containing recognized operator",
			Input:         "(() or ())",
			WantGrammar:   Spec(`""`),
			WantHeuristic: Diff(`(or "()" "()")`),
		},
		{
			Name:          "NOT expression inside parentheses",
			Input:         "r:foo (a/foo not .svg)",
			WantGrammar:   `(and "r:foo" (concat "a/foo" (not ".svg")))`,
			WantHeuristic: Same,
		},
		{
			Name:          "NOT expression inside parentheses",
			Input:         "r:foo (not .svg)",
			WantGrammar:   `(and "r:foo" (not ".svg"))`,
			WantHeuristic: Same,
		},
		// Escaping.
		{
			Input:         `\(\)`,
			WantGrammar:   `"\\(\\)"`,
			WantHeuristic: Same,
		},
		{
			Input:         `\( \) ()`,
			WantGrammar:   Spec(`(concat "\\(" "\\)")`),
			WantHeuristic: Diff(`(concat "\\(" "\\)" "()")`),
		},
		{
			Input:         `\ `,
			WantGrammar:   `"\\ "`,
			WantHeuristic: Same,
		},
		{
			Input:         `\  \ `,
			WantGrammar:   Spec(`(concat "\\ " "\\ ")`),
			WantHeuristic: Diff(`(concat "\\ " "\\ ")`),
		},
		// Dangling parentheses heuristic.
		{
			Input:         `(`,
			WantGrammar:   Spec(`expected operand at 1`),
			WantHeuristic: Diff(`"("`),
		},
		{
			Input:         `)(())(`,
			WantGrammar:   Spec(`unbalanced expression: unmatched closing parenthesis )`),
			WantHeuristic: Same,
		},
		{
			Input:         `foo( and bar(`,
			WantGrammar:   Spec(`expected operand at 5`),
			WantHeuristic: Diff(`(and "foo(" "bar(")`),
		},
		{
			Input:         `repo:foo foo( or bar(`,
			WantGrammar:   Spec(`expected operand at 14`),
			WantHeuristic: Diff(`(and "repo:foo" (or "foo(" "bar("))`),
		},
		{
			Input:         `(a or (b and )) or d)`,
			WantGrammar:   Spec(`unbalanced expression: unmatched closing parenthesis )`),
			WantHeuristic: Same,
		},
		// Quotes and escape sequences.
		{
			Input:         `"`,
			WantGrammar:   `"\""`,
			WantHeuristic: Same,
		},
		{
			Input:         `repo:foo' bar'`,
			WantGrammar:   `(and "repo:foo'" "bar'")`,
			WantHeuristic: Same,
		},
		{
			Input:         `repo:'foo' 'bar'`,
			WantGrammar:   `(and "repo:foo" "bar")`,
			WantHeuristic: Same,
		},
		{
			Input:         `repo:"foo" "bar"`,
			WantGrammar:   `(and "repo:foo" "bar")`,
			WantHeuristic: Same,
		},
		{
			Input:         `repo:"foo bar" "foo bar"`,
			WantGrammar:   `(and "repo:foo bar" "foo bar")`,
			WantHeuristic: Same,
		},
		{
			Input:         `repo:"fo\"o" "bar"`,
			WantGrammar:   Spec(`(and "repo:fo\"o" "bar")`),
			WantHeuristic: Same,
		},
		{
			Input:         `repo:foo /b\/ar/`,
			WantGrammar:   `(and "repo:foo" "b/ar")`,
			WantHeuristic: Same,
		},
		{
			Input:         `repo:foo /a/file/path`,
			WantGrammar:   `(and "repo:foo" "/a/file/path")`,
			WantHeuristic: Same,
		},
		{
			Input:         `repo:foo /a/file/path/`,
			WantGrammar:   `(and "repo:foo" "/a/file/path/")`,
			WantHeuristic: Same,
		},
		{
			Input:         `repo:foo /a/ /another/path/`,
			WantGrammar:   `(and "repo:foo" (concat "a" "/another/path/"))`,
			WantHeuristic: Same,
		},
		{
			Input:         `repo:foo /\s+b\d+ar/ `,
			WantGrammar:   `(and "repo:foo" "\\s+b\\d+ar")`,
			WantHeuristic: Same,
		},
		{
			Input:         `repo:foo /bar/ `,
			WantGrammar:   `(and "repo:foo" "bar")`,
			WantHeuristic: Same,
		},
		{
			Input:         `\t\r\n`,
			WantGrammar:   `"\\t\\r\\n"`,
			WantHeuristic: Same,
		},
		{
			Input:         `repo:foo\ bar \:\\`,
			WantGrammar:   `(and "repo:foo\\ bar" "\\:\\\\")`,
			WantHeuristic: Same,
		},
		{
			Input:         `a file:\.(ts(?:(?:)|x)|js(?:(?:)|x))(?m:$)`,
			WantGrammar:   `(and "file:\\.(ts(?:(?:)|x)|js(?:(?:)|x))(?m:$)" "a")`,
			WantHeuristic: Same,
		},
		{
			Input:         `(file:(a) file:(b))`,
			WantGrammar:   `(and "file:(a)" "file:(b)")`,
			WantHeuristic: Same,
		},
		{
			Input:         `(repohascommitafter:"7 days")`,
			WantGrammar:   `"repohascommitafter:7 days"`,
			WantHeuristic: Same,
		},
		{
			Input:         `(foo repohascommitafter:"7 days")`,
			WantGrammar:   `(and "repohascommitafter:7 days" "foo")`,
			WantHeuristic: Same,
		},
		// Fringe tests cases at the boundary of heuristics and invalid syntax.
		{
			Input:         `(0(F)(:())(:())(<0)0()`,
			WantGrammar:   Spec(`unbalanced expression: unmatched closing parenthesis )`),
			WantHeuristic: `"(0(F)(:())(:())(<0)0()"`,
		},
		// The space-looking character below is U+00A0.
		{
			Input:         `00 (000)`,
			WantGrammar:   `(concat "00" "000")`,
			WantHeuristic: `(concat "00" "(000)")`,
		},
	}
	for _, tt := range cases {
		t.Run(tt.Name, func(t *testing.T) {
			check := func(result []Node, err error, want string) {
				var resultStr []string
				if err != nil {
					if diff := cmp.Diff(want, err.Error()); diff != "" {
						t.Fatal(diff)
					}
					return
				}
				for _, node := range result {
					resultStr = append(resultStr, node.String())
				}
				got := strings.Join(resultStr, " ")
				if diff := cmp.Diff(want, got); diff != "" {
					t.Error(diff)
				}
			}
			var result []Node
			var err error
			result, err = parseAndOrGrammar(tt.Input) // Parse without heuristic.
			check(result, err, string(tt.WantGrammar))
			result, err = ParseAndOr(tt.Input, SearchTypeRegex)
			if tt.WantHeuristic == Same {
				check(result, err, string(tt.WantGrammar))
			} else {
				check(result, err, string(tt.WantHeuristic))
			}
		})
	}
}

func TestScanDelimited(t *testing.T) {
	type result struct {
		Value  string
		Count  int
		ErrMsg string
	}

	test := func(input string, delimiter rune) string {
		value, count, err := ScanDelimited([]byte(input), true, delimiter)
		var errMsg string
		if err != nil {
			errMsg = err.Error()
		}
		v, _ := json.Marshal(result{value, count, errMsg})
		return string(v)
	}

	autogold.Want(`""`, `{"Value":"","Count":2,"ErrMsg":""}`).Equal(t, test(`""`, '"'))
	autogold.Want(`"a"`, `{"Value":"a","Count":3,"ErrMsg":""}`).Equal(t, test(`"a"`, '"'))
	autogold.Want(`"\""`, `{"Value":"\"","Count":4,"ErrMsg":""}`).Equal(t, test(`"\""`, '"'))
	autogold.Want(`"\\""`, `{"Value":"\\","Count":4,"ErrMsg":""}`).Equal(t, test(`"\\""`, '"'))
	autogold.Want(`"\\\"`, `{"Value":"","Count":5,"ErrMsg":"unterminated literal: expected \""}`).Equal(t, test(`"\\\"`, '"'))
	autogold.Want(`"\\\""`, `{"Value":"\\\"","Count":6,"ErrMsg":""}`).Equal(t, test(`"\\\""`, '"'))
	autogold.Want(`"a`, `{"Value":"","Count":2,"ErrMsg":"unterminated literal: expected \""}`).Equal(t, test(`"a`, '"'))
	autogold.Want(`"\?"`, `{"Value":"","Count":3,"ErrMsg":"unrecognized escape sequence"}`).Equal(t, test(`"\?"`, '"'))
	autogold.Want(`/\//`, `{"Value":"/","Count":4,"ErrMsg":""}`).Equal(t, test(`/\//`, '/'))

	// The next invocation of test needs to panic.
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for ScanDelimited")
		}
	}()
	_ = test(`a"`, '"')
}

func TestMergePatterns(t *testing.T) {
	test := func(input string) string {
		p := &parser{buf: []byte(input), heuristics: parensAsPatterns}
		nodes, err := p.parseLeaves(Regexp)
		got := nodes[0].(Pattern).Annotation.Range.String()
		if err != nil {
			t.Error(err)
		}
		return got
	}

	autogold.Want("foo()bar", `{"start":{"line":0,"column":0},"end":{"line":0,"column":8}}`).Equal(t, test("foo()bar"))
	autogold.Want("()bar", `{"start":{"line":0,"column":0},"end":{"line":0,"column":5}}`).Equal(t, test("()bar"))
}

func TestMatchUnaryKeyword(t *testing.T) {

	test := func(input string, pos int) string {
		p := &parser{buf: []byte(input), pos: pos}
		return fmt.Sprintf("%t", p.matchUnaryKeyword("NOT"))
	}

	autogold.Want("NOT bar", "true").Equal(t, test("NOT bar", 0))
	autogold.Want("foo NOT bar", "true").Equal(t, test("foo NOT bar", 4))
	autogold.Want("foo NOT", "false").Equal(t, test("foo NOT", 4))
	autogold.Want("fooNOT bar", "false").Equal(t, test("fooNOT bar", 3))
	autogold.Want("NOTbar", "false").Equal(t, test("NOTbar", 0))
	autogold.Want("(not bar)", "true").Equal(t, test("(not bar)", 1))
}

func TestParseAndOrLiteral(t *testing.T) {
	type value struct {
		Want       string
		WantLabels string `json:",omitempty"`
		WantError  string `json:",omitempty"`
	}

	test := func(input string) string {
		result, err := ParseAndOr(input, SearchTypeLiteral)
		if err != nil {
			return fmt.Sprintf("ERROR: %s", err.Error())
		}
		wantLabels := heuristicLabels(result)
		var resultStr []string
		for _, node := range result {
			resultStr = append(resultStr, node.String())
		}
		want := fmt.Sprintf("%s", strings.Join(resultStr, " "))
		if wantLabels != "" {
			return fmt.Sprintf("%s (%s)", want, wantLabels)
		}
		return want
	}

	autogold.Want("()", `"()" (HeuristicParensAsPatterns,Literal)`).Equal(t, test("()"))
	autogold.Want(`"`, `"\"" (Literal)`).Equal(t, test(`"`))
	autogold.Want(`""`, `"\"\"" (Literal)`).Equal(t, test(`""`))
	autogold.Want("(", `"(" (HeuristicDanglingParens,Literal)`).Equal(t, test("("))
	autogold.Want("repo:foo foo( or bar(", `(and "repo:foo" (or "foo(" "bar(")) (HeuristicHoisted,Literal)`).Equal(t, test("repo:foo foo( or bar("))
	autogold.Want("x or", `(concat "x" "or") (Literal)`).Equal(t, test("x or"))
	autogold.Want("repo:foo (x", `(and "repo:foo" "(x") (HeuristicDanglingParens,Literal)`).Equal(t, test("repo:foo (x"))
	autogold.Want("(x or bar() )", `(or "x" "bar()") (Literal)`).Equal(t, test("(x or bar() )"))
	autogold.Want("(x", `"(x" (HeuristicDanglingParens,Literal)`).Equal(t, test("(x"))
	autogold.Want("x or (x", `(or "x" "(x") (HeuristicDanglingParens,HeuristicHoisted,Literal)`).Equal(t, test("x or (x"))
	autogold.Want("(y or (z", `(or "(y" "(z") (HeuristicDanglingParens,HeuristicHoisted,Literal)`).Equal(t, test("(y or (z"))
	autogold.Want("repo:foo (lisp)", `(and "repo:foo" "(lisp)") (HeuristicParensAsPatterns,Literal)`).Equal(t, test("repo:foo (lisp)"))
	autogold.Want("repo:foo (lisp lisp())", `(and "repo:foo" "(lisp lisp())") (HeuristicParensAsPatterns,Literal)`).Equal(t, test("repo:foo (lisp lisp())"))
	autogold.Want("repo:foo (lisp or lisp)", `(and "repo:foo" (or "lisp" "lisp")) (Literal)`).Equal(t, test("repo:foo (lisp or lisp)"))
	autogold.Want("repo:foo (lisp or lisp())", `(and "repo:foo" (or "lisp" "lisp()")) (Literal)`).Equal(t, test("repo:foo (lisp or lisp())"))
	autogold.Want("repo:foo (lisp or lisp()", `(and "repo:foo" (or "(lisp" "lisp()")) (HeuristicDanglingParens,HeuristicHoisted,Literal)`).Equal(t, test("repo:foo (lisp or lisp()"))
	autogold.Want("(y or bar())", `(or "y" "bar()") (Literal)`).Equal(t, test("(y or bar())"))
	autogold.Want("((x or bar(", `(or "((x" "bar(") (HeuristicDanglingParens,HeuristicHoisted,Literal)`).Equal(t, test("((x or bar("))
	autogold.Want("", " (None)").Equal(t, test(""))
	autogold.Want(" ", " (None)").Equal(t, test(" "))
	autogold.Want("  ", " (None)").Equal(t, test("  "))
	autogold.Want("a", `"a" (Literal)`).Equal(t, test("a"))
	autogold.Want(" a", `"a" (Literal)`).Equal(t, test(" a"))
	autogold.Want(`a `, `"a" (Literal)`).Equal(t, test(`a `))
	autogold.Want(` a b`, `(concat "a" "b") (Literal)`).Equal(t, test(` a b`))
	autogold.Want(`a  b`, `(concat "a" "b") (Literal)`).Equal(t, test(`a  b`))
	autogold.Want(`:`, `":" (Literal)`).Equal(t, test(`:`))
	autogold.Want(`:=`, `":=" (Literal)`).Equal(t, test(`:=`))
	autogold.Want(`:= range`, `(concat ":=" "range") (Literal)`).Equal(t, test(`:= range`))
	autogold.Want("`", "\"`\" (Literal)").Equal(t, test("`"))
	autogold.Want(`'`, `"'" (Literal)`).Equal(t, test(`'`))
	autogold.Want("file:a", `"file:a" (None)`).Equal(t, test("file:a"))
	autogold.Want(`"file:a"`, `"\"file:a\"" (Literal)`).Equal(t, test(`"file:a"`))
	autogold.Want(`"x foo:bar`, `(concat "\"x" "foo:bar") (Literal)`).Equal(t, test(`"x foo:bar`))
	// -repo:c" is considered valid. "repo:b is a literal pattern.
	autogold.Want(`"repo:b -repo:c"`, `(and "-repo:c\"" "\"repo:b") (Literal)`).Equal(t, test(`"repo:b -repo:c"`))
	autogold.Want(`".*"`, `"\".*\"" (Literal)`).Equal(t, test(`".*"`))
	autogold.Want(`-pattern: ok`, `(concat "-pattern:" "ok") (Literal)`).Equal(t, test(`-pattern: ok`))
	autogold.Want(`a:b "patterntype:regexp"`, `(concat "a:b" "\"patterntype:regexp\"") (Literal)`).Equal(t, test(`a:b "patterntype:regexp"`))
	autogold.Want(`not file:foo pattern`, `(and "-file:foo" "pattern") (Literal)`).Equal(t, test(`not file:foo pattern`))
	autogold.Want(`not literal.*pattern`, `(not "literal.*pattern") (Literal)`).Equal(t, test(`not literal.*pattern`))
	// Whitespace is removed. content: exists for preserving whitespace.
	autogold.Want(`lang:go func  main`, `(and "lang:go" (concat "func" "main")) (Literal)`).Equal(t, test(`lang:go func  main`))
	autogold.Want(`\n`, `"\\n" (Literal)`).Equal(t, test(`\n`))
	autogold.Want(`\t`, `"\\t" (Literal)`).Equal(t, test(`\t`))
	autogold.Want(`\\`, `"\\\\" (Literal)`).Equal(t, test(`\\`))
	autogold.Want(`foo\d "bar*"`, `(concat "foo\\d" "\"bar*\"") (Literal)`).Equal(t, test(`foo\d "bar*"`))
	autogold.Want(`\d`, `"\\d" (Literal)`).Equal(t, test(`\d`))
	autogold.Want(`type:commit message:"a commit message" after:"10 days ago"`, `(and "type:commit" "message:a commit message" "after:10 days ago") (None)`).Equal(t, test(`type:commit message:"a commit message" after:"10 days ago"`))
	autogold.Want(`type:commit message:"a commit message" after:"10 days ago" test test2`, `(and "type:commit" "message:a commit message" "after:10 days ago" (concat "test" "test2")) (Literal)`).Equal(t, test(`type:commit message:"a commit message" after:"10 days ago" test test2`))
	autogold.Want(`type:commit message:"a com"mit message" after:"10 days ago"`, `(and "type:commit" "message:a com" "after:10 days ago" (concat "mit" "message\"")) (Literal)`).Equal(t, test(`type:commit message:"a com"mit message" after:"10 days ago"`))
	autogold.Want(`bar and (foo or x\) ()`, `(or (and "bar" "(foo") (concat "x\\)" "()")) (HeuristicDanglingParens,HeuristicHoisted,Literal)`).Equal(t, test(`bar and (foo or x\) ()`))
	// For implementation simplicity, behavior preserves whitespace inside parentheses.
	autogold.Want("repo:foo (lisp    lisp)", `(and "repo:foo" "(lisp    lisp)") (HeuristicParensAsPatterns,Literal)`).Equal(t, test("repo:foo (lisp    lisp)"))
	autogold.Want("repo:foo main( or (lisp    lisp)", `(and "repo:foo" (or "main(" "(lisp    lisp)")) (HeuristicHoisted,HeuristicParensAsPatterns,Literal)`).Equal(t, test("repo:foo main( or (lisp    lisp)"))
	autogold.Want("repo:foo )foo(", "ERROR: unbalanced expression: unmatched closing parenthesis )").Equal(t, test("repo:foo )foo("))
	autogold.Want("repo:foo )main( or (lisp    lisp)", "ERROR: unbalanced expression: unmatched closing parenthesis )").Equal(t, test("repo:foo )main( or (lisp    lisp)"))
	autogold.Want("repo:foo ) main( or (lisp    lisp)", "ERROR: unbalanced expression: unmatched closing parenthesis )").Equal(t, test("repo:foo ) main( or (lisp    lisp)"))
	autogold.Want("repo:foo )))) main( or (lisp    lisp) and )))", "ERROR: unbalanced expression: unmatched closing parenthesis )").Equal(t, test("repo:foo )))) main( or (lisp    lisp) and )))"))
	autogold.Want(`repo:foo Args or main)`, "ERROR: unbalanced expression: unmatched closing parenthesis )").Equal(t, test(`repo:foo Args or main)`))
	autogold.Want(`repo:foo Args) and main`, "ERROR: unbalanced expression: unmatched closing parenthesis )").Equal(t, test(`repo:foo Args) and main`))
	autogold.Want(`repo:foo bar and baz)`, "ERROR: unbalanced expression: unmatched closing parenthesis )").Equal(t, test(`repo:foo bar and baz)`))
	autogold.Want(`repo:foo bar)) and baz`, "ERROR: unbalanced expression: unmatched closing parenthesis )").Equal(t, test(`repo:foo bar)) and baz`))
	autogold.Want(`repo:foo (bar and baz))`, "ERROR: unbalanced expression: unmatched closing parenthesis )").Equal(t, test(`repo:foo (bar and baz))`))
	autogold.Want(`repo:foo (bar and (baz)))`, "ERROR: unbalanced expression: unmatched closing parenthesis )").Equal(t, test(`repo:foo (bar and (baz)))`))
	autogold.Want(`repo:foo (bar( and baz())`, `(and "repo:foo" "bar(" "baz()") (Literal)`).Equal(t, test(`repo:foo (bar( and baz())`))
	autogold.Want(`"quoted"`, `"\"quoted\"" (Literal)`).Equal(t, test(`"quoted"`))
	// This test input should error because the single quote in 'after' is unclosed.
	autogold.Want(`type:commit message:'a commit message' after:'10 days ago" test test2`, "ERROR: unterminated literal: expected '").Equal(t, test(`type:commit message:'a commit message' after:'10 days ago" test test2`))
	// Fringe tests cases at the boundary of heuristics and invalid syntax.
	autogold.Want(`)(0 )0`, "ERROR: unbalanced expression: unmatched closing parenthesis )").Equal(t, test(`)(0 )0`))
	autogold.Want(`((R:)0))0`, "ERROR: unbalanced expression: unmatched closing parenthesis )").Equal(t, test(`((R:)0))0`))

}

func TestScanBalancedPattern(t *testing.T) {
	cases := []struct {
		Input       string
		Want        string
		WantFailure bool
	}{
		{
			Input: "foo OR bar",
			Want:  "foo",
		},
		{
			Input: "(hello there)",
			Want:  "(hello there)",
		},
		{
			Input: "( general:kenobi )",
			Want:  "( general:kenobi )",
		},
		{
			Input:       "(foo not bar)",
			WantFailure: true,
		},
		{
			Input:       "(foo OR bar)",
			WantFailure: true,
		},
		{
			Input:       "(foo not bar)",
			WantFailure: true,
		},
		{
			Input:       "repo:foo AND bar",
			WantFailure: true,
		},
		{
			Input:       "repo:foo bar",
			WantFailure: true,
		},
	}

	for _, c := range cases {
		t.Run("scan balanced pattern", func(t *testing.T) {
			want, _, ok := ScanBalancedPattern([]byte(c.Input))
			if ok && c.WantFailure {
				t.Errorf("Expected pattern to be rejected")
			}
			if diff := cmp.Diff(want, c.Want); diff != "" {
				t.Error(diff)
			}
		})
	}
}
