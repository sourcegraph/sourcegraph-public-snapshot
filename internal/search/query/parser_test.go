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
		parser := &parser{buf: []byte(input), heuristics: parensAsPatterns | balancedPattern | allowDanglingParens}
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

	// Queries with top-level parentheses, to mirror what we receive from client
	autogold.Expect(value{
		Grammar:   `(or (and "repo:foo" "count:2000" "internal") "src")`,
		Heuristic: `(and "repo:foo" "count:2000" (or "internal" "src"))`,
	}).Equal(t, test("(repo:foo count:2000 internal or src)"))
	autogold.Expect(value{
		Grammar:   `(and "lang:C++" "type:path" (or (and "repo:foo" "count:2000" "internal") "src"))`,
		Heuristic: `(and "lang:C++" "type:path" "repo:foo" "count:2000" (or "internal" "src"))`,
	}).Equal(t, test("(repo:foo count:2000 internal or src) lang:C++ type:path"))
	autogold.Expect(value{
		Grammar:   `(or (and "repo:foo" "count:2000" "internal" "limit") "src")`,
		Heuristic: "Same",
	}).Equal(t, test("((repo:foo count:2000 internal and limit) or src)"))
	autogold.Expect(value{
		Grammar:   `(and "lang:C++" "type:path" (or (and "repo:foo" "count:2000" "internal" "limit") "src"))`,
		Heuristic: "Same",
	}).Equal(t, test("((repo:foo count:2000 internal and limit) or src) lang:C++ type:path"))

	// More queries with repo
	autogold.Expect(value{
		Grammar:   `(or (and "repo:foo" "count:2000" "internal") "src")`,
		Heuristic: `(and "repo:foo" "count:2000" (or "internal" "src"))`,
	}).Equal(t, test("repo:foo count:2000 internal or src"))
	autogold.Expect(value{
		Grammar:   `(or (and "repo:foo" "count:2000" "internal") "src")`,
		Heuristic: "Same",
	}).Equal(t, test("(repo:foo count:2000 internal) or src"))
	autogold.Expect(value{
		Grammar:   `(or (and "repo:foo" "count:2000" (concat "internal" "limit")) "src")`,
		Heuristic: "Same",
	}).Equal(t, test("(repo:foo count:2000 internal limit) or src"))
	autogold.Expect(value{
		Grammar:   `(or (and "repo:foo" "count:2000" "internal" "limit") "src")`,
		Heuristic: "Same",
	}).Equal(t, test("(repo:foo count:2000 internal and limit) or src"))
	autogold.Expect(value{
		Grammar:   `(or (and "repo:foo" "count:2000" "internal" "limit") "src")`,
		Heuristic: `(and "repo:foo" "count:2000" (or (and "internal" "limit") "src"))`,
	}).Equal(t, test("repo:foo count:2000 internal and limit or src"))

	// Queries with context
	autogold.Expect(value{
		Grammar:   `(and "context:foo" "context:bar" (or (and "type:file" "a") "b"))`,
		Heuristic: `(and "context:foo" "context:bar" "type:file" (or "a" "b"))`,
	}).Equal(t, test("context:foo context:bar (type:file a or b)"))
	autogold.Expect(value{
		Grammar:   `(and "context:foo" "lang:go" (or (and "type:file" "a") "b"))`,
		Heuristic: `(and "context:foo" "lang:go" "type:file" (or "a" "b"))`,
	}).Equal(t, test("context:foo lang:go (type:file a or b) "))
	autogold.Expect(value{
		Grammar:   `(and "context:global" (or (and "type:file" "a") "b"))`,
		Heuristic: `(and "context:global" "type:file" (or "a" "b"))`,
	}).Equal(t, test("context:global (type:file a or b)"))
	autogold.Expect(value{
		Grammar:   `(or (and "context:foo" "type:file" "a") "b")`,
		Heuristic: `(and "context:foo" "type:file" (or "a" "b"))`,
	}).Equal(t, test("context:foo type:file a or b"))

	// Groups containing operators.
	autogold.Expect(value{Grammar: `(or (and "type:file" "a") "b")`, Heuristic: `(and "type:file" (or "a" "b"))`}).Equal(t, test("(type:file a or b)"))
	autogold.Expect(value{Grammar: `(or (and "type:file" "a") "b")`, Heuristic: "Same"}).Equal(t, test("(type:file a) or b"))

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
		p := &parser{buf: []byte(input), heuristics: parensAsPatterns | balancedPattern}
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
	test := func(input string, pos int) bool {
		p := &parser{buf: []byte(input), pos: pos}
		return p.matchUnaryKeyword("NOT")
	}

	testcases := []struct {
		input string
		pos   int
		want  bool
	}{
		{input: `NOT bar`, pos: 0, want: true},
		{input: `foo NOT bar`, pos: 4, want: true},
		{input: `foo NOT`, pos: 4, want: false},
		{input: `fooNOT bar`, pos: 3, want: false},
		{input: `NOTbar`, pos: 0, want: false},
		{input: `(not bar)`, pos: 1, want: true},
	}

	for _, tc := range testcases {
		t.Run(tc.input, func(t *testing.T) {
			if got := test(tc.input, tc.pos); got != tc.want {
				t.Errorf("got %t, want %t", got, tc.want)
			}
		})
	}
}

