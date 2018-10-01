# astrewrite [![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](http://godoc.org/github.com/fatih/astrewrite) [![Build Status](http://img.shields.io/travis/fatih/astrewrite.svg?style=flat-square)](https://travis-ci.org/fatih/astrewrite)

astrewrite provides a `Walk()` function, similar to [ast.Inspect()](https://godoc.org/go/ast#Inspect) from the
[ast](https://godoc.org/go/ast) package. The only difference is that the passed walk function can also
return a node, which is used to rewrite the parent node.  This provides an easy
way to rewrite a given ast.Node while walking the AST.

# Example

```go
package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"

	"github.com/fatih/astrewrite"
)

func main() {
	src := `package main

type Foo struct{}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "foo.go", src, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	rewriteFunc := func(n ast.Node) (ast.Node, bool) {
		x, ok := n.(*ast.TypeSpec)
		if !ok {
			return n, true
		}

		// change struct type name to "Bar"
		x.Name.Name = "Bar"
		return x, true
	}

	rewritten := astrewrite.Walk(file, rewriteFunc)

	var buf bytes.Buffer
	printer.Fprint(&buf, fset, rewritten)
	fmt.Println(buf.String())
	// Output:
	// package main
	//
	// type Bar struct{}
}
```

