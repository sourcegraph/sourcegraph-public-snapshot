// Package parserutil offers convenience functions for parsing Go code to AST.
package parserutil

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
)

// ParseStmt is a convenience function for obtaining the AST of a statement x.
// The position information recorded in the AST is undefined. The filename used
// in error messages is the empty string.
func ParseStmt(x string) (ast.Stmt, error) {
	file, err := parser.ParseFile(token.NewFileSet(), "", "package p;func _(){\n//line :1\n"+x+"\n;}", 0)
	if err != nil {
		return nil, err
	}
	return file.Decls[0].(*ast.FuncDecl).Body.List[0], nil
}

// ParseDecl is a convenience function for obtaining the AST of a declaration x.
// The position information recorded in the AST is undefined. The filename used
// in error messages is the empty string.
func ParseDecl(x string) (ast.Decl, error) {
	file, err := parser.ParseFile(token.NewFileSet(), "", "package p\n//line :1\n"+x+"\n", 0)
	if err != nil {
		return nil, err
	}
	if len(file.Decls) == 0 {
		return nil, errors.New("no declaration")
	}
	return file.Decls[0], nil
}
