package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"

	"github.com/sourcegraph/jsonrpc2"

	"github.com/sourcegraph/go-langserver/langserver"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/go-langserver/pkg/lspext"
)

// jsonrpc2ConnImpl implements langserver.JSONRPC2Conn. See
// langserver.JSONRPC2Conn for more information.
type jsonrpc2ConnImpl struct {
	rewriteURI func(lsp.DocumentURI) (lsp.DocumentURI, error)
	conn       *jsonrpc2.Conn
}

func (c *jsonrpc2ConnImpl) Call(ctx context.Context, method string, params, result interface{}, opt ...jsonrpc2.CallOption) error {
	switch method {
	case "xcache/get":
		// we just pass cache requests through
		return c.conn.Call(ctx, method, params, result, opt...)

	default:
		return &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("gobuildserver client: method not supported: %q", method)}
	}
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

	case "$/partialResult":
		params := params.(*lspext.PartialResultParams)

		if _, ok := params.Patch.(*json.RawMessage); ok {
			// initial patch, just pass on since it is empty
			return c.conn.Notify(ctx, method, params, opt...)
		}

		rewriter, ok := params.Patch.(langserver.RewriteURIer)
		if !ok {
			return errors.New("buildserver received partialResult which does not support RewriteURI")
		}

		var rewriteErr error
		rewriter.RewriteURI(func(u lsp.DocumentURI) lsp.DocumentURI {
			u, err := c.rewriteURI(u)
			if err != nil {
				rewriteErr = errors.Wrap(err, "buildserver failde to rewrite partialResult")
			}
			return u
		})
		if rewriteErr != nil {
			return rewriteErr
		}

		return c.conn.Notify(ctx, method, params, opt...)

	case "xcache/set":
		// we just pass cache requests through
		return c.conn.Notify(ctx, method, params, opt...)

	default:
		return &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("gobuildserver client: notify method not supported: %q", method)}
	}
}

func (c *jsonrpc2ConnImpl) Close() error {
	// we want to handle closing, so ignore
	return nil
}
