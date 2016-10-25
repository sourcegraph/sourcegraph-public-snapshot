package langserver

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
)

type diagnostics map[string][]*lsp.Diagnostic // map of URI to diagnostics (for PublishDiagnosticParams)

// publishDiagnostics sends diagnostic information (such as compile
// errors) to the client.
func (h *LangHandler) publishDiagnostics(ctx context.Context, conn JSONRPC2Conn, diags diagnostics) error {
	for filename, diags := range diags {
		params := lsp.PublishDiagnosticsParams{
			URI:         "file://" + filename,
			Diagnostics: make([]lsp.Diagnostic, len(diags)),
		}
		for i, d := range diags {
			params.Diagnostics[i] = *d
		}
		if err := conn.Notify(ctx, "textDocument/publishDiagnostics", params); err != nil {
			return err
		}
	}
	return nil
}

func parseLoaderError(errStr string) (filename string, diag *lsp.Diagnostic, err error) {
	c1 := strings.Index(errStr, ":")
	if c1 <= 0 || c1 == len(errStr)-1 {
		return "", nil, fmt.Errorf("invalid error message 1: %q", errStr)
	}
	c2 := c1 + 1 + strings.Index(errStr[c1+1:], ":")
	if c2 <= c1 || c2 == len(errStr)-1 {
		return "", nil, fmt.Errorf("invalid error message 2: %q", errStr)
	}
	c3 := c2 + 1 + strings.Index(errStr[c2+1:], ":")
	if c3 <= c2 || c3 == len(errStr)-1 {
		return "", nil, fmt.Errorf("invalid error message 3: %q", errStr)
	}

	filename = errStr[:c1]
	line, err := strconv.Atoi(errStr[c1+1 : c2])
	if err != nil {
		return "", nil, err
	}
	col, err := strconv.Atoi(errStr[c2+1 : c3])
	if err != nil {
		return "", nil, err
	}
	return filename, &lsp.Diagnostic{
		Range: lsp.Range{
			// LSP is 0-indexed, so subtract one from the numbers Go
			// reports.
			Start: lsp.Position{Line: line - 1, Character: col - 1},
			End:   lsp.Position{Line: line - 1, Character: col - 1},
		},
		Severity: lsp.Error,
		Source:   "go",
		Message:  strings.TrimSpace(errStr[c3+1:]),
	}, nil
}
