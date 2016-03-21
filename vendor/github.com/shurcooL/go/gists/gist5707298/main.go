// Package gist5707298 offers convenience functions for parsing Go code to AST.
package gist5707298

import (
	"go/ast"
	"go/parser"
	"go/token"
)

// ParseStmt is a convenience function for obtaining the AST of a statement x.
// The position information recorded in the AST is undefined.
func ParseStmt(x string) (ast.Stmt, error) {
	file, err := parser.ParseFile(token.NewFileSet(), "", "package p;func _(){\n//line :1\n"+x+"\n;}", 0)
	if err != nil {
		return nil, err
	}
	return file.Decls[0].(*ast.FuncDecl).Body.List[0], nil
}

// ParseDecl is a convenience function for obtaining the AST of a declaration x.
// The position information recorded in the AST is undefined.
func ParseDecl(x string) (ast.Decl, error) {
	file, err := parser.ParseFile(token.NewFileSet(), "", "package p\n//line :1\n"+x+"\n", 0)
	if err != nil {
		return nil, err
	}
	return file.Decls[0], nil
}
