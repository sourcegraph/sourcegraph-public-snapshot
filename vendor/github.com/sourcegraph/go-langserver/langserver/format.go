package langserver

import (
	"bytes"
	"context"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"path/filepath"

	"golang.org/x/tools/go/buildutil"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

func (h *LangHandler) handleTextDocumentFormatting(ctx context.Context, conn JSONRPC2Conn, req *jsonrpc2.Request, params lsp.DocumentFormattingParams) ([]lsp.TextEdit, error) {
	filename := h.FilePath(params.TextDocument.URI)
	bctx := h.BuildContext(ctx)
	fset := token.NewFileSet()
	file, err := buildutil.ParseFile(fset, bctx, nil, filepath.Dir(filename), filepath.Base(filename), parser.ParseComments)
	if err != nil {
		return nil, err
	}

	ast.SortImports(fset, file)

	var buf bytes.Buffer
	cfg := printer.Config{Mode: printer.UseSpaces | printer.TabIndent, Tabwidth: 8}
	err = cfg.Fprint(&buf, fset, file)
	if err != nil {
		return nil, err
	}

	b := buf.Bytes()
	orig, err := h.readFile(ctx, params.TextDocument.URI)
	if err != nil {
		return nil, err
	}
	if bytes.Equal(b, orig) {
		return nil, nil
	}

	return []lsp.TextEdit{
		{
			Range: lsp.Range{
				Start: lsp.Position{
					Line:      0,
					Character: 0,
				},
				End: lsp.Position{
					Line:      bytes.Count(orig, []byte("\n")),
					Character: len(orig) - bytes.LastIndexByte(orig, '\n') - 1,
				},
			},
		},
		{
			NewText: string(b),
		},
	}, nil
}
