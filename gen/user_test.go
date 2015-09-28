package gen

import (
	"go/ast"
	"go/parser"
	"testing"
)

func TestUserSpecExpr(t *testing.T) {
	tests := []struct {
		argTypeStr  string
		wantExprStr string
	}{
		{"*sourcegraph.UserSpec", "*x"},
		{"*sourcegraph.OrgsListOp", "x.Member"},
		{"*sourcegraph.RepoSpec", ""},
		{"*sourcegraph.RepoListOptions", ""},
	}
	for _, test := range tests {
		argType, err := parser.ParseExpr(test.argTypeStr)
		if err != nil {
			t.Errorf("arg type %s: ParseExpr: %s", test.argTypeStr, err)
			continue
		}
		expr := UserSpecExpr(ast.NewIdent("x"), argType)
		if expr == nil && test.wantExprStr == "" {
			continue
		}
		if expr == nil {
			t.Errorf("arg type %s: got nil, want %s", test.argTypeStr, test.wantExprStr)
			continue
		}
		if exprStr := AstString(expr); exprStr != test.wantExprStr {
			t.Errorf("arg type %s: got %s, want %s", test.argTypeStr, exprStr, test.wantExprStr)
		}
	}
}
