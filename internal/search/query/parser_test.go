package query

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func collectLabels(nodes []Node) (result labels) {
	for _, node := range nodes {
		switch v := node.(type) {
		case Operator:
			result |= v.Annotation.Labels
			result |= collectLabels(v.Operands)
		case Pattern:
			result |= v.Annotation.Labels
		case Parameter:
			result |= v.Annotation.Labels
		}
	}
	return result
}

func labelsToString(nodes []Node) string {
	labels := collectLabels(nodes)
	return strings.Join(labels.String(), ",")
}

func TestParseParameterList(t *testing.T) {
	type value struct {
		Result       string
		ResultLabels string
		ResultRange  string
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
			gotLabels = labelsToString([]Node{resultNode})
		}

		return value{
			Result:       string(got),
			ResultLabels: gotLabels,
			ResultRange:  gotRange,
		}
	}

	autogold.Expect(value{
		Result:      `{"field":"file","value":"README.md","negated":false}`,
		ResultRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":14}}`,
	}).Equal(t, test(`file:README.md`))

	autogold.Expect(value{
		Result:      `{"field":"file","value":"README.md","negated":false}`,
		ResultRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":14}}`,
	}).Equal(t, test(`file:README.md    `))

	autogold.Expect(value{
		Result: `{"value":":foo","negated":false}`, ResultLabels: "Regexp",
		ResultRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":4}}`,
	}).Equal(t, test(`:foo`))

	autogold.Expect(value{
		Result: `{"value":"foo:","negated":false}`, ResultLabels: "Regexp",
		ResultRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":4}}`,
	}).Equal(t, test(`foo:`))

	autogold.Expect(value{
		Result:      `{"field":"file","value":"bar:baz","negated":false}`,
		ResultRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":12}}`,
	}).Equal(t, test(`file:bar:baz`))

	autogold.Expect(value{
		Result: `{"value":"-:foo","negated":false}`, ResultLabels: "Regexp",
		ResultRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":5}}`,
	}).Equal(t, test(`-:foo`))

	autogold.Expect(value{
		Result:      `{"field":"file","value":"README.md","negated":true}`,
		ResultRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":15}}`,
	}).Equal(t, test(`-file:README.md`))

	autogold.Expect(value{
		Result:      `{"field":"file","value":"README.md","negated":true}`,
		ResultRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":18}}`,
	}).Equal(t, test(`NOT file:README.md`))

	autogold.Expect(value{
		Result: `{"value":"foo:bar","negated":true}`, ResultLabels: "Regexp",
		ResultRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":11}}`,
	}).Equal(t, test(`NOT foo:bar`))
	autogold.Expect(value{
		Result:      `{"field":"content","value":"bar","negated":true}`,
		ResultRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":15}}`,
	}).Equal(t, test(`NOT content:bar`))

	autogold.Expect(value{
		Result: `{"value":"NOT","negated":true}`, ResultLabels: "Regexp",
		ResultRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":7}}`,
	}).Equal(t, test(`NOT NOT`))

	autogold.Expect(value{
		Result:       `{"value":"--foo:bar","negated":false}`,
		ResultLabels: "Regexp",
		ResultRange:  `{"start":{"line":0,"column":0},"end":{"line":0,"column":9}}`,
	}).Equal(t, test(`--foo:bar`))

	autogold.Expect(value{
		Result:       `{"value":"fie-ld:bar","negated":false}`,
		ResultLabels: "Regexp",
		ResultRange:  `{"start":{"line":0,"column":0},"end":{"line":0,"column":10}}`,
	}).Equal(t, test(`fie-ld:bar`))

	autogold.Expect(value{
		Result:       `{"value":"a\\ pattern","negated":false}`,
		ResultLabels: "Regexp",
		ResultRange:  `{"start":{"line":0,"column":0},"end":{"line":0,"column":10}}`,
	}).Equal(t, test(`a\ pattern`))

	autogold.Expect(value{
		Result: `{"value":"quoted","negated":false}`, ResultLabels: "Literal,Quoted",
		ResultRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":8}}`,
	}).Equal(t, test(`"quoted"`))

	autogold.Expect(value{
		Result: `{"value":"'","negated":false}`, ResultLabels: "Literal,Quoted",
		ResultRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":4}}`,
	}).Equal(t, test(`'\''`))

	autogold.Expect(value{
		Result:       `{"value":"foo.*bar(","negated":false}`,
		ResultLabels: "HeuristicDanglingParens,Regexp",
		ResultRange:  `{"start":{"line":0,"column":0},"end":{"line":0,"column":9}}`,
	}).Equal(t, test(`foo.*bar(`))

	autogold.Expect(value{
		Result:       `{"value":"a regex pattern","negated":false}`,
		ResultLabels: "Regexp",
		ResultRange:  `{"start":{"line":0,"column":0},"end":{"line":0,"column":17}}`,
	}).Equal(t, test(`/a regex pattern/`))

	autogold.Expect(value{
		Result:       `{"value":"Search()\\(","negated":false}`,
		ResultLabels: "Regexp",
		ResultRange:  `{"start":{"line":0,"column":0},"end":{"line":0,"column":10}}`,
	}).Equal(t, test(`Search()\(`))

	autogold.Expect(value{
		Result:       `{"value":"Search(xxx)\\\\(","negated":false}`,
		ResultLabels: "HeuristicDanglingParens,Regexp",
		ResultRange:  `{"start":{"line":0,"column":0},"end":{"line":0,"column":14}}`,
	}).Equal(t, test(`Search(xxx)\\(`))

	autogold.Expect(value{
		Result: `{"value":"book","negated":false}`, ResultLabels: "Regexp",
		ResultRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":6}}`,
	}).Equal(t, test(`/book/`))

	autogold.Expect(value{
		Result: `{"value":"//","negated":false}`, ResultLabels: "Literal",
		ResultRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":2}}`,
	}).Equal(t, test(`//`))
}

