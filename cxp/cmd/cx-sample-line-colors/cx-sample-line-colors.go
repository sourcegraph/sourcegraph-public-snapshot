package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/sourcegraph/sourcegraph/cxp"
	"github.com/sourcegraph/sourcegraph/cxp/pkg/cxpmain"
	"github.com/sourcegraph/sourcegraph/xlang/lspext"
)

//docker:user sourcegraph

func main() {
	cxpmain.Main("cx-sample-line-colors", func() jsonrpc2.Handler { return jsonrpc2.HandlerWithError((&handler{}).handle) })
}

type handler struct {
	mu       sync.Mutex
	initOpts initializationOptions
}

type initializationOptions struct {
	Colors []string `json:"colors"`
}

func (h *handler) handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
	h.mu.Lock()
	initOpts := h.initOpts
	h.mu.Unlock()

	switch req.Method {
	case "initialize":
		var params struct {
			// unused: lsp.InitializeParams
			InitializationOptions initializationOptions `json:"initializationOptions"`
		}
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		h.mu.Lock()
		h.initOpts = params.InitializationOptions
		h.mu.Unlock()

		return lsp.InitializeResult{
			Capabilities: lsp.ServerCapabilities{
				Experimental: cxp.ExperimentalServerCapabilities{
					DecorationsProvider: true,
				},
			},
		}, nil

	case "shutdown", "exit":
		return nil, nil

	case "textDocument/decorations":
		if len(initOpts.Colors) == 0 {
			return nil, nil
		}
		const maxLines = 100
		decorations := make([]lspext.TextDocumentDecoration, maxLines)
		for i := 0; i < maxLines; i++ {
			decorations[i] = lspext.TextDocumentDecoration{
				Range:           lsp.Range{Start: lsp.Position{Line: i}, End: lsp.Position{Line: i}},
				IsWholeLine:     true,
				BackgroundColor: initOpts.Colors[i%len(initOpts.Colors)],
			}
		}
		return decorations, nil
	}

	return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("method not supported: %s", req.Method)}
}
