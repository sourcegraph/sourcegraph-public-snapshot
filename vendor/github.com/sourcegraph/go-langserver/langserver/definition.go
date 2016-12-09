package langserver

import (
	"context"
	"errors"
	"go/ast"
	"log"

	"github.com/sourcegraph/go-langserver/langserver/internal/refs"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/go-langserver/pkg/lspext"
	"github.com/sourcegraph/jsonrpc2"
)

func (h *LangHandler) handleDefinition(ctx context.Context, conn JSONRPC2Conn, req *jsonrpc2.Request, params lsp.TextDocumentPositionParams) ([]lsp.Location, error) {
	res, err := h.handleXDefinition(ctx, conn, req, params)
	if err != nil {
		return nil, err
	}
	locs := make([]lsp.Location, 0, len(res))
	for _, li := range res {
		locs = append(locs, li.Location)
	}
	return locs, nil
}

func (h *LangHandler) handleXDefinition(ctx context.Context, conn JSONRPC2Conn, req *jsonrpc2.Request, params lsp.TextDocumentPositionParams) ([]lspext.LocationInformation, error) {
	rootPath := h.FilePath(h.init.RootPath)
	bctx := h.OverlayBuildContext(ctx, h.defaultBuildContext(), !h.init.NoOSFileSystemAccess)

	fset, node, pathEnclosingInterval, _, pkg, err := h.typecheck(ctx, conn, params.TextDocument.URI, params.Position)
	if err != nil {
		// Invalid nodes means we tried to click on something which is
		// not an ident (eg comment/string/etc). Return no locations.
		if _, ok := err.(*invalidNodeError); ok {
			return nil, nil
		}
		return nil, err
	}

	var nodes []*ast.Ident
	obj, ok := pkg.Uses[node]
	if !ok {
		obj, ok = pkg.Defs[node]
	}
	if ok && obj != nil {
		if p := obj.Pos(); p.IsValid() {
			nodes = append(nodes, &ast.Ident{NamePos: p, Name: obj.Name()})
		} else {
			// Builtins have an invalid Pos. Just don't emit a definition for
			// them, for now. It's not that valuable to jump to their def.
			//
			// TODO(sqs): find a way to actually emit builtin locations
			// (pointing to builtin/builtin.go).
			return nil, nil
		}
	}
	if len(nodes) == 0 {
		return nil, errors.New("definition not found")
	}
	locs := make([]lspext.LocationInformation, 0, len(nodes))
	for _, node := range nodes {
		// Determine location information for the node.
		l := lspext.LocationInformation{
			Location: goRangeToLSPLocation(fset, node.Pos(), node.End()),
		}
		// LSP expects a range to be of the entire body, not just of the
		// identifier, so we pretend its just a position and not a range.
		l.Location.Range.End = l.Location.Range.Start

		// Determine metadata information for the node.

		if def, err := refs.DefInfo(pkg.Pkg, &pkg.Info, pathEnclosingInterval, node.Pos()); err == nil {
			symDesc, err := defSymbolDescriptor(bctx, rootPath, *def)
			if err != nil {
				// TODO: tracing
				log.Println("refs.DefInfo:", err)
			} else {
				l.Symbol = *symDesc
			}
		} else {
			// TODO: tracing
			log.Println("refs.DefInfo:", err)
		}
		locs = append(locs, l)
	}
	return locs, nil
}
