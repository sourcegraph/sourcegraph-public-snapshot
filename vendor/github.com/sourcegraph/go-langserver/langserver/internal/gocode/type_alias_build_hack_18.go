// +build !go1.9,!go1.8.typealias

package gocode

import (
	"go/ast"
)

func typeAliasSpec(name string, typ ast.Expr) *ast.TypeSpec {
	return &ast.TypeSpec{
		Name: ast.NewIdent(name),
		Type: typ,
	}
}

func isAliasTypeSpec(t *ast.TypeSpec) bool {
	return false
}
