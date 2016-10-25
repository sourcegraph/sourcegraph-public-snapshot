package langserver

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
)

func offsetForPosition(contents []byte, p lsp.Position) (offset int, valid bool, whyInvalid string) {
	line := 0
	col := 0
	// TODO(sqs): count chars, not bytes, per LSP. does that mean we
	// need to maintain 2 separate counters since we still need to
	// return the offset as bytes?
	for _, b := range contents {
		if line == p.Line && col == p.Character {
			return offset, true, ""
		}
		if (line == p.Line && col > p.Character) || line > p.Line {
			return 0, false, fmt.Sprintf("character %d is beyond line %d boundary", p.Character, p.Line)
		}
		offset++
		if b == '\n' {
			line++
			col = 0
		} else {
			col++
		}
	}
	if line == 0 {
		return 0, false, fmt.Sprintf("character %d is beyond first line boundary", p.Character)
	}
	return 0, false, fmt.Sprintf("file only has %d lines", line+1)
}

func rangeForNode(fset *token.FileSet, node ast.Node) lsp.Range {
	start := fset.Position(node.Pos())
	end := fset.Position(node.End()) // node.End is exclusive, but we want inclusive
	return lsp.Range{
		Start: lsp.Position{Line: start.Line - 1, Character: start.Column - 1},
		End:   lsp.Position{Line: end.Line - 1, Character: end.Column - 1},
	}
}

type fakeNode struct{ p, e token.Pos }

func (n fakeNode) Pos() token.Pos { return n.p }
func (n fakeNode) End() token.Pos { return n.e }

func goRangesToLSPLocations(fset *token.FileSet, nodes []*ast.Ident) []lsp.Location {
	locs := make([]lsp.Location, len(nodes))
	for i, node := range nodes {
		p := fset.Position(node.Pos())
		locs[i] = lsp.Location{
			URI:   "file://" + p.Filename,
			Range: rangeForNode(fset, node),
		}
	}
	return locs
}

func goRangeToLSPLocation(fset *token.FileSet, pos token.Pos, end token.Pos) lsp.Location {
	return lsp.Location{
		URI:   "file://" + fset.Position(pos).Filename,
		Range: rangeForNode(fset, fakeNode{p: pos, e: end}),
	}

}
