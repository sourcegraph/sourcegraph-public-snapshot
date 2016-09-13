package golang

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"

	"golang.org/x/tools/go/ast/astutil"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/lsp"
)

func offsetForPosition(contents []byte, p lsp.Position) (offset uint64, valid bool) {
	line := 0
	col := 0
	// TODO(sqs): count chars, not bytes, per LSP. does that mean we
	// need to maintain 2 separate counters since we still need to
	// return the offset as bytes?
	for _, b := range contents {
		if line == p.Line && col == p.Character {
			return offset, true
		}
		if line > p.Line || (line == p.Line && col > p.Character) {
			return 0, false
		}

		offset++
		if b == '\n' {
			line++
			col = 0
		} else {
			col++
		}
	}
	return 0, false
}

func rangeForNode(fset *token.FileSet, node ast.Node) lsp.Range {
	start := fset.Position(node.Pos())
	end := fset.Position(node.End() - 1) // node.End is exclusive, but we want inclusive
	return lsp.Range{
		Start: lsp.Position{Line: start.Line - 1, Character: start.Column - 1},
		End:   lsp.Position{Line: end.Line - 1, Character: end.Column - 1},
	}
}

func rangeAtPosition(p lsp.Position, contents []byte) (lsp.Range, error) {
	var r lsp.Range
	ofs, valid := offsetForPosition(contents, p)
	if !valid {
		return r, errors.New("invalid start position for def")
	}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "a.go", contents, 0)
	if err != nil {
		return r, err
	}
	pos := fset.File(f.Pos()).Pos(int(ofs))
	nodes, _ := astutil.PathEnclosingInterval(f, pos, pos)
	if len(nodes) == 0 {
		return r, errors.New("no nodes found at def")
	}
	return rangeForNode(fset, nodes[0]), nil
}

func docAtPosition(p lsp.Position, contents []byte) (string, error) {
	ofs, valid := offsetForPosition(contents, p)
	if !valid {
		return "", errors.New("invalid start position for def")
	}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "a.go", contents, parser.ParseComments)
	if err != nil {
		return "", err
	}
	pos := fset.File(f.Pos()).Pos(int(ofs))
	nodes, _ := astutil.PathEnclosingInterval(f, pos, pos)
	if len(nodes) == 0 {
		return "", errors.New("no nodes found at def")
	}

	var doc *ast.CommentGroup
	for _, node := range nodes {
		switch n := node.(type) {
		case *ast.Field:
			doc = n.Doc
		case *ast.FuncDecl:
			doc = n.Doc
		case *ast.GenDecl:
			doc = n.Doc
		case *ast.ImportSpec:
			doc = n.Doc
		case *ast.TypeSpec:
			doc = n.Doc
		case *ast.ValueSpec:
			doc = n.Doc
		}
		if doc != nil {
			break
		}
	}
	if doc == nil {
		return "", nil
	}
	return doc.Text(), nil
}
