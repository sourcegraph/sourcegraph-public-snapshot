package main

import (
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/astrewrite"
)

const (
	oldPrefix = "\"github.com/sourcegraph/sourcegraph/"
	newPrefix = "\"sourcegraph.com/"
)

func update(path string) error {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	rewriteFunc := func(n ast.Node) (ast.Node, bool) {
		d, ok := n.(*ast.GenDecl)
		if !ok || d.Tok != token.IMPORT {
			return n, true
		}

		for _, s := range d.Specs {
			imp := s.(*ast.ImportSpec)
			if strings.HasPrefix(imp.Path.Value, oldPrefix) {
				imp.Path.Value = newPrefix + imp.Path.Value[len(oldPrefix):]
			}
		}

		return n, false
	}

	rewritten := astrewrite.Walk(file, rewriteFunc)
	ast.SortImports(fset, file)

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	err = printer.Fprint(f, fset, rewritten)
	if err != nil {
		return err
	}
	return f.Close()
}

func main() {
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		return update(path)
	})
	if err != nil {
		log.Fatal(err)
	}
}
