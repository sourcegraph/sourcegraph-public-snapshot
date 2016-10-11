package langserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"runtime"
	"sync"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"

	"github.com/sourcegraph/jsonrpc2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/lsp"
)

// LangHandler is a Go language server LSP/JSON-RPC handler.
//
// It can operate as an LSP server 100% independently of the Go build
// server (BuildHandler) for use on a local file system.
type LangHandler struct {
	mu sync.Mutex
	HandlerCommon
	*HandlerShared
	init *InitializeParams // set by "initialize" request

	// cached symbols
	pkgSymCacheMu sync.Mutex
	pkgSymCache   map[string][]lsp.SymbolInformation

	// cached typechecking results
	cacheMus map[typecheckKey]*sync.Mutex
	cache    map[typecheckKey]typecheckResult
}

// reset clears all internal state in h.
func (h *LangHandler) reset(init *InitializeParams) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if err := h.HandlerCommon.Reset(init.RootPath); err != nil {
		return err
	}
	if !h.HandlerShared.Shared {
		// Only reset the shared data if this lang server is running
		// by itself.
		if err := h.HandlerShared.Reset(init.RootPath); err != nil {
			return err
		}
	}
	h.init = init
	h.cacheMus = map[typecheckKey]*sync.Mutex{}
	h.cache = map[typecheckKey]typecheckResult{}
	return nil
}

// Handle implements jsonrpc2.Handler.
func (h *LangHandler) Handle(ctx context.Context, conn JSONRPC2Conn, req *jsonrpc2.Request) (result interface{}, err error) {
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

	h.mu.Lock()
	if req.Method != "initialize" && h.init == nil {
		h.mu.Unlock()
		return nil, errors.New("server must be initialized")
	}
	h.mu.Unlock()
	if err := h.CheckReady(); err != nil {
		if req.Method == "exit" {
			err = nil
		}
		return nil, err
	}

	if conn, ok := conn.(*jsonrpc2.Conn); ok && conn != nil {
		h.InitTracer(conn)
	}
	span, ctx, err := h.SpanForRequest(ctx, "lang", req, opentracing.Tags{"mode": "go"})
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.LogEvent(fmt.Sprintf("error: %v", err))
		}
		span.Finish()
	}()

	switch req.Method {
	case "initialize":
		if h.init != nil {
			return nil, errors.New("language server is already initialized")
		}
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params InitializeParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		if err := h.reset(&params); err != nil {
			return nil, err
		}
		return lsp.InitializeResult{
			Capabilities: lsp.ServerCapabilities{
				TextDocumentSync:        lsp.TDSKFull,
				DefinitionProvider:      true,
				HoverProvider:           true,
				ReferencesProvider:      true,
				WorkspaceSymbolProvider: true,
			},
		}, nil

	case "shutdown":
		h.ShutDown()
		return nil, nil

	case "exit":
		if c, ok := conn.(*jsonrpc2.Conn); ok {
			c.Close()
		}
		return nil, nil

	case "textDocument/hover":
		var params lsp.TextDocumentPositionParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		return h.handleHover(ctx, conn, req, params)

	case "textDocument/definition":
		var params lsp.TextDocumentPositionParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		return h.handleDefinition(ctx, conn, req, params)

	case "textDocument/references":
		var params lsp.ReferenceParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		return h.handleReferences(ctx, conn, req, params)

	case "workspace/symbol":
		var params lsp.WorkspaceSymbolParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		return h.handleSymbol(ctx, conn, req, params)

	default:
		if IsFileSystemRequest(req.Method) {
			return nil, h.HandleFileSystemRequest(ctx, req)
		}

		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("method not supported: %s", req.Method)}
	}
}
