package langserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/tools/refactor/importgraph"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/go-langserver/pkg/lspext"
	"github.com/sourcegraph/jsonrpc2"
)

// NewHandler creates a Go language server handler.
func NewHandler() jsonrpc2.Handler {
	return jsonrpc2.HandlerWithError((&LangHandler{
		HandlerShared: &HandlerShared{},
	}).handle)
}

// LangHandler is a Go language server LSP/JSON-RPC handler.
type LangHandler struct {
	mu sync.Mutex
	HandlerCommon
	*HandlerShared
	init *InitializeParams // set by "initialize" request

	typecheckCache cache
	symbolCache    cache

	// cache the reverse import graph
	importGraphOnce sync.Once
	importGraph     importgraph.Graph
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
		if err := h.HandlerShared.Reset(init.RootPath, !init.NoOSFileSystemAccess); err != nil {
			return err
		}
	}
	h.init = init
	h.resetCaches(false)
	return nil
}

func (h *LangHandler) resetCaches(lock bool) {
	if lock {
		h.mu.Lock()
	}

	h.importGraphOnce = sync.Once{}
	h.importGraph = nil

	if h.typecheckCache == nil {
		h.typecheckCache = newTypecheckCache()
	} else {
		h.typecheckCache.Purge()
	}

	if h.symbolCache == nil {
		h.symbolCache = newSymbolCache()
	} else {
		h.symbolCache.Purge()
	}

	if lock {
		h.mu.Unlock()
	}
}

// handle implements jsonrpc2.Handler.
func (h *LangHandler) handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	return h.Handle(ctx, conn, req)
}

// Handle implements jsonrpc2.Handler, except conn is an interface
// type for testability. The handle method implements jsonrpc2.Handler
// exactly.
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

		// HACK: RootPath is not a URI, but historically we treated it
		// as such. Convert it to a file URI
		if !strings.HasPrefix(params.RootPath, "file://") {
			params.RootPath = "file://" + params.RootPath
		}

		if err := h.reset(&params); err != nil {
			return nil, err
		}

		// PERF: Kick off a workspace/symbol in the background to warm up the server
		if yes, _ := strconv.ParseBool(envWarmupOnInitialize); yes {
			go func() {
				ctx, cancel := context.WithDeadline(ctx, time.Now().Add(30*time.Second))
				defer cancel()
				_, _ = h.handleWorkspaceSymbol(ctx, conn, req, lspext.WorkspaceSymbolParams{
					Query: "",
					Limit: 100,
				})
			}()
		}

		return lsp.InitializeResult{
			Capabilities: lsp.ServerCapabilities{
				TextDocumentSync:             lsp.TDSKFull,
				DefinitionProvider:           true,
				DocumentFormattingProvider:   true,
				DocumentSymbolProvider:       true,
				HoverProvider:                true,
				ReferencesProvider:           true,
				WorkspaceSymbolProvider:      true,
				XWorkspaceReferencesProvider: true,
				XDefinitionProvider:          true,
				XWorkspaceSymbolByProperties: true,
				SignatureHelpProvider:        &lsp.SignatureHelpOptions{TriggerCharacters: []string{"(", ","}},
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
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params lsp.TextDocumentPositionParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		return h.handleHover(ctx, conn, req, params)

	case "textDocument/definition":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params lsp.TextDocumentPositionParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		return h.handleDefinition(ctx, conn, req, params)

	case "textDocument/xdefinition":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params lsp.TextDocumentPositionParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		return h.handleXDefinition(ctx, conn, req, params)

	case "textDocument/references":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params lsp.ReferenceParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		return h.handleTextDocumentReferences(ctx, conn, req, params)

	case "textDocument/documentSymbol":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params lsp.DocumentSymbolParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		return h.handleTextDocumentSymbol(ctx, conn, req, params)

	case "textDocument/signatureHelp":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params lsp.TextDocumentPositionParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		return h.handleTextDocumentSignatureHelp(ctx, conn, req, params)

	case "textDocument/formatting":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params lsp.DocumentFormattingParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		return h.handleTextDocumentFormatting(ctx, conn, req, params)

	case "workspace/symbol":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params lspext.WorkspaceSymbolParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		return h.handleWorkspaceSymbol(ctx, conn, req, params)

	case "workspace/xreferences":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params lspext.WorkspaceReferencesParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		return h.handleWorkspaceReferences(ctx, conn, req, params)

	default:
		if IsFileSystemRequest(req.Method) {
			err := h.HandleFileSystemRequest(ctx, req)
			h.resetCaches(true) // a file changed, so we must re-typecheck and re-enumerate symbols
			return nil, err
		}

		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("method not supported: %s", req.Method)}
	}
}
