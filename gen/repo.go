package gen

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/printer"
	"go/token"
	"path/filepath"
	"strings"
)

var fset = token.NewFileSet()

// RepoURIExpr returns the AST expression that evaluates to the repo
// URI, given a protobuf RPC service method argument (XxxOp,
// RepoRevSpec, RepoSpec, BuildSpec, etc.).
//
// For example, if arg="x" and argType="*sourcegraph.BuildsCreateOp",
// then RepoURIExpr returns an AST expression equivalent to
// "x.RepoRev.URI".
func RepoURIExpr(arg ast.Expr, argType ast.Expr) ast.Expr {
	if x := AstString(argType); x == "*sourcegraph.RepoSpec" || x == "sourcegraph.RepoSpec" || x == "RepoSpec" {
		return &ast.SelectorExpr{X: arg, Sel: ast.NewIdent("URI")}
	}
	if x := AstString(argType); x == "*sourcegraph.DefSpec" || x == "sourcegraph.DefSpec" || x == "DefSpec" {
		return &ast.SelectorExpr{X: arg, Sel: ast.NewIdent("Repo")}
	}

	switch t := argType.(type) {
	case *ast.StarExpr:
		return RepoURIExpr(arg, t.X)
	case *ast.Ident:
		if ast.IsExported(t.Name) {
			_, spec, err := typeSpec("sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph", t.Name)
			if err != nil {
				return nil
			}
			structType, ok := spec.Type.(*ast.StructType)
			if !ok {
				return nil
			}
			for _, field := range structType.Fields.List {
				var argField ast.Expr
				if len(field.Names) > 0 {
					argField = &ast.SelectorExpr{X: arg, Sel: field.Names[0]}
				} else {
					argField = arg
				}
				x := RepoURIExpr(argField, field.Type)
				if x != nil {
					return x
				}
			}
		}
	case *ast.SelectorExpr:
		if id, ok := t.X.(*ast.Ident); ok && id.Name == "sourcegraph" {
			return RepoURIExpr(arg, t.Sel)
		}
	}
	return nil
}

func typeSpec(path string, name string) (*build.Package, *ast.TypeSpec, error) {
	pkg, err := build.Import(path, "", 0)
	if err != nil {
		return nil, nil, err
	}
	for _, file := range pkg.GoFiles {
		f, err := parser.ParseFile(fset, filepath.Join(pkg.Dir, file), nil, 0)
		if err != nil {
			continue
		}
		for _, decl := range f.Decls {
			decl, ok := decl.(*ast.GenDecl)
			if !ok || decl.Tok != token.TYPE {
				continue
			}
			for _, spec := range decl.Specs {
				spec := spec.(*ast.TypeSpec)
				if spec.Name.Name != name {
					continue
				}
				return pkg, spec, nil
			}
		}
	}
	return nil, nil, fmt.Errorf("type %s not found in %s", name, path)
}

func AstString(x ast.Node) string {
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, x); err != nil {
		panic(err)
	}
	return strings.Replace(strings.Replace(buf.String(), "\n", "", -1), "\t", "", -1)
}