func TestScanPredicate(t *testing.T) {
	type value struct {
		Result       string
		ResultLabels string
	}

	test := func(input string) value {
		parser := &parser{buf: []byte(input), heuristics: parensAsPatterns | allowDanglingParens}
		result, err := parser.parseLeaves(Regexp)
		if err != nil {
			t.Fatal(fmt.Sprintf("Unexpected error: %s", err))
		}
		resultNode := result[0]
		got, _ := json.Marshal(resultNode)
		gotLabels := labelsToString([]Node{resultNode})

		return value{
			Result:       string(got),
			ResultLabels: gotLabels,
		}
	}

	autogold.Expect(value{
		Result:       `{"field":"repo","value":"contains.file(path:test)","negated":false}`,
		ResultLabels: "IsPredicate",
	}).Equal(t, test(`repo:contains.file(path:test)`))

	autogold.Expect(value{
		Result:       `{"field":"repo","value":"contains.path(test)","negated":false}`,
		ResultLabels: "IsPredicate",
	}).Equal(t, test(`repo:contains.path(test)`))

	autogold.Expect(value{
		Result:       `{"field":"repo","value":"contains.commit.after(last thursday)","negated":false}`,
		ResultLabels: "IsPredicate",
	}).Equal(t, test(`repo:contains.commit.after(last thursday)`))

	autogold.Expect(value{
		Result:       `{"field":"repo","value":"contains.commit.before(yesterday)","negated":false}`,
		ResultLabels: "None",
	}).Equal(t, test(`repo:contains.commit.before(yesterday)`))

	autogold.Expect(value{
		Result:       `{"field":"repo","value":"contains.file(content:\\()","negated":false}`,
		ResultLabels: "IsPredicate",
	}).Equal(t, test(`repo:contains.file(content:\()`))

	autogold.Expect(value{
		Result:       `{"field":"repo","value":"contains.file","negated":false}`,
		ResultLabels: "None",
	}).Equal(t, test(`repo:contains.file`))

	autogold.Expect(value{
		Result:       `{"Kind":1,"Operands":[{"field":"repo","value":"nopredicate","negated":false},{"value":"(file:foo","negated":false}],"Annotation":{"labels":0,"range":{"start":{"line":0,"column":0},"end":{"line":0,"column":0}}}}`,
		ResultLabels: "HeuristicDanglingParens,Regexp",
	}).Equal(t, test(`repo:nopredicate(file:foo or file:bar)`))

	autogold.Expect(value{
		Result:       `{"Kind":2,"Operands":[{"value":"abc","negated":false},{"value":"contains(file:test)","negated":false}],"Annotation":{"labels":0,"range":{"start":{"line":0,"column":0},"end":{"line":0,"column":0}}}}`,
		ResultLabels: "HeuristicDanglingParens,Regexp",
	}).Equal(t, test(`abc contains(file:test)`))

	autogold.Expect(value{
		Result:       `{"field":"r","value":"contains.file(sup)","negated":false}`,
		ResultLabels: "IsPredicate",
	}).Equal(t, test(`r:contains.file(sup)`))

	autogold.Expect(value{
		Result:       `{"field":"r","value":"has(key:value)","negated":false}`,
		ResultLabels: "IsPredicate",
	}).Equal(t, test(`r:has(key:value)`))

	autogold.Expect(value{
		Result:       `{"field":"r","value":"has.tag(tag)","negated":false}`,
		ResultLabels: "IsPredicate",
	}).Equal(t, test(`r:has.tag(tag)`))
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

	autogold.Expect(`{"Field":"repo","Negated":false,"Advance":5}`).Equal(t, test("repo:foo"))
	autogold.Expect(`{"Field":"RepO","Negated":false,"Advance":5}`).Equal(t, test("RepO:foo"))
	autogold.Expect(`{"Field":"after","Negated":false,"Advance":6}`).Equal(t, test("after:"))
	autogold.Expect(`{"Field":"repo","Negated":true,"Advance":6}`).Equal(t, test("-repo:"))
	autogold.Expect(`{"Field":"","Negated":false,"Advance":0}`).Equal(t, test(""))
	autogold.Expect(`{"Field":"","Negated":false,"Advance":0}`).Equal(t, test("-"))
	autogold.Expect(`{"Field":"","Negated":false,"Advance":0}`).Equal(t, test("-:"))
	autogold.Expect(`{"Field":"","Negated":false,"Advance":0}`).Equal(t, test(":"))
	autogold.Expect(`{"Field":"","Negated":false,"Advance":0}`).Equal(t, test("??:foo"))
	autogold.Expect(`{"Field":"","Negated":false,"Advance":0}`).Equal(t, test("repo"))
	autogold.Expect(`{"Field":"","Negated":false,"Advance":0}`).Equal(t, test("-repo"))
	autogold.Expect(`{"Field":"","Negated":false,"Advance":0}`).Equal(t, test("--repo:"))
	autogold.Expect(`{"Field":"","Negated":false,"Advance":0}`).Equal(t, test(":foo"))
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
	return NewOperator(nodes, And), nil
}

