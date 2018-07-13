package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/sourcegraph/sourcegraph/cxp"
	"github.com/sourcegraph/sourcegraph/cxp/pkg/cxpmain"
)

//docker:user sourcegraph

func main() {
	cxpmain.Main("cx-sample-lsp", func() jsonrpc2.Handler { return jsonrpc2.HandlerWithError((&handler{}).handle) })
}

type handler struct {
	mu      sync.Mutex
	initRaw []byte
}

func (h *handler) handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
	h.mu.Lock()
	initRaw := h.initRaw
	h.mu.Unlock()

	switch req.Method {
	case "initialize":
		h.mu.Lock()
		var buf bytes.Buffer
		_ = json.Indent(&buf, *req.Params, "", "  ")
		h.initRaw = buf.Bytes()
		h.mu.Unlock()

		return cxp.InitializeResult{
			Capabilities: cxp.ServerCapabilities{
				ServerCapabilities: lsp.ServerCapabilities{
					HoverProvider:      true,
					DefinitionProvider: true,
					ReferencesProvider: true,
				},
			},
		}, nil

	case "shutdown", "exit":
		return nil, nil

	case "textDocument/hover":
		var params lsp.TextDocumentPositionParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}

		contents := []lsp.MarkedString{lsp.RawMarkedString("**Sourcegraph extension** got LSP `initialize`:")}
		contents = append(contents, lsp.MarkedString{Language: "javascript", Value: string(initRaw)})

		pos := params.Position
		return lsp.Hover{
			Contents: contents,
			Range: &lsp.Range{
				Start: lsp.Position{Line: pos.Line, Character: pos.Character},
				End:   lsp.Position{Line: pos.Line, Character: pos.Character},
			},
		}, nil

	case "textDocument/definition":
		var params lsp.TextDocumentPositionParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		return lsp.Location{
			URI: params.TextDocument.URI,
			Range: lsp.Range{
				Start: lsp.Position{Line: 0, Character: 0},
				End:   lsp.Position{Line: 0, Character: 0},
			},
		}, nil

	case "textDocument/references":
		var params lsp.ReferenceParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		return []lsp.Location{lsp.Location{
			URI: params.TextDocument.URI,
			Range: lsp.Range{
				Start: lsp.Position{Line: 0, Character: 0},
				End:   lsp.Position{Line: 0, Character: 3},
			},
		}, lsp.Location{
			URI: params.TextDocument.URI,
			Range: lsp.Range{
				Start: lsp.Position{Line: 0, Character: 4},
				End:   lsp.Position{Line: 0, Character: 6},
			},
		}}, nil

	case "workspace/symbol":
		var params lsp.WorkspaceSymbolParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		return []lsp.SymbolInformation{{
			Name: "fooFunc",
			Kind: lsp.SKFunction,
			Location: lsp.Location{
				URI:   "file:///some-file",
				Range: lsp.Range{Start: lsp.Position{Line: 0, Character: 0}, End: lsp.Position{Line: 0, Character: 3}},
			},
			ContainerName: "pkg",
		}, {
			Name: "BarClass",
			Kind: lsp.SKClass,
			Location: lsp.Location{
				URI:   "file:///another-file",
				Range: lsp.Range{Start: lsp.Position{Line: 1, Character: 2}, End: lsp.Position{Line: 1, Character: 4}},
			},
			ContainerName: "pkg",
		}}, nil
	}

	return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("method not supported: %s", req.Method)}
}
