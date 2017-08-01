package godef

import (
	"bytes"
	"errors"
	"fmt"
	"go/build"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"go/token"

	"go/printer"

	"go/ast"

	"github.com/sourcegraph/go-langserver/langserver/internal/godef/go/parser"
	"github.com/sourcegraph/go-langserver/langserver/internal/godef/go/types"
)

type Result struct {
	// Start and end positions of the definition (only if not an import statement).
	Start, End token.Pos

	// Package in question, only present if an import statement OR package selector
	// ('http' in 'http.Router').
	Package *build.Package
}

func Godef(fset *token.FileSet, offset int, filename string, src []byte) (*Result, error) {
	pkgScope := ast.NewScope(parser.Universe)
	f, err := parser.ParseFile(fset, filename, src, 0, pkgScope, types.DefaultImportPathToName)
	if f == nil {
		return nil, fmt.Errorf("cannot parse %s: %v", filename, err)
	}

	o := findIdentifier(fset, f, offset)
	if o == nil {
		return nil, fmt.Errorf("no identifier found")
	}
	switch e := o.(type) {
	case *ast.ImportSpec:
		path, err := importPath(e)
		if err != nil {
			return nil, err
		}
		pkg, err := build.Default.Import(path, filepath.Dir(filename), build.FindOnly)
		if err != nil {
			return nil, fmt.Errorf("error finding import path for %s: %s", path, err)
		}
		return &Result{Package: pkg}, nil
	case ast.Expr:
		result := func(obj *ast.Object) (*Result, error) {
			p := types.DeclPos(obj)
			r := &Result{Start: p, End: p + token.Pos(len(obj.Name))}
			if imp, ok := obj.Decl.(*ast.ImportSpec); ok {
				path, err := importPath(imp)
				if err != nil {
					return nil, err
				}
				pkg, err := build.Default.Import(path, filepath.Dir(fset.Position(p).Filename), build.FindOnly)
				if err != nil {
					return nil, fmt.Errorf("error finding import path for %s: %s", path, err)
				}
				r.Package = pkg
			}
			return r, nil
		}
		importer := types.DefaultImporter(fset)
		// try local declarations only
		if obj, _ := types.ExprType(e, importer, fset); obj != nil {
			return result(obj)
		}

		// add declarations from other files in the local package and try again
		pkg, err := parseLocalPackage(fset, filename, f, pkgScope, types.DefaultImportPathToName)
		if pkg == nil {
			log.Printf("parseLocalPackage error: %v\n", err)
		}
		if obj, _ := types.ExprType(e, importer, fset); obj != nil {
			return result(obj)
		}
		return nil, fmt.Errorf("no declaration found for %v", pretty{fset, e})
	}
	return nil, fmt.Errorf("unreached")
}

func importPath(n *ast.ImportSpec) (string, error) {
	p, err := strconv.Unquote(n.Path.Value)
	if err != nil {
		return "", fmt.Errorf("invalid string literal %q in ast.ImportSpec", n.Path.Value)
	}
	return p, nil
}

// findIdentifier looks for an identifier at byte-offset searchpos
// inside the parsed source represented by node.
// If it is part of a selector expression, it returns
// that expression rather than the identifier itself.
//
// As a special case, if it finds an import
// spec, it returns ImportSpec.
//
func findIdentifier(fset *token.FileSet, f *ast.File, searchpos int) ast.Node {
	ec := make(chan ast.Node)
	found := func(startPos, endPos token.Pos) bool {
		start := fset.Position(startPos).Offset
		end := start + int(endPos-startPos)
		return start <= searchpos && searchpos <= end
	}
	go func() {
		var visit func(ast.Node) bool
		visit = func(n ast.Node) bool {
			var startPos token.Pos
			switch n := n.(type) {
			default:
				return true
			case *ast.Ident:
				startPos = n.NamePos
			case *ast.SelectorExpr:
				startPos = n.Sel.NamePos
			case *ast.ImportSpec:
				startPos = n.Pos()
			case *ast.StructType:
				// If we find an anonymous bare field in a
				// struct type, its definition points to itself,
				// but we actually want to go elsewhere,
				// so assume (dubiously) that the expression
				// works globally and return a new node for it.
				for _, field := range n.Fields.List {
					if field.Names != nil {
						continue
					}
					t := field.Type
					if pt, ok := field.Type.(*ast.StarExpr); ok {
						t = pt.X
					}
					if id, ok := t.(*ast.Ident); ok {
						if found(id.NamePos, id.End()) {
							e, err := parseExpr(fset, f.Scope, id.Name)
							if err != nil {
								log.Println(err) // TODO(slimsag): return to caller
							}
							ec <- e
							runtime.Goexit()
						}
					}
				}
				return true
			}
			if found(startPos, n.End()) {
				ec <- n
				runtime.Goexit()
			}
			return true
		}
		ast.Walk(visitorFunc(visit), f)
		ec <- nil
	}()
	return <-ec
}

func parseExpr(fset *token.FileSet, s *ast.Scope, expr string) (ast.Expr, error) {
	n, err := parser.ParseExpr(fset, "<arg>", expr, s, types.DefaultImportPathToName)
	if err != nil {
		return nil, fmt.Errorf("cannot parse expression: %v", err)
	}
	switch n := n.(type) {
	case *ast.Ident, *ast.SelectorExpr:
		return n, nil
	}
	return nil, fmt.Errorf("no identifier found in expression")
}

type visitorFunc func(n ast.Node) bool

func (f visitorFunc) Visit(n ast.Node) ast.Visitor {
	if f(n) {
		return f
	}
	return nil
}

var errNoPkgFiles = errors.New("no more package files found")

// parseLocalPackage reads and parses all go files from the
// current directory that implement the same package name
// the principal source file, except the original source file
// itself, which will already have been parsed.
//
func parseLocalPackage(fset *token.FileSet, filename string, src *ast.File, pkgScope *ast.Scope, pathToName parser.ImportPathToName) (*ast.Package, error) {
	pkg := &ast.Package{src.Name.Name, pkgScope, nil, map[string]*ast.File{filename: src}}
	d, f := filepath.Split(filename)
	if d == "" {
		d = "./"
	}
	fd, err := os.Open(d)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	list, err := fd.Readdirnames(-1)
	if err != nil {
		return nil, errNoPkgFiles
	}

	for _, pf := range list {
		file := filepath.Join(d, pf)
		if !strings.HasSuffix(pf, ".go") ||
			pf == f ||
			pkgName(fset, file) != pkg.Name {
			continue
		}
		src, err := parser.ParseFile(fset, file, nil, 0, pkg.Scope, types.DefaultImportPathToName)
		if err == nil {
			pkg.Files[file] = src
		}
	}
	if len(pkg.Files) == 1 {
		return nil, errNoPkgFiles
	}
	return pkg, nil
}

// pkgName returns the package name implemented by the
// go source filename.
//
func pkgName(fset *token.FileSet, filename string) string {
	prog, _ := parser.ParseFile(fset, filename, nil, parser.PackageClauseOnly, nil, types.DefaultImportPathToName)
	if prog != nil {
		return prog.Name.Name
	}
	return ""
}

type pretty struct {
	fset *token.FileSet
	n    interface{}
}

func (p pretty) String() string {
	var b bytes.Buffer
	printer.Fprint(&b, p.fset, p.n)
	return b.String()
}
