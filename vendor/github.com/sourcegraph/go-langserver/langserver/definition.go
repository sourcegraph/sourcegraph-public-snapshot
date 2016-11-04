package langserver

import (
	"context"
	"errors"
	"go/ast"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

func (h *LangHandler) handleDefinition(ctx context.Context, conn JSONRPC2Conn, req *jsonrpc2.Request, params lsp.TextDocumentPositionParams) ([]lsp.Location, error) {
	fset, node, _, pkg, err := h.typecheck(ctx, conn, params.TextDocument.URI, params.Position)
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
	locs := goRangesToLSPLocations(fset, nodes)
	for i := range locs {
		// LSP expects a range to be of the entire body, not just of the
		// identifier, so we pretend its just a position and not a range.
		locs[i].Range.End = locs[i].Range.Start
	}
	return locs, nil
}
