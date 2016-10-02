package golang

import (
	"context"
	"errors"
	"go/token"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/jsonrpc2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/lsp"
)

func (h *LangHandler) handleDefinition(ctx context.Context, conn jsonrpc2Conn, req *jsonrpc2.Request, params lsp.TextDocumentPositionParams) ([]lsp.Location, error) {
	fset, node, pkg, err := h.typecheck(ctx, conn, params.TextDocument.URI, params.Position)
	if err != nil {
		// Invalid nodes means we tried to click on something which is
		// not an ident (eg comment/string/etc). Return no locations.
		if _, ok := err.(*invalidNodeError); ok {
			return nil, nil
		}
		return nil, err
	}

	var nodes []posEnd
	obj, ok := pkg.Uses[node]
	if !ok {
		obj, ok = pkg.Defs[node]
	}
	if ok && obj != nil {
		nodes = append(nodes, fakeNode{obj.Pos(), obj.Pos() + token.Pos(len(obj.Name()))})
	}
	if len(nodes) == 0 {
		return nil, errors.New("definition not found")
	}
	return goRangesToLSPLocations(fset, nodes), nil
}
