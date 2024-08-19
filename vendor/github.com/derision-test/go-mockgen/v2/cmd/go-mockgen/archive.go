package main

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"go/types"
	"path/filepath"
	"strings"

	"github.com/derision-test/go-mockgen/v2/internal"
)

type archive struct {
	// ImportMap refers to the actual import path to the library this archive represents.
	// See https://github.com/bazelbuild/rules_go/blob/a9b312afd2866f4316356b456df1971bff6cd244/go/core.rst#go_library.
	ImportMap string
	File      string
}

// The following is the format expected by this function:
//
//	IMPORTMAP=EXPORT e.g. github.com/foo/bar=bar_export.a
//
// The flag is structured in this format to loosely follow https://sourcegraph.com/github.com/bazelbuild/rules_go@a9b312afd2866f4316356b456df1971bff6cd244/-/blob/go/private/actions/compilepkg.bzl?L22-29;
// however, the IMPORTPATHS section is omitted. There may be future
// work involved in resolving import aliases/vendoring using IMPORTPATHS.
func parseArchive(a string) (archive, error) {
	args := strings.Split(a, "=")
	if len(args) != 2 {
		return archive{}, fmt.Errorf("expected 2 elements, got %d: %v", len(args), a)
	}

	return archive{
		ImportMap: args[0],
		File:      args[1],
	}, nil
}

func PackagesArchive(p loadParams) (packages []*internal.GoPackage, err error) {
	fset := token.NewFileSet()
	for _, importpath := range p.importPaths {
		files := make([]*ast.File, 0, len(p.sources))
		for _, src := range p.sources[importpath] {
			if ok, err := build.Default.MatchFile(filepath.Dir(src), filepath.Base(src)); err != nil {
				return nil, fmt.Errorf("error checking if file matches constraints: %w", err)
			} else if !ok || filepath.Ext(src) == ".s" {
				fmt.Printf("skipping %q\n", src)
				continue
			}

			f, err := parser.ParseFile(fset, src, nil, parser.ParseComments)
			if err != nil {
				return nil, fmt.Errorf("error parsing %q: %v", src, err)
			}

			files = append(files, f)
		}

		imp, err := newImporter(fset, p.archives, p.stdlibRoot)
		if err != nil {
			return nil, err
		}
		conf := types.Config{Importer: imp, Error: func(err error) {
			fmt.Println(err)
		}}
		typesInfo := &types.Info{
			Types:      make(map[ast.Expr]types.TypeAndValue),
			Defs:       make(map[*ast.Ident]types.Object),
			Uses:       make(map[*ast.Ident]types.Object),
			Implicits:  make(map[ast.Node]types.Object),
			Selections: make(map[*ast.SelectorExpr]*types.Selection),
			Scopes:     make(map[ast.Node]*types.Scope),
		}

		pkg, err := conf.Check(importpath, fset, files, typesInfo)
		if err != nil {
			return nil, fmt.Errorf("error building pkg %q: %w", importpath, err)
		}
		packages = append(packages, &internal.GoPackage{
			PkgPath:         pkg.Path(),
			CompiledGoFiles: p.sources[importpath],
			Syntax:          files,
			Types:           pkg,
			TypesInfo:       typesInfo,
		})
	}
	return
}
