package golang

import (
	"context"
	"fmt"
	"go/types"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/jsonrpc2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/lsp"
)

func (h *LangHandler) handleHover(ctx context.Context, conn jsonrpc2Conn, req *jsonrpc2.Request, params lsp.TextDocumentPositionParams) (*lsp.Hover, error) {
	fset, node, pkg, err := h.typecheck(ctx, conn, params.TextDocument.URI, params.Position)
	if err != nil {
		return nil, err
	}

	o := pkg.ObjectOf(node)
	t := pkg.TypeOf(node)
	if o == nil && t == nil {
		return nil, fmt.Errorf("type/object not found at %+v", params.Position)
	}

	// Don't package-qualify the string output.
	qf := func(*types.Package) string { return "" }

	var s string
	if o != nil {
		s = types.ObjectString(o, qf)
	} else if t != nil {
		s = types.TypeString(t, qf)
	}
	if strings.HasPrefix(s, "field ") && t != nil {
		s += ": " + t.String()
	}

	return &lsp.Hover{
		Contents: []lsp.MarkedString{{Language: "go", Value: s}},
		Range:    rangeForNode(fset, node),
	}, nil
}