func TestParseSearchTypeKeyword(t *testing.T) {
	test := func(input string) string {
		plan, err := Pipeline(
			Init(input, SearchTypeKeyword),
		)
		if err != nil {
			return err.Error()
		}

		return plan.ToQ().String()
	}

	testcases := []struct {
		input string
		want  string
	}{
		// parens as grouping
		{input: `foo and bar and bas`, want: `(and "foo" "bar" "bas")`},
		{input: `(foo and bar) and bas`, want: `(and "foo" "bar" "bas")`},
		{input: `foo bar bas`, want: `(and "foo" "bar" "bas")`},
		{input: `(foo bar) bas`, want: `(and "foo" "bar" "bas")`},
		{input: `foo (bar bas)`, want: `(and "foo" "bar" "bas")`},
		{input: `(foo bar bas)`, want: `(and "foo" "bar" "bas")`},
		{input: `(foo) bar bas`, want: `(and "foo" "bar" "bas")`},
		{input: `foo (bar) bas`, want: `(and "foo" "bar" "bas")`},
		{input: `foo bar (bas)`, want: `(and "foo" "bar" "bas")`},

		// not
		{input: `foo and not bar`, want: `(and "foo" (not "bar"))`},
		{input: `foo (not bar)`, want: `(and "foo" (not "bar"))`},
		{input: `foo not bar`, want: `(and "foo" (not "bar"))`},

		// literal
		{input: `"(foo bar)" bas`, want: `(and "(foo bar)" "bas")`},

		// mix implicit AND and explicit OR
		{input: `(foo bar) and bas`, want: `(and "foo" "bar" "bas")`},
		{input: `(foo bar) or bas`, want: `(or (and "foo" "bar") "bas")`},
		{input: `(foo or bar) bas`, want: `(and (or "foo" "bar") "bas")`},
		{input: `(foo or bar) bas qux`, want: `(and (or "foo" "bar") "bas" "qux")`},

		// nested
		{input: `foo (bar (bas or qux))`, want: `(and "foo" "bar" (or "bas" "qux"))`},
		{input: `(foo or bas) (bar or qux) hoge`, want: `(and (or "foo" "bas") (or "bar" "qux") "hoge")`},
		{input: `(foo or (bas and qux and (hoge or fuga)))`, want: `(or "foo" (and "bas" "qux" (or "hoge" "fuga")))`},
		{input: `(foo and bas) or (hoge and fuga)`, want: `(or (and "foo" "bas") (and "hoge" "fuga"))`},

		// regex
		{input: `(foo /ba.*/) bas`, want: `(and "foo" "ba.*" "bas")`},
		{input: `(foo or /bar/) and bas`, want: `(and (or "foo" "bar") "bas")`},

		// function signatures
		{input: `func() error`, want: `(and "func()" "error")`},
		{input: `func(a int, b bool) error`, want: `(and "func(a int, b bool)" "error")`},

		// parentheses
		{input: `()`, want: `"()"`},
		{input: `(())`, want: `"()"`},
		{input: `(     )`, want: `"()"`},
		{input: `() => {}`, want: `(and "()" "=>" "{}")`},
		{input: `(err error, ok bool)`, want: `(and "err" "error," "ok" "bool")`},

		// unbalanced parentheses
		{input: `(`, want: `"("`},
		{input: `(()`, want: `"(()"`},
		{input: `())`, want: `unsupported expression. The combination of parentheses in the query has an unclear meaning. Use "..." to quote patterns that contain parentheses`},
		{input: `foo(`, want: `"foo("`},

		// unescaped quotes
		{input: `"`, want: `"\""`},
		{input: `""`, want: `""`},
		{input: `"""`, want: `"\"\"\""`},
		{input: `""""`, want: `"\"\"\"\""`},
		{input: `"""""`, want: `"\"\"\"\"\""`},
		{input: `""foo"`, want: `"\"\"foo\""`},
		{input: `""foo""`, want: `"\"\"foo\"\""`},
		{input: `"foo"bar"bas"`, want: `"\"foo\"bar\"bas\""`},

		// detect keywords at boundaries
		{input: `(a or b) and c`, want: `(and (or "a" "b") "c")`},
		{input: `(a or b)and c`, want: `(and (or "a" "b") "c")`},
		{input: `c and(a or b)`, want: `(and "c" (or "a" "b"))`},
		{input: `c and (a or b)`, want: `(and "c" (or "a" "b"))`},
		{input: `(a or b)and(c or d)`, want: `(and (or "a" "b") (or "c" "d"))`},

		{input: `(a and b) or c`, want: `(or (and "a" "b") "c")`},
		{input: `(a and b)or c`, want: `(or (and "a" "b") "c")`},
		{input: `(a and)or c`, want: `(or (and "a" "and") "c")`},
		{input: `a or(b and c)`, want: `(or "a" (and "b" "c"))`},

		{input: `(a or b) not c`, want: `(and (or "a" "b") (not "c"))`},
		{input: `(a or b)not c`, want: `(and (or "a" "b") (not "c"))`},
		{input: `(a not b)not c`, want: `(and "a" (not "b") (not "c"))`},
		{input: `not a b`, want: `(and (not "a") "b")`},
		{input: `a or not b`, want: `(or "a" (not "b"))`},
		{input: `not b`, want: `(not "b")`},
		{input: ` not b`, want: `(not "b")`},

		{input: `a or (bandc)`, want: `(or "a" "bandc")`},
		{input: `a andor b`, want: `(and "a" "andor" "b")`},
		{input: `a (and b`, want: `(and "a" "(and" "b")`},
		{input: `a )and b`, want: `unsupported expression. The combination of parentheses in the query has an unclear meaning. Use "..." to quote patterns that contain parentheses`},

		{input: `(a or b)or c`, want: `(or "a" "b" "c")`},
		{input: `(a or b) or c`, want: `(or "a" "b" "c")`},
		{input: `(a and b or c) or d`, want: `(or (and "a" "b") "c" "d")`},
		{input: `(a or b and c)or d`, want: `(or "a" (and "b" "c") "d")`},

		// first token
		{input: `  and b`, want: `(and "and" "b")`},
		{input: `and b`, want: `(and "and" "b")`},
	}

	for _, tc := range testcases {
		t.Run(tc.input, func(t *testing.T) {
			got := test(tc.input)
			if got != tc.want {
				t.Errorf("got %s, expected %s", got, tc.want)
			}
		})
	}
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

	testcases := []struct {
		input string
		want  string
	}{
		{input: `()`, want: `"()" (HeuristicParensAsPatterns,Literal)`},
		{input: `"`, want: `"\"" (Literal)`},
		{input: `""`, want: `"\"\"" (Literal)`},
		{input: `(`, want: `"(" (HeuristicDanglingParens,Literal)`},
		{input: `repo:foo foo( or bar(`, want: `(and "repo:foo" (or "foo(" "bar(")) (HeuristicHoisted,Literal)`},
		{input: `x or`, want: `(concat "x" "or") (Literal)`},
		{input: `repo:foo (x`, want: `(and "repo:foo" "(x") (HeuristicDanglingParens,Literal)`},
		{input: `(x or bar() )`, want: `(or "x" "bar()") (HeuristicHoisted,Literal)`},
		{input: `(x`, want: `"(x" (HeuristicDanglingParens,Literal)`},
		{input: `x or (x`, want: `(or "x" "(x") (HeuristicDanglingParens,Literal)`},
		{input: `(y or (z`, want: `(or "(y" "(z") (HeuristicDanglingParens,Literal)`},
		{input: `repo:foo (lisp)`, want: `(and "repo:foo" "(lisp)") (HeuristicParensAsPatterns,Literal)`},
		{input: `repo:foo (lisp lisp())`, want: `(and "repo:foo" "(lisp lisp())") (HeuristicParensAsPatterns,Literal)`},
		{input: `repo:foo (lisp or lisp)`, want: `(and "repo:foo" (or "lisp" "lisp")) (HeuristicHoisted,Literal)`},
		{input: `repo:foo (lisp or lisp())`, want: `(and "repo:foo" (or "lisp" "lisp()")) (HeuristicHoisted,Literal)`},
		{input: `repo:foo (lisp or lisp()`, want: `(and "repo:foo" (or "(lisp" "lisp()")) (HeuristicDanglingParens,HeuristicHoisted,Literal)`},
		{input: `(y or bar())`, want: `(or "y" "bar()") (HeuristicHoisted,Literal)`},
		{input: `((x or bar(`, want: `(or "((x" "bar(") (HeuristicDanglingParens,Literal)`},
		{input: ``, want: ` (None)`},
		{input: ` `, want: ` (None)`},
		{input: `  `, want: ` (None)`},
		{input: `a`, want: `"a" (Literal)`},
		{input: ` a`, want: `"a" (Literal)`},
		{input: `a `, want: `"a" (Literal)`},
		{input: ` a b`, want: `(concat "a" "b") (Literal)`},
		{input: `a  b`, want: `(concat "a" "b") (Literal)`},
		{input: `:`, want: `":" (Literal)`},
		{input: `:=`, want: `":=" (Literal)`},
		{input: `:= range`, want: `(concat ":=" "range") (Literal)`},
		{input: "`", want: "\"`\" (Literal)"},
		{input: `'`, want: `"'" (Literal)`},
		{input: `file:a`, want: `"file:a" (None)`},
		{input: `"file:a"`, want: `"\"file:a\"" (Literal)`},
		{input: `"x foo:bar`, want: `(concat "\"x" "foo:bar") (Literal)`},

		// -repo:c" is considered valid. "repo:b is a literal pattern.
		{input: `"repo:b -repo:c"`, want: `(and "-repo:c\"" "\"repo:b") (Literal)`},
		{input: `".*"`, want: `"\".*\"" (Literal)`},
		{input: `-pattern: ok`, want: `(concat "-pattern:" "ok") (Literal)`},
		{input: `a:b "patterntype:regexp"`, want: `(concat "a:b" "\"patterntype:regexp\"") (Literal)`},
		{input: `not file:foo pattern`, want: `(and "-file:foo" "pattern") (Literal)`},
		{input: `not literal.*pattern`, want: `(not "literal.*pattern") (Literal)`},

		// Whitespace is removed. content: exists for preserving whitespace.
		{input: `lang:go func  main`, want: `(and "lang:go" (concat "func" "main")) (Literal)`},
		{input: `\n`, want: `"\\n" (Literal)`},
		{input: `\t`, want: `"\\t" (Literal)`},
		{input: `\\`, want: `"\\\\" (Literal)`},
		{input: `foo\d "bar*"`, want: `(concat "foo\\d" "\"bar*\"") (Literal)`},
		{input: `\d`, want: `"\\d" (Literal)`},
		{input: `type:commit message:"a commit message" after:"10 days ago"`, want: `(and "type:commit" "message:a commit message" "after:10 days ago") (Quoted)`},
		{input: `type:commit message:"a commit message" after:"10 days ago" test test2`, want: `(and "type:commit" "message:a commit message" "after:10 days ago" (concat "test" "test2")) (Literal,Quoted)`},
		{input: `type:commit message:"a com"mit message" after:"10 days ago"`, want: `(and "type:commit" "message:a com" "after:10 days ago" (concat "mit" "message\"")) (Literal,Quoted)`},
		{input: `bar and (foo or x\) ()`, want: `(or (and "bar" "(foo") (concat "x\\)" "()")) (HeuristicDanglingParens,Literal)`},

		// For implementation simplicity, behavior preserves whitespace inside parentheses.
		{input: `repo:foo (lisp    lisp)`, want: `(and "repo:foo" "(lisp    lisp)") (HeuristicParensAsPatterns,Literal)`},
		{input: `repo:foo main( or (lisp    lisp)`, want: `(and "repo:foo" (or "main(" "(lisp    lisp)")) (HeuristicHoisted,HeuristicParensAsPatterns,Literal)`},
		{input: `repo:foo )foo(`, want: `ERROR: unsupported expression. The combination of parentheses in the query have an unclear meaning. Try using the content: filter to quote patterns that contain parentheses`},
		{input: `repo:foo ) main( or (lisp    lisp)`, want: `ERROR: unsupported expression. The combination of parentheses in the query have an unclear meaning. Try using the content: filter to quote patterns that contain parentheses`},
		{input: `repo:foo )))) main( or (lisp    lisp) and )))`, want: `ERROR: unsupported expression. The combination of parentheses in the query have an unclear meaning. Try using the content: filter to quote patterns that contain parentheses`},
		{input: `repo:foo Args or main)`, want: `ERROR: unsupported expression. The combination of parentheses in the query have an unclear meaning. Try using the content: filter to quote patterns that contain parentheses`},
		{input: `repo:foo Args) and main`, want: `ERROR: unsupported expression. The combination of parentheses in the query have an unclear meaning. Try using the content: filter to quote patterns that contain parentheses`},
		{input: `repo:foo bar and baz)`, want: `ERROR: unsupported expression. The combination of parentheses in the query have an unclear meaning. Try using the content: filter to quote patterns that contain parentheses`},
		{input: `repo:foo bar)) and baz`, want: `ERROR: unsupported expression. The combination of parentheses in the query have an unclear meaning. Try using the content: filter to quote patterns that contain parentheses`},
		{input: `repo:foo (bar and baz))`, want: `ERROR: unsupported expression. The combination of parentheses in the query have an unclear meaning. Try using the content: filter to quote patterns that contain parentheses`},
		{input: `repo:foo (bar and (baz)))`, want: `ERROR: unsupported expression. The combination of parentheses in the query have an unclear meaning. Try using the content: filter to quote patterns that contain parentheses`},
		{input: `repo:foo (bar( and baz())`, want: `(and "repo:foo" "bar(" "baz()") (HeuristicHoisted,Literal)`},
		{input: `"quoted"`, want: `"\"quoted\"" (Literal)`},
		{input: `not (stocks or stonks)`, want: `ERROR: it looks like you tried to use an expression after NOT. The NOT operator can only be used with simple search patterns or filters, and is not supported for expressions or subqueries`},

		// This test input should error because the single quote in 'after' is unclosed.
		{input: `type:commit message:'a commit message' after:'10 days ago" test test2`, want: `ERROR: unterminated literal: expected '`},

		// Fringe tests cases at the boundary of heuristics and invalid syntax.
		{input: `x()(y or z)`, want: `ERROR: unsupported expression. The combination of parentheses in the query have an unclear meaning. Try using the content: filter to quote patterns that contain parentheses`},
		{input: `)(0 )0`, want: `ERROR: unsupported expression. The combination of parentheses in the query have an unclear meaning. Try using the content: filter to quote patterns that contain parentheses`},
		{input: `((R:)0))0`, want: `ERROR: unsupported expression. The combination of parentheses in the query have an unclear meaning. Try using the content: filter to quote patterns that contain parentheses`},
	}

	for _, tc := range testcases {
		t.Run(tc.input, func(t *testing.T) {
			got := test(tc.input)
			if got != tc.want {
				t.Errorf("got %s, expected %s", got, tc.want)
			}
		})
	}
}

