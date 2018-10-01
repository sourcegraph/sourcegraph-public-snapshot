package syntax

import (
	"reflect"
	"testing"
)

func TestParser(t *testing.T) {
	tests := map[string]struct {
		wantExpr   []*Expr
		wantString string
		wantErr    *ParseError
	}{
		"":   {wantExpr: []*Expr{}},
		" ":  {wantExpr: []*Expr{}, wantString: ""},
		"  ": {wantExpr: []*Expr{}, wantString: ""},
		"a": {
			wantExpr: []*Expr{{Value: "a", ValueType: TokenLiteral}},
		},
		"a ": {
			wantExpr:   []*Expr{{Value: "a", ValueType: TokenLiteral}},
			wantString: "a",
		},
		"a:": {wantExpr: []*Expr{{Field: "a", Value: "", ValueType: TokenLiteral}}},
		"a-": {
			wantExpr: []*Expr{{Value: "a-", ValueType: TokenLiteral}},
		},
		`"a"`: {
			wantExpr: []*Expr{{Value: `"a"`, ValueType: TokenQuoted}},
		},
		"-a": {
			wantExpr: []*Expr{{Not: true, Value: "a", ValueType: TokenLiteral}},
		},
		"a:b": {
			wantExpr: []*Expr{{Field: "a", Value: "b", ValueType: TokenLiteral}},
		},
		"a:b-:": {
			wantExpr: []*Expr{{Field: "a", Value: "b-:", ValueType: TokenLiteral}},
		},
		`a:"b"`: {
			wantExpr: []*Expr{{Field: "a", Value: `"b"`, ValueType: TokenQuoted}},
		},
		"-a:b": {
			wantExpr: []*Expr{{Not: true, Field: "a", Value: "b", ValueType: TokenLiteral}},
		},
		"/a/": {
			wantExpr: []*Expr{{Value: "a", ValueType: TokenPattern}},
		},
		`-/a/`: {
			wantExpr: []*Expr{{Not: true, Value: "a", ValueType: TokenPattern}},
		},
		"a b": {
			wantExpr: []*Expr{
				{Value: "a", ValueType: TokenLiteral},
				{Value: "b", ValueType: TokenLiteral},
			},
		},
		"a:b c:d": {
			wantExpr: []*Expr{
				{Field: "a", Value: "b", ValueType: TokenLiteral},
				{Field: "c", Value: "d", ValueType: TokenLiteral},
			},
		},
		"a: b:": {
			wantExpr: []*Expr{
				{Field: "a", Value: "", ValueType: TokenLiteral},
				{Field: "b", Value: "", ValueType: TokenLiteral},
			},
		},
		"--": {
			wantErr: &ParseError{Pos: 1, Msg: "got TokenMinus, want expr"},
		},
		`a:"b"-`: {
			wantErr: &ParseError{Pos: 5, Msg: "got TokenMinus, want separator or EOF"},
		},
		`"a"-`: {
			wantErr: &ParseError{Pos: 3, Msg: "got TokenMinus, want separator or EOF"},
		},
		`"a":b`: {
			wantErr: &ParseError{Pos: 3, Msg: "got TokenColon, want separator or EOF"},
		},
	}
	for input, test := range tests {
		t.Run(input, func(t *testing.T) {
			query, err := Parse(input)
			if err != nil && test.wantErr == nil {
				t.Fatal(err)
			} else if err == nil && test.wantErr != nil {
				t.Fatalf("got err == nil, want %q", test.wantErr)
			} else if test.wantErr != nil && !reflect.DeepEqual(err, test.wantErr) {
				t.Fatalf("got err == %q, want %q", err, test.wantErr)
			}
			if err != nil {
				return
			}
			if len(query.Expr) == 0 {
				query.Expr = []*Expr{}
			}
			for _, expr := range query.Expr {
				expr.Pos = 0
			}
			if !reflect.DeepEqual(query.Expr, test.wantExpr) {
				t.Errorf("expr: %s\ngot  %v\nwant %v", input, query.Expr, test.wantExpr)
			}
			if test.wantString == "" && len(query.Expr) > 0 {
				test.wantString = input
			}
			if exprString := ExprString(query.Expr); exprString != test.wantString {
				t.Errorf("expr string: %s\ngot  %s\nwant %s", input, exprString, test.wantString)
			}
		})
	}
}
