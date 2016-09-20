package golang

import (
	"context"
	"fmt"
	"go/build"
	"go/doc"
	"go/parser"
	"go/token"
	"log"
	"os"
	"os/exec"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/jsonrpc2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/lsp"
)

func (h *Handler) handleSymbol(ctx context.Context, req *jsonrpc2.Request, params lsp.WorkspaceSymbolParams) ([]lsp.SymbolInformation, error) {
	pkgs, err := expandPackages(ctx, h.goEnv(), []string{params.Query})
	if err != nil {
		return nil, err
	}

	var symbols []lsp.SymbolInformation
	var failed int
	emit := func(name, container string, kind lsp.SymbolKind, fs *token.FileSet, pos token.Pos) {
		if pos == 0 {
			symbols = append(symbols, lsp.SymbolInformation{
				Name: name,
				Kind: kind,
				// TODO Location.URI
				ContainerName: container,
			})
			return
		}
		start := fs.Position(pos)
		end := fs.Position(pos + token.Pos(len(name)) - 1)
		uri, err := h.fileURI(start.Filename)
		if err != nil {
			failed++
			return
		}
		symbols = append(symbols, lsp.SymbolInformation{
			Name: name,
			Kind: kind,
			Location: lsp.Location{
				URI: uri,
				Range: lsp.Range{
					Start: lsp.Position{Line: start.Line - 1, Character: start.Column - 1},
					End:   lsp.Position{Line: end.Line - 1, Character: end.Column - 1},
				},
			},
			ContainerName: container,
		})
	}
	buildCtx := build.Default
	buildCtx.GOPATH = h.filePath("gopath")
	buildCtx.CgoEnabled = false
	for _, pkg := range pkgs {
		err := symbolDo(buildCtx, pkg, emit)
		if err != nil {
			return nil, err
		}
	}
	if failed > 0 {
		log.Printf("WARNING: failed to create %d symbols", failed)
	}
	return symbols, nil
}

type emitFunc func(name, container string, kind lsp.SymbolKind, fs *token.FileSet, pos token.Pos)

func symbolDo(buildCtx build.Context, pkgPath string, emit emitFunc) error {
	// Package must be importable.
	bpkg, err := buildCtx.Import(pkgPath, "", 0)
	if err != nil {
		return err
	}
	pkg, err := parsePackage(bpkg)
	if err != nil {
		return err
	}

	// TODO
	// * go/doc doesn't parse out Fields of structs
	// * v.Decl.TokPos is not correct
	emit(pkg.build.ImportPath, "", lsp.SKPackage, pkg.fs, 0)
	for _, t := range pkg.doc.Types {
		for _, v := range t.Funcs {
			emit(v.Name, pkg.build.ImportPath, lsp.SKFunction, pkg.fs, v.Decl.Name.NamePos)
		}
		for _, v := range t.Methods {
			emit(v.Name, pkg.build.ImportPath+"."+t.Name, lsp.SKMethod, pkg.fs, v.Decl.Name.NamePos)
		}
		for _, v := range t.Consts {
			for _, name := range v.Names {
				emit(name, pkg.build.ImportPath, lsp.SKConstant, pkg.fs, v.Decl.TokPos)
			}
		}
		for _, v := range t.Vars {
			for _, name := range v.Names {
				emit(name, pkg.build.ImportPath, lsp.SKVariable, pkg.fs, v.Decl.TokPos)
			}
		}
	}
	for _, v := range pkg.doc.Consts {
		for _, name := range v.Names {
			emit(name, pkg.build.ImportPath, lsp.SKConstant, pkg.fs, v.Decl.TokPos)
		}
	}
	for _, v := range pkg.doc.Vars {
		for _, name := range v.Names {
			emit(name, pkg.build.ImportPath, lsp.SKVariable, pkg.fs, v.Decl.TokPos)
		}
	}
	for _, v := range pkg.doc.Funcs {
		emit(v.Name, pkg.build.ImportPath, lsp.SKFunction, pkg.fs, v.Decl.Name.NamePos)
	}

	return nil
}

type parsedPackage struct {
	name  string // Package name, json for encoding/json.
	doc   *doc.Package
	build *build.Package
	fs    *token.FileSet // Needed for printing.
}

// parsePackage turns the build package we found into a parsed package
// we can then use to generate documentation.
func parsePackage(pkg *build.Package) (*parsedPackage, error) {
	fs := token.NewFileSet()
	// include tells parser.ParseDir which files to include.
	// That means the file must be in the build package's GoFiles or CgoFiles
	// list only (no tag-ignored files, tests, swig or other non-Go files).
	include := func(info os.FileInfo) bool {
		for _, name := range pkg.GoFiles {
			if name == info.Name() {
				return true
			}
		}
		for _, name := range pkg.CgoFiles {
			if name == info.Name() {
				return true
			}
		}
		return false
	}
	pkgs, err := parser.ParseDir(fs, pkg.Dir, include, 0)
	if err != nil {
		return nil, err
	}
	astPkg, ok := pkgs[pkg.Name]
	if !ok {
		// Should not happen, but protect against it just in case
		return nil, fmt.Errorf("could not find pkg %s in %s", pkg.Name, pkg.Dir)
	}

	docPkg := doc.New(astPkg, pkg.ImportPath, doc.AllDecls)

	return &parsedPackage{
		name:  pkg.Name,
		doc:   docPkg,
		build: pkg,
		fs:    fs,
	}, nil
}

func expandPackages(ctx context.Context, env, pkgs []string) ([]string, error) {
	if len(pkgs) == 1 && pkgs[0] == "github.com/golang/go/..." {
		b, err := cmdOutput(ctx, env, exec.Command("go", "list", "-e", "-f", "{{if .Standard}}{{.ImportPath}}{{end}}", "..."))
		if err != nil {
			return nil, err
		}
		return strings.Fields(string(b)), nil
	}
	args := append([]string{"list"}, pkgs...)
	b, err := cmdOutput(ctx, env, exec.Command("go", args...))
	if err != nil {
		return nil, err
	}
	return strings.Fields(string(b)), nil
}
