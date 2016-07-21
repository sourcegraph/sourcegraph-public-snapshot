package golang

import (
	"errors"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"go/types"
	"log"
	"sort"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/buildutil"
	"golang.org/x/tools/go/loader"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/jsonrpc2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/lsp"
)

func (h *Session) handleReferences(req *jsonrpc2.Request, params lsp.ReferenceParams) ([]lsp.Location, error) {
	buildCtx := buildutil.OverlayContext(&build.Default, h.overlayFiles)

	var importPath string
	bpkg, _ := buildutil.ContainingPackage(buildCtx, h.init.RootPath, params.TextDocument.URI)
	if bpkg != nil {
		importPath = bpkg.ImportPath
	}

	contents, err := h.readFile(params.TextDocument.URI)
	if err != nil {
		return nil, err
	}

	if h.fset == nil {
		h.fset = token.NewFileSet()
	}
	f, err := parser.ParseFile(h.fset, h.filePath(params.TextDocument.URI), contents, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	ofs, valid := offsetForPosition(contents, params.Position)
	if !valid {
		return nil, errors.New("invalid position")
	}

	pos := h.fset.File(f.Pos()).Pos(int(ofs))
	p := h.fset.Position(pos)
	loc := fmt.Sprintf("%s:%d:%d", p.Filename, p.Line, p.Column)

	// Fast-path to short-circuit when we're not even over an ident
	// node, and avoid doing a full typecheck in that case.
	nodes, _ := astutil.PathEnclosingInterval(f, pos, pos)
	if len(nodes) == 0 {
		return nil, errors.New("no nodes found at cursor")
	}
	node, ok := nodes[0].(*ast.Ident)
	if !ok {
		return nil, fmt.Errorf("node is %T, not ident, at %s", nodes[0], loc)
	}

	conf := loader.Config{
		Fset: h.fset,
		TypeChecker: types.Config{
			DisableUnusedImportCheck: true,
			FakeImportC:              true,
			Error:                    func(err error) {},
		},
		Build:       buildCtx,
		Cwd:         h.init.RootPath,
		AllowErrors: true,
		// TODO(sqs): investigate using AfterTypeCheck for better perf
	}
	conf.CreateFromFiles(importPath, f)
	prog, err := conf.Load()
	if err != nil {
		log.Printf("typechecking %s: %s", params.TextDocument.URI, err)
		if prog == nil {
			return nil, err
		}
	}

	pkg := prog.InitialPackages()[0]
	if pkg == nil {
		return nil, errors.New("no package found")
	}
	o := pkg.ObjectOf(node)

	var locs []lsp.Location
	if params.Context.IncludeDeclaration {
		locs = append(locs, lsp.Location{
			URI:   params.TextDocument.URI,
			Range: rangeForNode(h.fset, node),
		})
	}
	for astIdent, obj := range pkg.Info.Uses {
		if sameObj(o, obj) {
			locs = append(locs, lsp.Location{
				URI:   params.TextDocument.URI,
				Range: rangeForNode(h.fset, astIdent),
			})
		}
	}

	sort.Sort(locationList(locs))
	return locs, nil
}

// same reports whether x and y are identical, or both are PkgNames
// that import the same Package.
//
func sameObj(x, y types.Object) bool {
	if x == y {
		return true
	}
	if x, ok := x.(*types.PkgName); ok {
		if y, ok := y.(*types.PkgName); ok {
			return x.Imported() == y.Imported()
		}
	}
	return false
}

type locationList []lsp.Location

func (l locationList) Less(a, b int) bool {
	if l[a].URI != l[b].URI {
		return l[a].URI < l[b].URI
	}
	if l[a].Range.Start.Line != l[b].Range.Start.Line {
		return l[a].Range.Start.Line < l[b].Range.Start.Line
	}
	return l[a].Range.Start.Character < l[b].Range.Start.Character
}

func (l locationList) Swap(a, b int) {
	l[a], l[b] = l[b], l[a]
}
func (l locationList) Len() int {
	return len(l)
}
