package golang

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"

	"github.com/sourcegraph/jsonrpc2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/cmdutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/lsp"
)

// cmdOutput is a helper around c.Output which logs the command, how long it
// took to run, and a nice error in the event of failure.
//
// The specified env is set as cmd.Env (because we do this at ALL callsites
// today anyway).
func cmdOutput(ctx context.Context, env []string, c *exec.Cmd) ([]byte, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, c.Args[0])
	defer span.Finish()
	span.SetTag("command", strings.Join(c.Args, " "))
	span.SetTag("env", strings.Join(env, "; "))

	c.Env = env
	start := time.Now()
	stdout, err := cmdutil.Output(c)
	log.Printf("TIME: %v '%s'\n", time.Since(start), strings.Join(c.Args, " "))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return stdout, nil
}

// Handler represents an LSP handler for one user.
type Handler struct {
	mu sync.Mutex

	init *lsp.InitializeParams // set by "initialize" req
}

// reset clears all internal state in h.
func (h *Handler) reset(init *lsp.InitializeParams) {
	h.init = init
}

func (h *Handler) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
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
				DefinitionProvider:      true,
				HoverProvider:           true,
				ReferencesProvider:      true,
				WorkspaceSymbolProvider: true,
			},
		}, nil

	case "shutdown":
		// Result is undefined, per
		// https://github.com/Microsoft/language-server-protocol/blob/master/protocol.md#shutdown-request.
		return nil, nil

	case "textDocument/hover":
		var params lsp.TextDocumentPositionParams
		if err = json.Unmarshal(*req.Params, &params); err != nil {
			return
		}
		return h.handleHover(ctx, req, params)

	case "textDocument/definition":
		var params lsp.TextDocumentPositionParams
		if err = json.Unmarshal(*req.Params, &params); err != nil {
			return
		}
		return h.handleDefinition(ctx, req, params)

	case "textDocument/references":
		var params lsp.ReferenceParams
		if err = json.Unmarshal(*req.Params, &params); err != nil {
			break
		}
		return h.handleReferences(ctx, req, params)

	case "workspace/symbol":
		var params lsp.WorkspaceSymbolParams
		if err = json.Unmarshal(*req.Params, &params); err != nil {
			break
		}
		return h.handleSymbol(ctx, req, params)
	}

	return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("method not supported: %s", req.Method)}
}