func TestParse(t *testing.T) {
	type value struct {
		Grammar   string
		Heuristic string
	}

	test := func(input string) value {
		var queryGrammar, queryHeuristic []Node
		var err error
		var resultGrammar, resultHeuristic string
		queryGrammar, err = parseAndOrGrammar(input) // Parse without heuristic.
		if err != nil {
			resultGrammar = err.Error()
		} else {
			resultGrammar = toString(queryGrammar)
		}

		queryHeuristic, err = Parse(input, SearchTypeRegex)
		if err != nil {
			resultHeuristic = err.Error()
		} else {
			resultHeuristic = toString(queryHeuristic)
		}

		if resultHeuristic == resultGrammar {
			resultHeuristic = "Same"
		}

		return value{
			Grammar:   resultGrammar,
			Heuristic: resultHeuristic,
		}
	}

	autogold.Expect(value{Grammar: "", Heuristic: "Same"}).Equal(t, test(""))
	autogold.Expect(value{Grammar: "", Heuristic: "Same"}).Equal(t, test("             "))
	autogold.Expect(value{Grammar: `"a"`, Heuristic: "Same"}).Equal(t, test("a"))
	autogold.Expect(value{Grammar: `(concat "a" "b")`, Heuristic: "Same"}).Equal(t, test("a b"))
	autogold.Expect(value{Grammar: `(and "a" "b" "c")`, Heuristic: "Same"}).Equal(t, test("a and b and c"))

	autogold.Expect(value{
		Grammar:   `(concat "f" "x" "oo" "a|b" "bar")`,
		Heuristic: `"(f(x)oo((a|b))bar)"`,
	}).Equal(t, test("(f(x)oo((a|b))bar)"))

	autogold.Expect(value{Grammar: `"aorb"`, Heuristic: "Same"}).Equal(t, test("aorb"))
	autogold.Expect(value{Grammar: `"aANDb"`, Heuristic: "Same"}).Equal(t, test("aANDb"))
	autogold.Expect(value{Grammar: `(concat "a" "oror" "b")`, Heuristic: "Same"}).Equal(t, test("a oror b"))

	autogold.Expect(value{
		Grammar:   `(or (and "a" "b" "c") (and "d" (concat (or "e" "f") "g" "h" "i")) "j")`,
		Heuristic: "Same",
	}).Equal(t, test("a and b AND c or d and (e OR f) g h i or j"))

	autogold.Expect(value{
		Grammar:   `(or (and "a" "b") (and "c" "d") "e")`,
		Heuristic: "Same",
	}).Equal(t, test("a and b or c and d or e"))

	autogold.Expect(value{
		Grammar:   `(or (and "a" "b") (and "c" "d") "e")`,
		Heuristic: "Same",
	}).Equal(t, test("(a and b or c and d) or e"))

	autogold.Expect(value{Grammar: `(or (and "a" "b") "c" "d")`, Heuristic: "Same"}).Equal(t, test("(a and b or c) or d"))

	autogold.Expect(value{
		Grammar:   `(or (and "a" "b") (and "c" "d") "f" "e")`,
		Heuristic: "Same",
	}).Equal(t, test("(a and b or (c and d or f)) or e"))

	autogold.Expect(value{
		Grammar:   `(or (and "a" (or "b" "c") "d") "e")`,
		Heuristic: "Same",
	}).Equal(t, test("(a and (b or c) and d) or e"))

	autogold.Expect(value{Grammar: `(and (concat "a" "b" "c") "d")`, Heuristic: `(and "(((a b c)))" "d")`}).Equal(t, test("(((a b c))) and d"))

	// Partition parameters and concatenated patterns.
	autogold.Expect(value{Grammar: `(concat "a" (and "b" "c") "d")`, Heuristic: "Same"}).Equal(t, test("a (b and c) d"))

	autogold.Expect(value{
		Grammar:   `(and (concat "a" "b" "c") (concat "d" "e" "f") (concat "g" "h" "i"))`,
		Heuristic: `(and "(a b c)" "(d e f)" "(g h i)")`,
	}).Equal(t, test("(a b c) and (d e f) and (g h i)"))

	autogold.Expect(value{
		Grammar:   `(and "repo:foo" (concat "a" "b"))`,
		Heuristic: `(and "repo:foo" (concat "(a)" "(b)"))`,
	}).Equal(t, test("(a) repo:foo (b)"))

	autogold.Expect(value{Grammar: "expected operand at 15", Heuristic: `(and "repo:foo" (or "func(" "func(.*)"))`}).Equal(t, test("repo:foo func( or func(.*)"))

	autogold.Expect(value{
		Grammar:   `(and (and "repo:foo" (concat "main" "{")) (concat "bar" "{"))`,
		Heuristic: `(and "repo:foo" (concat "main" "{") (concat "bar" "{"))`,
	}).Equal(t, test("repo:foo main { and bar {"))

	autogold.Expect(value{
		Grammar:   `(concat "a" "b" (and "repo:foo" (concat "c" "d")))`,
		Heuristic: "Same",
	}).Equal(t, test("a b (repo:foo c d)"))

	autogold.Expect(value{
		Grammar:   `(concat "a" "b" (and "repo:foo" (concat "c" "d")))`,
		Heuristic: "Same",
	}).Equal(t, test("a b (c d repo:foo)"))

	autogold.Expect(value{
		Grammar:   `(and "repo:b" "repo:c" (concat "a" (and "repo:e" "repo:f" "d")))`,
		Heuristic: "Same",
	}).Equal(t, test("a repo:b repo:c (d repo:e repo:f)"))

	autogold.Expect(value{
		Grammar:   `(and "repo:b" "repo:c" "repo:e" "repo:f" "repo:g" "repo:h" "a")`,
		Heuristic: "Same",
	}).Equal(t, test("a repo:b repo:c (repo:e repo:f (repo:g repo:h))"))

	autogold.Expect(value{
		Grammar:   `(and "repo:b" "repo:c" "repo:e" "repo:f" "repo:g" "repo:h" (concat "a" "b"))`,
		Heuristic: "Same",
	}).Equal(t, test("a repo:b repo:c (repo:e repo:f (repo:g repo:h)) b"))
	autogold.Expect(value{
		Grammar:   `(and "repo:b" "repo:c" (concat "a" (and "repo:e" "repo:f" "repo:g" "repo:h" "b")))`,
		Heuristic: "Same",
	}).Equal(t, test("a repo:b repo:c (repo:e repo:f (repo:g repo:h b)) "))

	autogold.Expect(value{
		Grammar:   `(and "repo:foo" (concat "a" (and "repo:bar" (concat "b" (and "repo:qux" "c")))))`,
		Heuristic: "Same",
	}).Equal(t, test("(repo:foo a (repo:bar b (repo:qux c)))"))

	autogold.Expect(value{
		Grammar:   `(and "repo:b" "repo:c" (concat "a" (and "repo:e" "repo:f" (concat "d" "e"))))`,
		Heuristic: "Same",
	}).Equal(t, test("a repo:b repo:c (d repo:e repo:f e)"))

	// Errors.
	autogold.Expect(value{
		Grammar:   "unbalanced expression: unmatched closing parenthesis )",
		Heuristic: `(concat "(foo)" "(bar")`,
	}).Equal(t, test("(foo) (bar"))
	autogold.Expect(value{Grammar: "expected operand at 5", Heuristic: "Same"}).Equal(t, test("a or or b"))
	autogold.Expect(value{Grammar: `(and "a" "OR")`, Heuristic: "Same"}).Equal(t, test("a and OR"))
	autogold.Expect(value{Grammar: `(and "a" "b" "c" "d")`, Heuristic: "Same"}).Equal(t, test("(a and b) and (c and d)"))
	autogold.Expect(value{Grammar: `(or "a" "b" "c" "d")`, Heuristic: "Same"}).Equal(t, test("(a or b) or (c or d)"))
	autogold.Expect(value{Grammar: `(concat "a" "b" "c" "d")`, Heuristic: `(concat "(((a b c)))" "d")`}).Equal(t, test("(((a b c))) d"))
	autogold.Expect(value{Grammar: `(concat "a" "b" "c" "d")`, Heuristic: `(concat "(a b)" "c" "d")`}).Equal(t, test("(a b) c d"))
	autogold.Expect(value{Grammar: `(concat "a" "b" "c" "d")`, Heuristic: `(concat "a" "b" "(c d)")`}).Equal(t, test("a b (c d)"))
	autogold.Expect(value{Grammar: `(concat "a" "b" "c" "d")`, Heuristic: `(concat "(a b)" "(c d)")`}).Equal(t, test("(a b) (c d)"))

	// Escaping.
	autogold.Expect(value{Grammar: `(concat "a" "b" "c" "d" "e" "f")`, Heuristic: `(concat "(a b)" "(c d)" "(e f)")`}).Equal(t, test("(a b) (c d) (e f)"))

	autogold.Expect(value{Grammar: `(concat "a" "b" "c" "d" "e" "f")`, Heuristic: `(concat "(a b)" "c" "d" "(e f)")`}).Equal(t, test("(a b) c d (e f)"))

	autogold.Expect(value{
		Grammar:   `(and "a" "b" (or "z" "q") "c" "d" "e" "f")`,
		Heuristic: "Same",
	}).Equal(t, test("(a and b and (z or q)) and (c and d) and (e and f)"))

	autogold.Expect(value{Grammar: `""`, Heuristic: `"()"`}).Equal(t, test("()"))
	autogold.Expect(value{Grammar: `(concat "foo" "bar")`, Heuristic: `"foo()bar"`}).Equal(t, test("foo()bar"))
	autogold.Expect(value{
		Grammar:   `(and "x" (concat "regex" "s" "?"))`,
		Heuristic: `(and "x" "regex(s)?")`,
	}).Equal(t, test("(x and regex(s)?)"))

	autogold.Expect(value{Grammar: `(concat "foo" "bar")`, Heuristic: `"foo(   )bar"`}).Equal(t, test("foo(   )bar"))
	autogold.Expect(value{Grammar: `"x"`, Heuristic: `"(x())"`}).Equal(t, test("(x())"))
	autogold.Expect(value{Grammar: `"x"`, Heuristic: `"(()x(  )(())())"`}).Equal(t, test("(()x(  )(())())"))
	autogold.Expect(value{Grammar: `""`, Heuristic: `(or "()" "()")`}).Equal(t, test("() or ()"))
	autogold.Expect(value{Grammar: `"x"`, Heuristic: `(or "()" "(x)")`}).Equal(t, test("() or (x)"))
	autogold.Expect(value{Grammar: `(concat "x" (or "y" "f"))`, Heuristic: `(concat "()" "x" "()" (or "y" "()" "(f)") "()")`}).Equal(t, test("(()x(  )(y or () or (f))())"))
	autogold.Expect(value{Grammar: `""`, Heuristic: `(or "()" "()")`}).Equal(t, test("(() or ())"))

	autogold.Expect(value{
		Grammar:   `(and "r:foo" (concat "a/foo" (not ".svg")))`,
		Heuristic: "Same",
	}).Equal(t, test("r:foo (a/foo not .svg)"))

	autogold.Expect(value{Grammar: `(and "r:foo" (not ".svg"))`, Heuristic: "Same"}).Equal(t, test("r:foo (not .svg)"))

	// Escaping
	autogold.Expect(value{Grammar: `"\\(\\)"`, Heuristic: "Same"}).Equal(t, test(`\(\)`))
	autogold.Expect(value{Grammar: `(concat "\\(" "\\)")`, Heuristic: `(concat "\\(" "\\)" "()")`}).Equal(t, test(`\( \) ()`))
	autogold.Expect(value{Grammar: `"\\ "`, Heuristic: "Same"}).Equal(t, test(`\ `))
	autogold.Expect(value{Grammar: `(concat "\\ " "\\ ")`, Heuristic: "Same"}).Equal(t, test(`\  \ `))

	// Dangling parentheses heuristic.
	autogold.Expect(value{Grammar: "expected operand at 1", Heuristic: `"("`}).Equal(t, test(`(`))
	autogold.Expect(value{
		Grammar:   "unsupported expression. The combination of parentheses in the query have an unclear meaning. Try using the content: filter to quote patterns that contain parentheses",
		Heuristic: "Same",
	}).Equal(t, test(`)(())(`))
	autogold.Expect(value{Grammar: "expected operand at 5", Heuristic: `(and "foo(" "bar(")`}).Equal(t, test(`foo( and bar(`))
	autogold.Expect(value{Grammar: "expected operand at 14", Heuristic: `(and "repo:foo" (or "foo(" "bar("))`}).Equal(t, test(`repo:foo foo( or bar(`))
	autogold.Expect(value{
		Grammar:   "unsupported expression. The combination of parentheses in the query have an unclear meaning. Try using the content: filter to quote patterns that contain parentheses",
		Heuristic: "Same",
	}).Equal(t, test(`(a or (b and )) or d)`))

	// Quotes and escape sequences.
	autogold.Expect(value{Grammar: `"\""`, Heuristic: "Same"}).Equal(t, test(`"`))
	autogold.Expect(value{Grammar: `(and "repo:foo'" "bar'")`, Heuristic: "Same"}).Equal(t, test(`repo:foo' bar'`))
	autogold.Expect(value{Grammar: `(and "repo:foo" "bar")`, Heuristic: "Same"}).Equal(t, test(`repo:'foo' 'bar'`))
	autogold.Expect(value{Grammar: `(and "repo:foo" "bar")`, Heuristic: "Same"}).Equal(t, test(`repo:"foo" "bar"`))
	autogold.Expect(value{Grammar: `(and "repo:foo bar" "foo bar")`, Heuristic: "Same"}).Equal(t, test(`repo:"foo bar" "foo bar"`))
	autogold.Expect(value{Grammar: `(and "repo:fo\"o" "bar")`, Heuristic: "Same"}).Equal(t, test(`repo:"fo\"o" "bar"`))
	autogold.Expect(value{Grammar: `(and "repo:foo" "b/ar")`, Heuristic: "Same"}).Equal(t, test(`repo:foo /b\/ar/`))
	autogold.Expect(value{Grammar: `(and "repo:foo" "/a/file/path")`, Heuristic: "Same"}).Equal(t, test(`repo:foo /a/file/path`))
	autogold.Expect(value{Grammar: `(and "repo:foo" "/a/file/path/")`, Heuristic: "Same"}).Equal(t, test(`repo:foo /a/file/path/`))

	autogold.Expect(value{
		Grammar:   `(and "repo:foo" (concat "a" "/another/path/"))`,
		Heuristic: "Same",
	}).Equal(t, test(`repo:foo /a/ /another/path/`))

	autogold.Expect(value{Grammar: `(and "repo:foo" "\\s+b\\d+ar")`, Heuristic: "Same"}).Equal(t, test(`repo:foo /\s+b\d+ar/ `))
	autogold.Expect(value{Grammar: `(and "repo:foo" "bar")`, Heuristic: "Same"}).Equal(t, test(`repo:foo /bar/ `))
	autogold.Expect(value{Grammar: `"\\t\\r\\n"`, Heuristic: "Same"}).Equal(t, test(`\t\r\n`))
	autogold.Expect(value{Grammar: `(and "repo:foo\\ bar" "\\:\\\\")`, Heuristic: "Same"}).Equal(t, test(`repo:foo\ bar \:\\`))

	autogold.Expect(value{
		Grammar:   `(and "file:\\.(ts(?:(?:)|x)|js(?:(?:)|x))(?m:$)" "a")`,
		Heuristic: "Same",
	}).Equal(t, test(`a file:\.(ts(?:(?:)|x)|js(?:(?:)|x))(?m:$)`))

	autogold.Expect(value{Grammar: `(and "file:(a)" "file:(b)")`, Heuristic: "Same"}).Equal(t, test(`(file:(a) file:(b))`))
	autogold.Expect(value{Grammar: `"repohascommitafter:7 days"`, Heuristic: "Same"}).Equal(t, test(`(repohascommitafter:"7 days")`))

	autogold.Expect(value{
		Grammar:   `(and "repohascommitafter:7 days" "foo")`,
		Heuristic: "Same",
	}).Equal(t, test(`(foo repohascommitafter:"7 days")`))

	// Fringe tests cases at the boundary of heuristics and invalid syntax.
	autogold.Expect(value{
		Grammar:   "unbalanced expression: unmatched closing parenthesis )",
		Heuristic: `"(0(F)(:())(:())(<0)0()"`,
	}).Equal(t, test(`(0(F)(:())(:())(<0)0()`))

	// The space-looking character below is U+00A0.
	autogold.Expect(value{Grammar: `(concat "00" "000")`, Heuristic: `(concat "00" "(000)")`}).Equal(t, test(`00 (000)`))

}