func TestScanBalancedPattern(t *testing.T) {
	testcases := []struct {
		input    string
		balanced bool
		want     string
	}{
		// balanced pattern
		{input: `foo OR bar`, balanced: true, want: `foo`},
		{input: `(hello there)`, balanced: true, want: `(hello there)`},
		{input: `( general:kenobi )`, balanced: true, want: `( general:kenobi )`},
		// negative cases
		{input: `(foo OR bar)`},
		{input: `(foo not bar)`},
		{input: `repo:foo AND bar`},
		{input: `repo:foo bar`},
	}
	for _, tc := range testcases {
		t.Run(tc.input, func(t *testing.T) {
			result, _, ok := ScanBalancedPattern([]byte(tc.input))
			if tc.balanced != ok {
				t.Errorf("expected %t, got %t", tc.balanced, ok)
			}
			if result != tc.want {
				t.Errorf("got %s, expected %s", result, tc.want)
			}
		})
	}
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

func TestParseKeywordPattern(t *testing.T) {
	test := func(input string) string {
		result, err := Parse(input, SearchTypeKeyword)
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

func TestIsSpace(t *testing.T) {
	cases := []struct {
		input []byte
		want  bool
	}{
		{[]byte{'\xa0'}, false},
		{[]byte{' ', ' '}, true},
		{[]byte{' ', '\t'}, true},
		{[]byte{' ', '\t', '\f', '\r'}, true},
		{[]byte{' ', '\t', '\f', '\r', 'a'}, false},
	}

	for _, tc := range cases {
		t.Run(string(tc.input), func(t *testing.T) {
			got := isSpace([]byte(tc.input))
			if got != tc.want {
				t.Errorf("got %t, want %t, first byte %d", got, tc.want, rune(tc.input[0]))
			}
		})
	}
}
