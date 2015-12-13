package gen

import (
	"go/ast"
	"go/parser"
	"testing"
)

func TestRepoURIExpr(t *testing.T) {
	tests := []struct {
		argTypeStr  string
		wantExprStr string
	}{
		{"*sourcegraph.RepoSpec", "x.URI"},
		{"*sourcegraph.RepoRevSpec", "x.URI"},
		{"*sourcegraph.BuildsCreateOp", "x.RepoRev.URI"},
		{"*sourcegraph.RepoTreeGetOp", "x.Entry.RepoRev.URI"},
		{"*sourcegraph.BuildsGetTaskLogOp", "x.Task.Build.Repo.URI"},
		{"*sourcegraph.DefsListRefsOp", "x.Def.Repo"},
		{"*sourcegraph.UserSpec", ""},
		{"*sourcegraph.UsersListOptions", ""},
	}
	for _, test := range tests {
		argType, err := parser.ParseExpr(test.argTypeStr)
		if err != nil {
			t.Errorf("arg type %s: ParseExpr: %s", test.argTypeStr, err)
			continue
		}
		expr := RepoURIExpr(ast.NewIdent("x"), argType)
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