func TestScanDelimited(t *testing.T) {
	type value struct {
		Result string
		Count  int
		ErrMsg string
	}

	test := func(input string, delimiter rune) string {
		result, count, err := ScanDelimited([]byte(input), true, delimiter)
		var errMsg string
		if err != nil {
			errMsg = err.Error()
		}
		v, _ := json.Marshal(value{result, count, errMsg})
		return string(v)
	}

	autogold.Expect(`{"Result":"","Count":2,"ErrMsg":""}`).Equal(t, test(`""`, '"'))
	autogold.Expect(`{"Result":"a","Count":3,"ErrMsg":""}`).Equal(t, test(`"a"`, '"'))
	autogold.Expect(`{"Result":"\"","Count":4,"ErrMsg":""}`).Equal(t, test(`"\""`, '"'))
	autogold.Expect(`{"Result":"\\","Count":4,"ErrMsg":""}`).Equal(t, test(`"\\""`, '"'))
	autogold.Expect(`{"Result":"","Count":5,"ErrMsg":"unterminated literal: expected \""}`).Equal(t, test(`"\\\"`, '"'))
	autogold.Expect(`{"Result":"\\\"","Count":6,"ErrMsg":""}`).Equal(t, test(`"\\\""`, '"'))
	autogold.Expect(`{"Result":"","Count":2,"ErrMsg":"unterminated literal: expected \""}`).Equal(t, test(`"a`, '"'))
	autogold.Expect(`{"Result":"","Count":3,"ErrMsg":"unrecognized escape sequence"}`).Equal(t, test(`"\?"`, '"'))
	autogold.Expect(`{"Result":"/","Count":4,"ErrMsg":""}`).Equal(t, test(`/\//`, '/'))

	// The next invocation of test needs to panic.
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for ScanDelimited")
		}
	}()
	_ = test(`a"`, '"')
}

