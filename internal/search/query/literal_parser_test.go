package query

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
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

func TestParseAndOrLiteral(t *testing.T) {
	cases := []struct {
		Input      string
		Want       string
		WantLabels string
		WantError  string
	}{
		{
			Input:      "()",
			Want:       `"()"`,
			WantLabels: "HeuristicParensAsPatterns,Literal",
		},
		{
			Input:      `"`,
			Want:       `"\""`,
			WantLabels: "Literal",
		},
		{
			Input:      `""`,
			Want:       `"\"\""`,
			WantLabels: "Literal",
		},
		{
			Input:      "(",
			Want:       `"("`,
			WantLabels: "HeuristicDanglingParens,Literal",
		},
		{
			Input:      "repo:foo foo( or bar(",
			Want:       `(and "repo:foo" (or "foo(" "bar("))`,
			WantLabels: "HeuristicHoisted,Literal",
		},
		{
			Input:      "x or",
			Want:       `(concat "x" "or")`,
			WantLabels: "Literal",
		},
		{
			Input:      "repo:foo (x",
			Want:       `(and "repo:foo" "(x")`,
			WantLabels: "HeuristicDanglingParens,Literal",
		},
		{
			Input:      "(x or bar() )",
			Want:       `(or "x" "bar()")`,
			WantLabels: "Literal",
		},
		{
			Input:      "(x",
			Want:       `"(x"`,
			WantLabels: "HeuristicDanglingParens,Literal",
		},
		{
			Input:      "x or (x",
			Want:       `(or "x" "(x")`,
			WantLabels: "HeuristicDanglingParens,HeuristicHoisted,Literal",
		},
		{
			Input:      "(y or (z",
			Want:       `(or "(y" "(z")`,
			WantLabels: "HeuristicDanglingParens,HeuristicHoisted,Literal",
		},
		{
			Input:      "repo:foo (lisp)",
			Want:       `(and "repo:foo" "(lisp)")`,
			WantLabels: "HeuristicParensAsPatterns,Literal",
		},
		{
			Input:      "repo:foo (lisp lisp())",
			Want:       `(and "repo:foo" "(lisp lisp())")`,
			WantLabels: "HeuristicParensAsPatterns,Literal",
		},
		{
			Input:      "repo:foo (lisp or lisp)",
			Want:       `(and "repo:foo" (or "lisp" "lisp"))`,
			WantLabels: "Literal",
		},
		{
			Input:      "repo:foo (lisp or lisp())",
			Want:       `(and "repo:foo" (or "lisp" "lisp()"))`,
			WantLabels: "Literal",
		},
		{
			Input:      "repo:foo (lisp or lisp())",
			Want:       `(and "repo:foo" (or "lisp" "lisp()"))`,
			WantLabels: "Literal",
		},
		{
			Input:      "repo:foo (lisp or lisp()",
			Want:       `(and "repo:foo" (or "(lisp" "lisp()"))`,
			WantLabels: "HeuristicDanglingParens,HeuristicHoisted,Literal",
		},
		{
			Input:      "(y or bar())",
			Want:       `(or "y" "bar()")`,
			WantLabels: "Literal",
		},
		{
			Input:      "((x or bar(",
			Want:       `(or "((x" "bar(")`,
			WantLabels: "HeuristicDanglingParens,HeuristicHoisted,Literal",
		},
		{
			Input:      "",
			Want:       "",
			WantLabels: "None",
		},
		{
			Input:      " ",
			Want:       "",
			WantLabels: "None",
		},
		{
			Input:      "  ",
			Want:       "",
			WantLabels: "None",
		},
		{
			Input:      "a",
			Want:       `"a"`,
			WantLabels: "Literal",
		},
		{
			Input:      " a",
			Want:       `"a"`,
			WantLabels: "Literal",
		},
		{
			Input:      `a `,
			Want:       `"a"`,
			WantLabels: "Literal",
		},
		{
			Input:      ` a b`,
			Want:       `(concat "a" "b")`,
			WantLabels: "Literal",
		},
		{
			Input:      `a  b`,
			Want:       `(concat "a" "b")`,
			WantLabels: "Literal",
		},
		{
			Input:      `:`,
			Want:       `":"`,
			WantLabels: "Literal",
		},
		{
			Input:      `:=`,
			Want:       `":="`,
			WantLabels: "Literal",
		},
		{
			Input:      `:= range`,
			Want:       `(concat ":=" "range")`,
			WantLabels: "Literal",
		},
		{
			Input:      "`",
			Want:       "\"`\"",
			WantLabels: "Literal",
		},
		{
			Input:      `'`,
			Want:       `"'"`,
			WantLabels: "Literal",
		},
		{
			Input:      "file:a",
			Want:       `"file:a"`,
			WantLabels: "None",
		},
		{
			Input:      `"file:a"`,
			Want:       `"\"file:a\""`,
			WantLabels: "Literal",
		},
		{
			Input:      `"x foo:bar`,
			Want:       `(concat "\"x" "foo:bar")`,
			WantLabels: "Literal",
		},
		// -repo:c" is considered valid. "repo:b is a literal pattern.
		{
			Input:      `"repo:b -repo:c"`,
			Want:       `(and "-repo:c\"" "\"repo:b")`,
			WantLabels: "Literal",
		},
		{
			Input:      `".*"`,
			Want:       `"\".*\""`,
			WantLabels: "Literal",
		},
		{
			Input:      `-pattern: ok`,
			Want:       `(concat "-pattern:" "ok")`,
			WantLabels: "Literal",
		},
		{
			Input:      `a:b "patterntype:regexp"`,
			Want:       `(concat "a:b" "\"patterntype:regexp\"")`,
			WantLabels: "Literal",
		},
		// Whitespace is removed. content: exists for preserving whitespace.
		{
			Input:      `lang:go func  main`,
			Want:       `(and "lang:go" (concat "func" "main"))`,
			WantLabels: "Literal",
		},
		{
			Input:      `\n`,
			Want:       `"\\n"`,
			WantLabels: "Literal",
		},
		{
			Input:      `\t`,
			Want:       `"\\t"`,
			WantLabels: "Literal",
		},
		{
			Input:      `\\`,
			Want:       `"\\\\"`,
			WantLabels: "Literal",
		},
		{
			Input:      `foo\d "bar*"`,
			Want:       `(concat "foo\\d" "\"bar*\"")`,
			WantLabels: "Literal",
		},
		{
			Input:      `\d`,
			Want:       `"\\d"`,
			WantLabels: "Literal",
		},
		{
			Input:      `type:commit message:"a commit message" after:"10 days ago"`,
			Want:       `(and "type:commit" "message:a commit message" "after:10 days ago")`,
			WantLabels: "None",
		},
		{
			Input:      `type:commit message:"a commit message" after:"10 days ago" test test2`,
			Want:       `(and "type:commit" "message:a commit message" "after:10 days ago" (concat "test" "test2"))`,
			WantLabels: "Literal",
		},
		{
			Input:      `type:commit message:'a commit message' after:'10 days ago' test test2`,
			Want:       `(and "type:commit" "message:a commit message" "after:10 days ago" (concat "test" "test2"))`,
			WantLabels: "Literal",
		},
		{
			Input:      `type:commit message:"a com"mit message" after:"10 days ago"`,
			Want:       `(and "type:commit" "message:a com" "after:10 days ago" (concat "mit" "message\""))`,
			WantLabels: "Literal",
		},
		{
			Input:      `bar and (foo or x\) ()`,
			Want:       `(or (and "bar" "(foo") (concat "x\\)" "()"))`,
			WantLabels: "HeuristicDanglingParens,HeuristicHoisted,HeuristicParensAsPatterns,Literal",
		},
		// For implementation simplicity, behavior preserves whitespace
		// inside parentheses.
		{
			Input:      "repo:foo (lisp    lisp)",
			Want:       `(and "repo:foo" "(lisp    lisp)")`,
			WantLabels: "HeuristicParensAsPatterns,Literal",
		},
		{
			Input:      "repo:foo main( or (lisp    lisp)",
			Want:       `(and "repo:foo" (or "main(" "(lisp    lisp)"))`,
			WantLabels: "HeuristicHoisted,HeuristicParensAsPatterns,Literal",
		},
		{
			Input:      "repo:foo )main( or (lisp    lisp)",
			Want:       `(and "repo:foo" (or ")main(" "(lisp    lisp)"))`,
			WantLabels: "HeuristicDanglingParens,HeuristicHoisted,HeuristicParensAsPatterns,Literal",
		},
		{
			Input:      "repo:foo ) main( or (lisp    lisp)",
			Want:       `(and "repo:foo" (or (concat ")" "main(") "(lisp    lisp)"))`,
			WantLabels: "HeuristicDanglingParens,HeuristicHoisted,HeuristicParensAsPatterns,Literal",
		},
		{
			Input:      "repo:foo )))) main( or (lisp    lisp) and )))",
			Want:       `(and "repo:foo" (or (concat "))))" "main(") (and "(lisp    lisp)" ")))")))`,
			WantLabels: "HeuristicDanglingParens,HeuristicHoisted,HeuristicParensAsPatterns,Literal",
		},

		{
			Input:      `"quoted"`,
			Want:       `"\"quoted\""`,
			WantLabels: "Literal",
		},
		{
			Input:      `repo:foo Args or main)`,
			Want:       `(and "repo:foo" (or "Args" "main)"))`,
			WantLabels: "HeuristicDanglingParens,HeuristicHoisted,Literal",
		},
		{
			Input:      `repo:foo Args) and main`,
			Want:       `(and "repo:foo" "Args)" "main")`,
			WantLabels: "HeuristicDanglingParens,Literal",
		},
		{
			Input:      `repo:foo bar and baz)`,
			Want:       `(and "repo:foo" "bar" "baz)")`,
			WantLabels: "HeuristicDanglingParens,Literal",
		},
		{
			Input:      `repo:foo bar)) and baz`,
			Want:       `(and "repo:foo" "bar))" "baz")`,
			WantLabels: "HeuristicDanglingParens,Literal",
		},
		{
			Input:      `repo:foo (bar( and baz())`,
			Want:       `(and "repo:foo" "bar(" "baz()")`,
			WantLabels: "Literal",
		},
		{
			Input:      `repo:foo (bar and baz))`,
			WantError:  `i'm having trouble understanding that query. The combination of parentheses is the problem. Try using the content: filter to quote patterns that contain parentheses`,
			WantLabels: "None",
		},
		{
			Input:      `repo:foo (bar and (baz)))`,
			WantError:  `i'm having trouble understanding that query. The combination of parentheses is the problem. Try using the content: filter to quote patterns that contain parentheses`,
			WantLabels: "None",
		},
		// This test input should error because the single quote in 'after' is unclosed.
		{
			Input:      `type:commit message:'a commit message' after:'10 days ago" test test2`,
			WantError:  "unterminated literal: expected '",
			WantLabels: "None",
		},
		// Fringe tests cases at the boundary of heuristics and invalid syntax.
		{
			Input:      `)(0 )0`,
			Want:       `(concat ")(0" ")0")`,
			WantLabels: "HeuristicDanglingParens,Literal",
		},
		{
			Input:      `((R:)0))0`,
			WantError:  `invalid query syntax`,
			WantLabels: "None",
		},
	}
	for _, tt := range cases {
		t.Run("literal search parse", func(t *testing.T) {
			result, err := ParseAndOr(tt.Input, SearchTypeLiteral)
			if err != nil {
				if diff := cmp.Diff(tt.WantError, err.Error()); diff != "" {
					t.Error(diff)
				}
			}
			var resultStr []string
			for _, node := range result {
				resultStr = append(resultStr, node.String())
			}
			got := strings.Join(resultStr, " ")
			if diff := cmp.Diff(tt.Want, got); diff != "" {
				t.Error(diff)
			}
			gotLabels := heuristicLabels(result)
			if diff := cmp.Diff(tt.WantLabels, gotLabels); diff != "" {
				t.Error(diff)
			}
		})
	}
}
