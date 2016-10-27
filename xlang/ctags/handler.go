package ctags

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"runtime"
	"sync"

	opentracing "github.com/opentracing/opentracing-go"

	"github.com/sourcegraph/ctxvfs"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/ctags/parser"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/lspext"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/vfsutil"
)

var emptyArray = make([]string, 0)

// Handler handlers LSP requests for one repository
type Handler struct {
	// fs is the virtual filesystem backed by xlang infrastrucuture.
	fs ctxvfs.FileSystem

	// tagsMu protects tags. We want to be careful to not run ctags more than
	// once for one project, so this is used in the get tags method.
	tagsMu sync.Mutex

	// tags is the Go form of the ctags output for this project. We compute and
	// save it so that we don't have to parse the ctags file each time, and so
	// we don't have to store as much state on disk.
	tags []parser.Tag

	// mode is the language that we care about for this connection.
	mode string
}

var ErrMustInit = errors.New("initialize must be called before other methods")

func (h *Handler) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (_ interface{}, err error) {
	// Prevent any uncaught panics from taking the entire server down.
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("unexpected panic: %v", r)

			// Same as net/http
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			log.Printf("panic serving %v: %v\n%s", req.Method, r, buf)
		}
	}()
	response, err := h.HandleRequest(ctx, conn, req)
	if err != nil {
		log.Println(err)
	}
	return response, err
}

func (h *Handler) HandleRequest(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
	operationName := "LS Serve: " + req.Method
	span, ctx := opentracing.StartSpanFromContext(ctx, operationName)
	defer span.Finish()

	h.fs = &vfsutil.RemoteProxyFS{Conn: conn}

	switch req.Method {
	case "initialize":
		var params lspext.InitializeParams
		json.Unmarshal(*req.Params, &params)
		h.mode = params.Mode

		// Start downloading and analyzing the tags asynchronously.
		go h.getTags(ctx)
		return lsp.InitializeResult{
			Capabilities: lsp.ServerCapabilities{
				WorkspaceSymbolProvider: true,
				DefinitionProvider:      true,
				ReferencesProvider:      true,
				HoverProvider:           true,
			},
		}, nil

	case "shutdown":
		return nil, nil

	case "exit":
		return nil, nil

	case "workspace/symbol":
		var params lsp.WorkspaceSymbolParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		s, err := h.handleSymbol(ctx, params)
		if err != nil {
			return nil, err
		}
		if s == nil {
			return emptyArray, nil
		}
		return s, nil

	case "textDocument/definition":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params lsp.TextDocumentPositionParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		return h.handleDefinition(ctx, params)

	case "textDocument/references":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params lsp.ReferenceParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		return h.handleReferences(ctx, params)

	case "textDocument/hover":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params lsp.TextDocumentPositionParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		return h.handleHover(ctx, params)
	}

	return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("method not supported: %s", req.Method)}
}
