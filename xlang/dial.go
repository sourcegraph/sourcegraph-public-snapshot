package xlang

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/sourcegraph/jsonrpc2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/lsp"
)

// NewDefaultClient returns a new one-shot connection to the LSP proxy server at
// $LSP_PROXY. This is quick (except TCP connection time) because the LSP proxy
// server retains the same underlying connection.
func NewDefaultClient() (*jsonrpc2.Conn, error) {
	addr := os.Getenv("LSP_PROXY")
	if addr == "" {
		return nil, errors.New("no LSP_PROXY env var set (need address to LSP proxy)")
	}

	dialCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return DialProxy(dialCtx, addr, nil)
}

// DialProxy creates a new JSON-RPC 2.0 connection to the LSP proxy
// server at the given address.
func DialProxy(dialCtx context.Context, addr string, h *ClientHandler, connOpt ...jsonrpc2.ConnOpt) (*jsonrpc2.Conn, error) {
	if h == nil {
		h = &ClientHandler{}
	}
	conn, err := (&net.Dialer{}).DialContext(dialCtx, "tcp", addr)
	if err != nil {
		return nil, err
	}
	return jsonrpc2.NewConn(context.Background(), conn, jsonrpc2.HandlerWithError(h.handle), connOpt...), nil
}

// ClientHandler is a JSON-RPC 2.0 handler for the client that
// communicates with the LSP proxy.
type ClientHandler struct {
	RecvDiagnostics func(uri string, diags []lsp.Diagnostic) // called when textDocument/publishDiagnostics is received
}

func (h *ClientHandler) handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	switch req.Method {
	case "textDocument/publishDiagnostics":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params lsp.PublishDiagnosticsParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		if h.RecvDiagnostics != nil {
			h.RecvDiagnostics(params.URI, params.Diagnostics)
		}

	default:
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("client handler: method not found: %q", req.Method)}
	}
	return nil, nil
}
