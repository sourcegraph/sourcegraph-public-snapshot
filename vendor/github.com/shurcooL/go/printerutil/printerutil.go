// Package printerutil provides formatted printing of AST nodes.
package printerutil

import (
	"bytes"
	"fmt"
	"go/printer"
	"go/token"
)

// Consistent with the default gofmt behavior.
var config = printer.Config{Mode: printer.UseSpaces | printer.TabIndent, Tabwidth: 8}

// SprintAst prints node, using fset, and returns it as string.
func SprintAst(fset *token.FileSet, node interface{}) string {
	var buf bytes.Buffer
	config.Fprint(&buf, fset, node)
	return buf.String()
}

// SprintAstBare prints node and returns it as string.
func SprintAstBare(node interface{}) string {
	fset := token.NewFileSet()
	return SprintAst(fset, node)
}

// PrintlnAst prints node, using fset, to stdout.
func PrintlnAst(fset *token.FileSet, node interface{}) {
	fmt.Println(SprintAst(fset, node))
}

// PrintlnAstBare prints node to stdout.
func PrintlnAstBare(node interface{}) {
	fset := token.NewFileSet()
	PrintlnAst(fset, node)
}
