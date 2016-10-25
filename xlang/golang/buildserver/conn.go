package buildserver

import (
	"context"

	"github.com/sourcegraph/jsonrpc2"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
)

// jsonrpc2ConnImpl implements langserver.JSONRPC2Conn. See
// langserver.JSONRPC2Conn for more information.
type jsonrpc2ConnImpl struct {
	rewriteURI func(string) (string, error)
	conn       *jsonrpc2.Conn
}

func (c *jsonrpc2ConnImpl) Notify(ctx context.Context, method string, params interface{}, opt ...jsonrpc2.CallOption) error {
	// Rewrite URIs from lang server (file:///src/github.com/foo/bar/f.go -> file:///f.go).
	switch method {
	case "textDocument/publishDiagnostics":
		params := params.(lsp.PublishDiagnosticsParams)

		newURI, err := c.rewriteURI(params.URI)
		if err != nil {
			return err
		}
		params.URI = newURI
		return c.conn.Notify(ctx, method, params, opt...)

	default:
		panic("build server wrapper for lang server notification sending does not support method " + method)
	}
}
