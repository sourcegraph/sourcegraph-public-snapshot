package golang

import (
	"context"
	"errors"
	"go/ast"
	"go/types"
	"sort"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/jsonrpc2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/lsp"
)

func (h *LangHandler) handleReferences(ctx context.Context, conn jsonrpc2Conn, req *jsonrpc2.Request, params lsp.ReferenceParams) ([]lsp.Location, error) {
	fset, node, pkg, err := h.typecheck(ctx, conn, params.TextDocument.URI, params.Position)
	if err != nil {
		return nil, err
	}

	obj, ok := pkg.Uses[node]
	if !ok {
		obj, ok = pkg.Defs[node]
	}
	if !ok {
		return nil, errors.New("references object not found")
	}

	var nodes []*ast.Ident
	if params.Context.IncludeDeclaration {
		nodes = append(nodes, &ast.Ident{NamePos: obj.Pos(), Name: obj.Name()})
	}
	for node, o := range pkg.Info.Uses {
		if sameObj(obj, o) {
			nodes = append(nodes, node)
		}
	}

	// TODO(sqs): I think the AfterTypeCheck clearing of data
	// structures limits the references data to only those in the same
	// file.

	locs := goRangesToLSPLocations(fset, nodes)
	sort.Sort(locationList(locs))
	return locs, nil
}

// same reports whether x and y are identical, or both are PkgNames
// that import the same Package.
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
