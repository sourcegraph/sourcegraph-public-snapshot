// Package gist5639599 provides formatted printing of AST nodes.
package gist5639599

import (
	"bytes"
	"fmt"
	"go/printer"
	"go/token"
)

// Consistent with the default gofmt behavior.
var config = printer.Config{Mode: printer.UseSpaces | printer.TabIndent, Tabwidth: 8}

func SprintAst(fset *token.FileSet, node interface{}) string {
	var buf bytes.Buffer
	config.Fprint(&buf, fset, node)
	return buf.String()
}

func SprintAstBare(node interface{}) string {
	fset := token.NewFileSet()
	return SprintAst(fset, node)
}

func PrintlnAst(fset *token.FileSet, node interface{}) {
	fmt.Println(SprintAst(fset, node))
}

func PrintlnAstBare(node interface{}) {
	fset := token.NewFileSet()
	PrintlnAst(fset, node)
}
