package ctags

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/jsonrpc2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/lsp"
)

func vslog(out ...string) {
	os.Stderr.WriteString(strings.Join(out, "\n") + "\n")
}

// Handler represents an LSP handler for one user.
type Handler struct {
	mu   sync.Mutex
	init *lsp.InitializeParams // set by "initialize" req
}

// reset clears all internal state in h.
func (h *Handler) reset(init *lsp.InitializeParams) {
	h.init = init
}

func (h *Handler) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (_ interface{}, err error) {
	// Coarse lock (for now) to protect h's internal state.
	h.mu.Lock()
	defer h.mu.Unlock()

	// Prevent any uncaught panics from taking the entire server down.
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("unexpected panic: %v", r)

			// Same as net/http
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			log.Printf("panic serving %v: %v\n%s", req.Method, r, buf)
			return
		}
	}()

	if req.Method != "initialize" && h.init == nil {
		return nil, errors.New("server must be initialized")
	}

	operationName := "LS Serve: " + req.Method
	var span opentracing.Span
	if req.Meta != nil {
		var header http.Header
		if err := json.Unmarshal(*req.Meta, &header); err != nil {
			return nil, err
		}
		carrier := opentracing.HTTPHeadersCarrier(header)
		clientContext, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, carrier)
		if err != nil {
			return nil, err
		}
		span = opentracing.GlobalTracer().StartSpan(operationName, ext.RPCServerOption(clientContext))
		ctx = opentracing.ContextWithSpan(ctx, span)
	} else {
		span, ctx = opentracing.StartSpanFromContext(ctx, operationName)
	}
	defer span.Finish()

	switch req.Method {
	case "initialize":
		var params lsp.InitializeParams
		if err = json.Unmarshal(*req.Params, &params); err != nil {
			return
		}
		if params.RootPath == "" {
			params.RootPath = "/"
		}
		h.reset(&params)
		return lsp.InitializeResult{
			Capabilities: lsp.ServerCapabilities{
				WorkspaceSymbolProvider: true,
			},
		}, nil

	case "shutdown":
		// Result is undefined, per
		// https://github.com/Microsoft/language-server-protocol/blob/master/protocol.md#shutdown-request.
		return nil, nil

	case "workspace/symbol":
		var params lsp.WorkspaceSymbolParams
		if err = json.Unmarshal(*req.Params, &params); err != nil {
			vslog(err.Error())
			return
		}
		return h.handleSymbol(ctx, req, params)
	}

	return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("method not supported: %s", req.Method)}
}
