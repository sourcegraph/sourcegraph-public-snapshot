package ctags

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"runtime"

	opentracing "github.com/opentracing/opentracing-go"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/lspext"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/vfsutil"
)

func Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
	// Prevent any uncaught panics from taking the entire server down.
	var err error
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
	response, err := handleRequest(ctx, conn, req)
	if err != nil {
		log.Println(err)
	}
	return response, err
}

func handleRequest(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
	operationName := "LS Serve: " + req.Method
	span, ctx := opentracing.StartSpanFromContext(ctx, operationName)
	defer span.Finish()

	switch req.Method {
	case "initialize":
		info := ctxInfo(ctx)
		var params lspext.InitializeParams
		json.Unmarshal(*req.Params, &params)
		info.mode = params.Mode
		info.fs = &vfsutil.RemoteProxyFS{Conn: conn}

		// Start downloading and analyzing the tags asynchronously.
		go getTags(ctx)
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
		s, err := handleSymbol(ctx, params)
		if err != nil {
			return nil, err
		}
		if s == nil {
			return nil, nil
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
		return handleDefinition(ctx, params)

	case "textDocument/references":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params lsp.ReferenceParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		return handleReferences(ctx, params)

	case "textDocument/hover":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params lsp.TextDocumentPositionParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		return handleHover(ctx, params)
	}

	return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("method not supported: %s", req.Method)}
}
