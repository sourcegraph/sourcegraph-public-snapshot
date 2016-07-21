package golang

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"

	"golang.org/x/tools/go/ast/astutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/jsonrpc2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/lsp"
)

func (h *Session) handleDefinition(req *jsonrpc2.Request, params lsp.TextDocumentPositionParams) ([]lsp.Location, error) {
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

	// Did the parser resolve it to a local object?
	obj := node.Obj
	if obj == nil || !obj.Pos().IsValid() {
		return nil, fmt.Errorf("Could not resolve to local object")
	}

	// TODO(keegancsmith) the end value here is a hack
	objNode := fakeNode{obj.Pos(), obj.Pos() + token.Pos(len(obj.Name))}

	var locs []lsp.Location
	locs = append(locs, lsp.Location{
		URI:   params.TextDocument.URI,
		Range: rangeForNode(h.fset, objNode),
	})
	return locs, nil
}

type fakeNode struct{ p, e token.Pos }

func (n fakeNode) Pos() token.Pos { return n.p }
func (n fakeNode) End() token.Pos { return n.e }
