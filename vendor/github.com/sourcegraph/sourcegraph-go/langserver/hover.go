package langserver

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	"github.com/sourcegraph/jsonrpc2"
	"github.com/sourcegraph/sourcegraph-go/pkg/lsp"
)

func (h *LangHandler) handleHover(ctx context.Context, conn JSONRPC2Conn, req *jsonrpc2.Request, params lsp.TextDocumentPositionParams) (*lsp.Hover, error) {
	fset, node, pkg, err := h.typecheck(ctx, conn, params.TextDocument.URI, params.Position)
	if err != nil {
		// Invalid nodes means we tried to click on something which is
		// not an ident (eg comment/string/etc). Return no information.
		if _, ok := err.(*invalidNodeError); ok {
			return nil, nil
		}
		return nil, err
	}

	o := pkg.ObjectOf(node)
	t := pkg.TypeOf(node)
	if o == nil && t == nil {
		// Package statement idents don't have an object, so try that separately.
		if pkgName := packageStatementName(fset, pkg.Files, node); pkgName != "" {
			return &lsp.Hover{
				Contents: []lsp.MarkedString{{Language: "go", Value: "package " + pkgName}},
				Range:    rangeForNode(fset, node),
			}, nil
		}

		return nil, fmt.Errorf("type/object not found at %+v", params.Position)
	}

	// Don't package-qualify the string output.
	qf := func(*types.Package) string { return "" }

	var s string
	if f, ok := o.(*types.Var); ok && f.IsField() {
		// TODO(sqs): make this be like (T).F not "struct field F string".
		s = "struct " + o.String()
	} else if o != nil {
		if obj, ok := o.(*types.TypeName); ok {
			typ := obj.Type().Underlying()
			if _, ok := typ.(*types.Struct); ok {
				s = "type " + obj.Name() + " struct"
			}
		}
		if s == "" {
			s = types.ObjectString(o, qf)
		}
	} else if t != nil {
		s = types.TypeString(t, qf)
	}

	return &lsp.Hover{
		Contents: []lsp.MarkedString{{Language: "go", Value: s}},
		Range:    rangeForNode(fset, node),
	}, nil
}

// packageStatementName returns the package name ((*ast.Ident).Name)
// of node iff node is the package statement of a file ("package p").
func packageStatementName(fset *token.FileSet, files []*ast.File, node *ast.Ident) string {
	for _, f := range files {
		if f.Name == node {
			return node.Name
		}
	}
	return ""
}