func TestDelimited(t *testing.T) {
	inputs := []string{
		"test",
		"test\nabc",
		"test\r\nabc",
		"test\a\fabc",
		"test\t\tabc",
		"'test'",
		"\"test\"",
		"\"/test/\"",
		"/test/",
		"/test\\/abc/",
		"\\\\",
		"\\",
		"\\/",
	}
	delimiters := []rune{'/', '"', '\''}

	for _, input := range inputs {
		for _, delimiter := range delimiters {
			delimited := Delimit(input, delimiter)
			undelimited, _, err := ScanDelimited([]byte(delimited), false, delimiter)
			if err != nil {
				t.Fatal(err)
			}
			redelimited := Delimit(undelimited, delimiter)
			require.Equal(t, delimited, redelimited)
		}
	}
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

	autogold.Expect(`{"start":{"line":0,"column":0},"end":{"line":0,"column":8}}`).Equal(t, test("foo()bar"))
	autogold.Expect(`{"start":{"line":0,"column":0},"end":{"line":0,"column":5}}`).Equal(t, test("()bar"))
}

func TestMatchUnaryKeyword(t *testing.T) {
	test := func(input string, pos int) string {
		p := &parser{buf: []byte(input), pos: pos}
		return fmt.Sprintf("%t", p.matchUnaryKeyword("NOT"))
	}

	autogold.Expect("true").Equal(t, test("NOT bar", 0))
	autogold.Expect("true").Equal(t, test("foo NOT bar", 4))
	autogold.Expect("false").Equal(t, test("foo NOT", 4))
	autogold.Expect("false").Equal(t, test("fooNOT bar", 3))
	autogold.Expect("false").Equal(t, test("NOTbar", 0))
	autogold.Expect("true").Equal(t, test("(not bar)", 1))
}

