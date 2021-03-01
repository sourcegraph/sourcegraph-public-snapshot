package main

// This is a temporary script to migrate our go import paths from
// sourcegraph.com to github.com

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/imports"

	"github.com/fatih/astrewrite"

	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
)

var vendorRe = lazyregexp.New(`(/|^)vendor/`)

func main() {
	rewriteFunc := func(n ast.Node) (ast.Node, bool) {
		x, ok := n.(*ast.ImportSpec)
		if !ok {
			return n, true
		}
		if strings.HasPrefix(x.Path.Value, "\"sourcegraph.com/sourcegraph/") && !strings.Contains(x.Path.Value, "go-diff") {
			x.Path.Value = "\"github.com" + x.Path.Value[len("\"sourcegraph.com"):]
		}
		return x, true
	}

	err := filepath.Walk(os.Args[1], func(path string, f os.FileInfo, err error) error {
		if err != nil || !strings.HasSuffix(path, ".go") || vendorRe.MatchString(path) {
			return err
		}

		src, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, src, parser.ParseComments)
		if err != nil {
			return err
		}

		// Rewrite imports
		rewritten := astrewrite.Walk(file, rewriteFunc)

		// Import order may need changing. So format with goimports which will
		// sort imports.
		var buf bytes.Buffer
		printer.Fprint(&buf, fset, rewritten)
		res, err := imports.Process(path, buf.Bytes(), &imports.Options{Comments: true, TabIndent: true, TabWidth: 8, FormatOnly: true})
		if err != nil {
			return err
		}

		if !bytes.Equal(src, res) {
			fmt.Printf("updating %s\n", path)
			return ioutil.WriteFile(path, res, f.Mode().Perm())
		}

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}
