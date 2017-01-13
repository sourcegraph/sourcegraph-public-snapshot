package langserver

import (
	"context"
	"go/ast"
	"go/token"
	"go/types"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

func (h *LangHandler) handleTextDocumentSignatureHelp(ctx context.Context, conn JSONRPC2Conn, req *jsonrpc2.Request, params lsp.TextDocumentPositionParams) (*lsp.SignatureHelp, error) {
	fset, _, nodes, _, pkg, err := h.typecheck(ctx, conn, params.TextDocument.URI, params.Position)
	if err != nil {
		if _, ok := err.(*invalidNodeError); !ok {
			return nil, err
		}
	}

	call := callExpr(fset, nodes)
	if call == nil {
		return nil, nil
	}
	t := pkg.TypeOf(call.Fun)
	signature, ok := t.(*types.Signature)
	if !ok {
		return nil, nil
	}
	info := lsp.SignatureInformation{Label: shortType(signature)}
	sParams := signature.Params()
	info.Parameters = make([]lsp.ParameterInformation, sParams.Len())
	for i := 0; i < sParams.Len(); i++ {
		info.Parameters[i] = lsp.ParameterInformation{Label: shortParam(sParams.At(i))}
	}
	activeParameter := len(info.Parameters)
	if activeParameter > 0 {
		activeParameter = activeParameter - 1
	}
	numArguments := len(call.Args)
	if activeParameter > numArguments {
		activeParameter = numArguments
	}

	return &lsp.SignatureHelp{Signatures: []lsp.SignatureInformation{info}, ActiveSignature: 0, ActiveParameter: activeParameter}, nil
}

// callExpr climbs AST tree up until call expression
func callExpr(fset *token.FileSet, nodes []ast.Node) *ast.CallExpr {
	for _, node := range nodes {
		callExpr, ok := node.(*ast.CallExpr)
		if ok {
			return callExpr
		}
	}
	return nil
}

// shortTyoe returns shorthand type notation without specifying type's import path
func shortType(t types.Type) string {
	return types.TypeString(t, func(*types.Package) string {
		return ""
	})
}

// shortParam returns shorthand parameter notation in form "name type" without specifying type's import path
func shortParam(param *types.Var) string {
	ret := param.Name()
	if ret != "" {
		ret += " "
	}
	return ret + shortType(param.Type())
}