func TestParseAndOrLiteral(t *testing.T) {
	test := func(input string) string {
		result, err := Parse(input, SearchTypeLiteral)
		if err != nil {
			return fmt.Sprintf("ERROR: %s", err.Error())
		}
		wantLabels := labelsToString(result)
		var resultStr []string
		for _, node := range result {
			resultStr = append(resultStr, node.String())
		}
		want := strings.Join(resultStr, " ")
		if wantLabels != "" {
			return fmt.Sprintf("%s (%s)", want, wantLabels)
		}
		return want
	}

	autogold.Expect(`"()" (HeuristicParensAsPatterns,Literal)`).Equal(t, test("()"))
	autogold.Expect(`"\"" (Literal)`).Equal(t, test(`"`))
	autogold.Expect(`"\"\"" (Literal)`).Equal(t, test(`""`))
	autogold.Expect(`"(" (HeuristicDanglingParens,Literal)`).Equal(t, test("("))
	autogold.Expect(`(and "repo:foo" (or "foo(" "bar(")) (HeuristicHoisted,Literal)`).Equal(t, test("repo:foo foo( or bar("))
	autogold.Expect(`(concat "x" "or") (Literal)`).Equal(t, test("x or"))
	autogold.Expect(`(and "repo:foo" "(x") (HeuristicDanglingParens,Literal)`).Equal(t, test("repo:foo (x"))
	autogold.Expect(`(or "x" "bar()") (Literal)`).Equal(t, test("(x or bar() )"))
	autogold.Expect(`"(x" (HeuristicDanglingParens,Literal)`).Equal(t, test("(x"))
	autogold.Expect(`(or "x" "(x") (HeuristicDanglingParens,Literal)`).Equal(t, test("x or (x"))
	autogold.Expect(`(or "(y" "(z") (HeuristicDanglingParens,Literal)`).Equal(t, test("(y or (z"))
	autogold.Expect(`(and "repo:foo" "(lisp)") (HeuristicParensAsPatterns,Literal)`).Equal(t, test("repo:foo (lisp)"))
	autogold.Expect(`(and "repo:foo" "(lisp lisp())") (HeuristicParensAsPatterns,Literal)`).Equal(t, test("repo:foo (lisp lisp())"))
	autogold.Expect(`(and "repo:foo" (or "lisp" "lisp")) (Literal)`).Equal(t, test("repo:foo (lisp or lisp)"))
	autogold.Expect(`(and "repo:foo" (or "lisp" "lisp()")) (Literal)`).Equal(t, test("repo:foo (lisp or lisp())"))
	autogold.Expect(`(and "repo:foo" (or "(lisp" "lisp()")) (HeuristicDanglingParens,HeuristicHoisted,Literal)`).Equal(t, test("repo:foo (lisp or lisp()"))
	autogold.Expect(`(or "y" "bar()") (Literal)`).Equal(t, test("(y or bar())"))
	autogold.Expect(`(or "((x" "bar(") (HeuristicDanglingParens,Literal)`).Equal(t, test("((x or bar("))
	autogold.Expect(" (None)").Equal(t, test(""))
	autogold.Expect(" (None)").Equal(t, test(" "))
	autogold.Expect(" (None)").Equal(t, test("  "))
	autogold.Expect(`"a" (Literal)`).Equal(t, test("a"))
	autogold.Expect(`"a" (Literal)`).Equal(t, test(" a"))
	autogold.Expect(`"a" (Literal)`).Equal(t, test(`a `))
	autogold.Expect(`(concat "a" "b") (Literal)`).Equal(t, test(` a b`))
	autogold.Expect(`(concat "a" "b") (Literal)`).Equal(t, test(`a  b`))
	autogold.Expect(`":" (Literal)`).Equal(t, test(`:`))
	autogold.Expect(`":=" (Literal)`).Equal(t, test(`:=`))
	autogold.Expect(`(concat ":=" "range") (Literal)`).Equal(t, test(`:= range`))
	autogold.Expect("\"`\" (Literal)").Equal(t, test("`"))
	autogold.Expect(`"'" (Literal)`).Equal(t, test(`'`))
	autogold.Expect(`"file:a" (None)`).Equal(t, test("file:a"))
	autogold.Expect(`"\"file:a\"" (Literal)`).Equal(t, test(`"file:a"`))
	autogold.Expect(`(concat "\"x" "foo:bar") (Literal)`).Equal(t, test(`"x foo:bar`))

	// -repo:c" is considered valid. "repo:b is a literal pattern.
	autogold.Expect(`(and "-repo:c\"" "\"repo:b") (Literal)`).Equal(t, test(`"repo:b -repo:c"`))
	autogold.Expect(`"\".*\"" (Literal)`).Equal(t, test(`".*"`))
	autogold.Expect(`(concat "-pattern:" "ok") (Literal)`).Equal(t, test(`-pattern: ok`))
	autogold.Expect(`(concat "a:b" "\"patterntype:regexp\"") (Literal)`).Equal(t, test(`a:b "patterntype:regexp"`))
	autogold.Expect(`(and "-file:foo" "pattern") (Literal)`).Equal(t, test(`not file:foo pattern`))
	autogold.Expect(`(not "literal.*pattern") (Literal)`).Equal(t, test(`not literal.*pattern`))

	// Whitespace is removed. content: exists for preserving whitespace.
	autogold.Expect(`(and "lang:go" (concat "func" "main")) (Literal)`).Equal(t, test(`lang:go func  main`))
	autogold.Expect(`"\\n" (Literal)`).Equal(t, test(`\n`))
	autogold.Expect(`"\\t" (Literal)`).Equal(t, test(`\t`))
	autogold.Expect(`"\\\\" (Literal)`).Equal(t, test(`\\`))
	autogold.Expect(`(concat "foo\\d" "\"bar*\"") (Literal)`).Equal(t, test(`foo\d "bar*"`))
	autogold.Expect(`"\\d" (Literal)`).Equal(t, test(`\d`))
	autogold.Expect(`(and "type:commit" "message:a commit message" "after:10 days ago") (Quoted)`).Equal(t, test(`type:commit message:"a commit message" after:"10 days ago"`))
	autogold.Expect(`(and "type:commit" "message:a commit message" "after:10 days ago" (concat "test" "test2")) (Literal,Quoted)`).Equal(t, test(`type:commit message:"a commit message" after:"10 days ago" test test2`))
	autogold.Expect(`(and "type:commit" "message:a com" "after:10 days ago" (concat "mit" "message\"")) (Literal,Quoted)`).Equal(t, test(`type:commit message:"a com"mit message" after:"10 days ago"`))
	autogold.Expect(`(or (and "bar" "(foo") (concat "x\\)" "()")) (HeuristicDanglingParens,Literal)`).Equal(t, test(`bar and (foo or x\) ()`))

	// For implementation simplicity, behavior preserves whitespace inside parentheses.
	autogold.Expect(`(and "repo:foo" "(lisp    lisp)") (HeuristicParensAsPatterns,Literal)`).Equal(t, test("repo:foo (lisp    lisp)"))
	autogold.Expect(`(and "repo:foo" (or "main(" "(lisp    lisp)")) (HeuristicHoisted,HeuristicParensAsPatterns,Literal)`).Equal(t, test("repo:foo main( or (lisp    lisp)"))
	autogold.Expect("ERROR: unsupported expression. The combination of parentheses in the query have an unclear meaning. Try using the content: filter to quote patterns that contain parentheses").Equal(t, test("repo:foo )foo("))
	autogold.Expect("ERROR: unsupported expression. The combination of parentheses in the query have an unclear meaning. Try using the content: filter to quote patterns that contain parentheses").Equal(t, test("repo:foo )main( or (lisp    lisp)"))
	autogold.Expect("ERROR: unsupported expression. The combination of parentheses in the query have an unclear meaning. Try using the content: filter to quote patterns that contain parentheses").Equal(t, test("repo:foo ) main( or (lisp    lisp)"))
	autogold.Expect("ERROR: unsupported expression. The combination of parentheses in the query have an unclear meaning. Try using the content: filter to quote patterns that contain parentheses").Equal(t, test("repo:foo )))) main( or (lisp    lisp) and )))"))
	autogold.Expect("ERROR: unsupported expression. The combination of parentheses in the query have an unclear meaning. Try using the content: filter to quote patterns that contain parentheses").Equal(t, test(`repo:foo Args or main)`))
	autogold.Expect("ERROR: unsupported expression. The combination of parentheses in the query have an unclear meaning. Try using the content: filter to quote patterns that contain parentheses").Equal(t, test(`repo:foo Args) and main`))
	autogold.Expect("ERROR: unsupported expression. The combination of parentheses in the query have an unclear meaning. Try using the content: filter to quote patterns that contain parentheses").Equal(t, test(`repo:foo bar and baz)`))
	autogold.Expect("ERROR: unsupported expression. The combination of parentheses in the query have an unclear meaning. Try using the content: filter to quote patterns that contain parentheses").Equal(t, test(`repo:foo bar)) and baz`))
	autogold.Expect("ERROR: unsupported expression. The combination of parentheses in the query have an unclear meaning. Try using the content: filter to quote patterns that contain parentheses").Equal(t, test(`repo:foo (bar and baz))`))
	autogold.Expect("ERROR: unsupported expression. The combination of parentheses in the query have an unclear meaning. Try using the content: filter to quote patterns that contain parentheses").Equal(t, test(`repo:foo (bar and (baz)))`))
	autogold.Expect(`(and "repo:foo" "bar(" "baz()") (Literal)`).Equal(t, test(`repo:foo (bar( and baz())`))
	autogold.Expect(`"\"quoted\"" (Literal)`).Equal(t, test(`"quoted"`))
	autogold.Expect("ERROR: it looks like you tried to use an expression after NOT. The NOT operator can only be used with simple search patterns or filters, and is not supported for expressions or subqueries").Equal(t, test(`not (stocks or stonks)`))

	// This test input should error because the single quote in 'after' is unclosed.
	autogold.Expect("ERROR: unterminated literal: expected '").Equal(t, test(`type:commit message:'a commit message' after:'10 days ago" test test2`))

	// Fringe tests cases at the boundary of heuristics and invalid syntax.
	autogold.Expect("ERROR: unsupported expression. The combination of parentheses in the query have an unclear meaning. Try using the content: filter to quote patterns that contain parentheses").Equal(t, test(`x()(y or z)`))
	autogold.Expect("ERROR: unsupported expression. The combination of parentheses in the query have an unclear meaning. Try using the content: filter to quote patterns that contain parentheses").Equal(t, test(`)(0 )0`))
	autogold.Expect("ERROR: unsupported expression. The combination of parentheses in the query have an unclear meaning. Try using the content: filter to quote patterns that contain parentheses").Equal(t, test(`((R:)0))0`))
}

