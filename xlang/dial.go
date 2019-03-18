package xlang

import (
	"context"
	"encoding/json"
	"log"
	"net"

	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/go-lsp/lspext"
	"github.com/sourcegraph/jsonrpc2"
)

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
	return jsonrpc2.NewConn(context.Background(), jsonrpc2.NewBufferedStream(conn, jsonrpc2.VSCodeObjectCodec{}), jsonrpc2.AsyncHandler(jsonrpc2.HandlerWithError(h.handle)), connOpt...), nil
}

// ClientHandler is a JSON-RPC 2.0 handler for the client that
// communicates with the LSP proxy.
type ClientHandler struct {
	// RecvDiagnostics is called when textDocument/publishDiagnostics is received
	RecvDiagnostics func(uri lsp.DocumentURI, diags []lsp.Diagnostic)

	// RecvPartialResult is called when $/partialResult is received
	RecvPartialResult func(id lsp.ID, patch interface{})
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

	case "$/partialResult":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params lspext.PartialResultParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		if h.RecvPartialResult != nil {
			h.RecvPartialResult(params.ID, params.Patch)
		}

	case "window/showMessage", "window/logMessage":
		// These messages have already been logged by lsp-proxy.
		return nil, nil

	default:
		log.Printf("xlang client handler: ignoring %q from language server", req.Method)
		return nil, nil
	}
	return nil, nil
}