func TestScanBalancedPattern(t *testing.T) {
	test := func(input string) string {
		result, _, ok := ScanBalancedPattern([]byte(input))
		if !ok {
			return "ERROR"
		}
		return result
	}

	autogold.Expect("foo").Equal(t, test("foo OR bar"))
	autogold.Expect("(hello there)").Equal(t, test("(hello there)"))
	autogold.Expect("( general:kenobi )").Equal(t, test("( general:kenobi )"))
	autogold.Expect("ERROR").Equal(t, test("(foo OR bar)"))
	autogold.Expect("ERROR").Equal(t, test("(foo not bar)"))
	autogold.Expect("ERROR").Equal(t, test("repo:foo AND bar"))
	autogold.Expect("ERROR").Equal(t, test("repo:foo bar"))
}

func Test_newOperator(t *testing.T) {
	cases := []struct {
		query string
		want  autogold.Value
	}{{
		query: `(repo:a and repo:b) (repo:d or repo:e) repo:f`,
		want:  autogold.Expect(`(and (and "repo:a" "repo:b") (or "repo:d" "repo:e") "repo:f")`),
	}, {
		query: `(a and b) and (d or e) and f`,
		want:  autogold.Expect(`(and (and "a" "b") (or "d" "e") "f")`),
	}, {
		query: `a and (b and c)`,
		want:  autogold.Expect(`(and "a" "b" "c")`),
	}}

	for _, tc := range cases {
		t.Run(tc.query, func(t *testing.T) {
			q, err := ParseRegexp(tc.query)
			require.NoError(t, err)

			got := NewOperator(q, And)
			tc.want.Equal(t, Q(got).String())
		})
	}
}

func TestParseStandard(t *testing.T) {
	test := func(input string) string {
		result, err := Parse(input, SearchTypeStandard)
		if err != nil {
			return err.Error()
		}
		jsonStr, _ := PrettyJSON(result)
		return jsonStr
	}

	t.Run("patterns are literal and slash-delimited patterns slash...slash are regexp", func(t *testing.T) {
		autogold.ExpectFile(t, autogold.Raw(test("anjou /saumur/")))
	})

	t.Run("quoted patterns are still literal", func(t *testing.T) {
		autogold.ExpectFile(t, autogold.Raw(test(`"veneto"`)))
	})

	t.Run("parens around slash...slash", func(t *testing.T) {
		autogold.ExpectFile(t, autogold.Raw(test("(sancerre and /pouilly-fume/)")))
	})
}

func TestParseNewStandard(t *testing.T) {
	test := func(input string) string {
		result, err := Parse(input, SearchTypeNewStandardRC1)
		if err != nil {
			return err.Error()
		}
		jsonStr, _ := PrettyJSON(result)
		return jsonStr
	}

	t.Run("patterns are literal and slash-delimited patterns slash...slash are regexp", func(t *testing.T) {
		autogold.ExpectFile(t, autogold.Raw(test(`anjou /saumur/`)))
	})

	t.Run("quotes which are part of the pattern have to be escaped", func(t *testing.T) {
		autogold.ExpectFile(t, autogold.Raw(test(`\"veneto\"`)))
	})

	t.Run("parens around slash...slash", func(t *testing.T) {
		autogold.ExpectFile(t, autogold.Raw(test(`(sancerre and /pouilly-fume/)`)))
	})

	t.Run("quoted patterns are interpreted literally", func(t *testing.T) {
		autogold.ExpectFile(t, autogold.Raw(test(`"foo bar"`)))
	})

	t.Run("literal quotes 1. Double quotes within single quotes", func(t *testing.T) {
		autogold.ExpectFile(t, autogold.Raw(test(`'foo "bar"'`)))
	})

	t.Run("literal quotes 2. Double quotes within double quotes", func(t *testing.T) {
		autogold.ExpectFile(t, autogold.Raw(test(`"foo \"bar\""`)))
	})
}

func TestGlobToRegex(t *testing.T) {
	type value struct {
		Result       string
		ResultLabels string
		ResultRange  string
	}

	test := func(input string) value {
		parser := &parser{buf: []byte(input), heuristics: parensAsPatterns | allowDanglingParens}
		result, err := parser.parseLeaves(Standard | Literal | GlobFilters)
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
			gotLabels = labelsToString([]Node{resultNode})
		}

		return value{
			Result:       string(got),
			ResultLabels: gotLabels,
			ResultRange:  gotRange,
		}
	}

	autogold.Expect(value{
		Result:      `{"field":"f","value":"^$","negated":false}`,
		ResultRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":2}}`,
	}).Equal(t, test(`f:`))
	autogold.Expect(value{
		Result:      `{"field":"f","value":"^ $","negated":false}`,
		ResultRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":5}}`,
	}).Equal(t, test(`f:" "`))
	autogold.Expect(value{
		Result:      `{"field":"f","value":" $","negated":false}`,
		ResultRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":6}}`,
	}).Equal(t, test(`f:"* "`))
	autogold.Expect(value{
		Result:      `{"field":"f","value":"^ ","negated":false}`,
		ResultRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":6}}`,
	}).Equal(t, test(`f:" *"`))
	autogold.Expect(value{
		Result:      `{"field":"f","value":" ","negated":false}`,
		ResultRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":7}}`,
	}).Equal(t, test(`f:"* *"`))
	autogold.Expect(value{
		Result:      `{"field":"f","value":"^foo bar$","negated":false}`,
		ResultRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":11}}`,
	}).Equal(t, test(`f:"foo bar"`))
	autogold.Expect(value{
		Result:      `{"field":"r","value":".*","negated":false}`,
		ResultRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":3}}`,
	}).Equal(t, test(`r:*`))
	autogold.Expect(value{
		Result:      `{"field":"repo","value":".*","negated":false}`,
		ResultRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":6}}`,
	}).Equal(t, test(`repo:*`))
	autogold.Expect(value{
		Result:      `{"field":"f","value":".*","negated":false}`,
		ResultRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":3}}`,
	}).Equal(t, test(`f:*`))
	autogold.Expect(value{
		Result:      `{"field":"file","value":".*","negated":false}`,
		ResultRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":6}}`,
	}).Equal(t, test(`file:*`))
	autogold.Expect(value{
		Result:      `{"field":"repo","value":"^sourcegraph$","negated":false}`,
		ResultRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":16}}`,
	}).Equal(t, test(`repo:sourcegraph`))
	autogold.Expect(value{
		Result:      `{"field":"repo","value":"^github\\.com/","negated":false}`,
		ResultRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":17}}`,
	}).Equal(t, test(`repo:github.com/*`))
	autogold.Expect(value{
		Result:      `{"field":"repo","value":"/sourcegraph$","negated":false}`,
		ResultRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":18}}`,
	}).Equal(t, test(`repo:*/sourcegraph`))
	autogold.Expect(value{
		Result:      `{"field":"file","value":"^README\\.md$","negated":false}`,
		ResultRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":14}}`,
	}).Equal(t, test(`file:README.md`))
	autogold.Expect(value{
		Result:      `{"field":"file","value":"^client/README\\.md$","negated":false}`,
		ResultRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":21}}`,
	}).Equal(t, test(`file:client/README.md`))
	autogold.Expect(value{
		Result:      `{"field":"file","value":"/Dockerfile$","negated":false}`,
		ResultRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":17}}`,
	}).Equal(t, test(`file:*/Dockerfile`))
	autogold.Expect(value{
		Result:      `{"field":"file","value":"\\.go$","negated":false}`,
		ResultRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":9}}`,
	}).Equal(t, test(`file:*.go`))
	autogold.Expect(value{
		Result:      `{"field":"file","value":"^src/","negated":false}`,
		ResultRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":10}}`,
	}).Equal(t, test(`file:src/*`))
	autogold.Expect(value{Result: `{"Kind":1,"Operands":[{"field":"context","value":"global","negated":false},{"field":"repo","value":"/sourcegraph$","negated":false},{"field":"f","value":"\\.md$","negated":false},{"value":"zoekt","negated":false}],"Annotation":{"labels":0,"range":{"start":{"line":0,"column":0},"end":{"line":0,"column":0}}}}`}).Equal(t, test(`context:global repo:*/sourcegraph zoekt f:*.md`))
	autogold.Expect(value{Result: `{"Kind":1,"Operands":[{"field":"f","value":"_test\\.go$","negated":true},{"field":"f","value":"search.*\\.go$","negated":false}],"Annotation":{"labels":0,"range":{"start":{"line":0,"column":0},"end":{"line":0,"column":0}}}}`}).Equal(t, test(`-f:*_test.go f:*search*.go`))
	// Make sure we don't convert predicates to regex patterns
	autogold.Expect(value{
		Result:      `{"field":"r","value":"has.meta(language)","negated":false}`,
		ResultRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":20}}`,
	}).Equal(t, test(`r:has.meta(language)`))
	autogold.Expect(value{
		Result:      `{"field":"r","value":"has.file(go.mod)","negated":false}`,
		ResultRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":18}}`,
	}).Equal(t, test(`r:has.file(go.mod)`))
	autogold.Expect(value{
		Result:      `{"field":"r","value":"has.content(apple)","negated":false}`,
		ResultRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":20}}`,
	}).Equal(t, test(`r:has.content(apple)`))
	autogold.Expect(value{
		Result:      `{"field":"f","value":"has.content(apple)","negated":false}`,
		ResultRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":20}}`,
	}).Equal(t, test(`f:has.content(apple)`))
	autogold.Expect(value{
		Result: `{"Kind":1,"Operands":[{"field":"r","value":"go$","negated":false},{"field":"r","value":"has.topic(language)","negated":false}],"Annotation":{"labels":0,"range":{"start":{"line":0,"column":0},"end":{"line":0,"column":0}}}}`}).Equal(t, test(`r:*go r:has.topic(language)`))
	autogold.Expect(value{
		Result:      `{"field":"r","value":"^sourcegraph$@ae3f1c","negated":false}`,
		ResultRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":20}}`,
	}).Equal(t, test(`r:sourcegraph@ae3f1c`))
	autogold.Expect(value{
		Result:      `{"field":"r","value":"^sourcegraph$@main","negated":false}`,
		ResultRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":18}}`,
	}).Equal(t, test(`r:sourcegraph@main`))
	autogold.Expect(value{
		Result:      `{"field":"r","value":"sourcegraph$@*refs/heads*","negated":false}`,
		ResultRange: `{"start":{"line":0,"column":0},"end":{"line":0,"column":27}}`,
	}).Equal(t, test(`r:*sourcegraph@*refs/heads*`))
}
