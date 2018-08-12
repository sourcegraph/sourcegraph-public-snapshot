// Command cx-codecov is a Sourcegraph extension that decorates text documents
// based on code coverage data from Codecov.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/sourcegraph/sourcegraph/cxp"
	"github.com/sourcegraph/sourcegraph/cxp/pkg/cxpmain"
	"github.com/sourcegraph/sourcegraph/xlang/lspext"
	"github.com/sourcegraph/sourcegraph/xlang/uri"
)

//docker:user sourcegraph

func main() {
	cxpmain.Main("cx-codecov", func() jsonrpc2.Handler { return jsonrpc2.AsyncHandler(jsonrpc2.HandlerWithError((&handler{}).handle)) })
}

type handler struct {
	mu            sync.Mutex
	rootURI       *uri.URI
	clientCap     *cxp.ClientCapabilities
	openDocuments map[lsp.DocumentURI]struct{}
	settings      extensionSettings
}

type extensionSettings struct {
	Token           string `json:"token,omitempty"`
	LineDecorations *bool  `json:"lineDecorations,omitempty"`
}

func (h *handler) handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "CXP: "+req.Method)
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()

	h.mu.Lock()
	rootURI := h.rootURI
	clientCap := h.clientCap
	openDocuments := cloneMap(h.openDocuments)
	settings := h.settings
	h.mu.Unlock()

	switch req.Method {
	case "initialize":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		cap, err := cxp.ParseClientCapabilities(*req.Params)
		if err != nil {
			return nil, err
		}
		if cap.Decoration == nil || !cap.Decoration.Static {
			return nil, errors.New("client does not support decorations")
		}

		var params cxp.InitializeParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}

		var settings extensionSettings
		if merged := params.InitializationOptions.Settings.Merged; merged != nil {
			if err := json.Unmarshal(*merged, &settings); err != nil {
				return nil, err
			}
		}

		var rootURI *uri.URI
		if params.OriginalRootURI != "" {
			rootURI, err = uri.Parse(string(params.OriginalRootURI))
		} else {
			rootURI, err = uri.Parse(string(params.RootOrRootURI()))
		}
		if err != nil {
			return nil, err
		}

		h.mu.Lock()
		h.clientCap = cap
		h.settings = settings
		h.openDocuments = map[lsp.DocumentURI]struct{}{}
		h.rootURI = rootURI
		h.mu.Unlock()

		return cxp.InitializeResult{
			Capabilities: cxp.ServerCapabilities{
				ServerCapabilities: lsp.ServerCapabilities{
					TextDocumentSync: &lsp.TextDocumentSyncOptionsOrKind{
						Options: &lsp.TextDocumentSyncOptions{OpenClose: true},
					},
				},
				DecorationProvider: &cxp.DecorationProviderServerCapabilities{
					DecorationCapabilityOptions: cxp.DecorationCapabilityOptions{Dynamic: true},
				},
				Contributions: &cxp.Contributions{
					Commands: []*cxp.CommandContribution{
						{
							Command: setAPITokenCommandID,
							Title:   "Codecov: Set API token to view coverage for private repositories (TODO: not implemented, needs command registration and handling)",
							IconURL: iconURL,
						},
						{
							Command: toggleLineDecorationsCommandID,
							Title:   "Coverage: 73% (TODO: not implemented, needs command registration and handling)",
						},
					},
					Menus: &cxp.MenuContributions{
						EditorTitle: []*cxp.MenuItemContribution{
							{Command: setAPITokenCommandID},
							{Command: toggleLineDecorationsCommandID},
						},
					},
				},
			},
		}, nil

	case "initialized", "shutdown", "exit":
		return nil, nil

	case "workspace/didChangeConfiguration":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params struct {
			ConfigurationCascade struct {
				Merged extensionSettings `json:"merged"`
			} `json:"configurationCascade"`
		}
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		h.mu.Lock()
		h.settings = params.ConfigurationCascade.Merged
		h.mu.Unlock()

		if err := publishDecorationsForOpenDocuments(ctx, conn, clientCap, params.ConfigurationCascade.Merged, rootURI, openDocuments); err != nil {
			log.Printf("Error getting decorations after configuration change: %s.", err)
		}
		return nil, nil

	case "textDocument/didOpen":
		var params lsp.DidOpenTextDocumentParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		h.mu.Lock()
		h.openDocuments[params.TextDocument.URI] = struct{}{}
		openDocuments[params.TextDocument.URI] = struct{}{} // update clone too
		h.mu.Unlock()

		if err := publishDecorationsForOpenDocuments(ctx, conn, clientCap, settings, rootURI, openDocuments); err != nil {
			log.Printf("Error getting decorations after configuration change: %s.", err)
		}
		return nil, nil

	case "textDocument/decoration":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params lspext.TextDocumentDecorationParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		return getDecorations(ctx, settings, rootURI, params.TextDocument.URI)
	}

	return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("method not supported: %s", req.Method)}
}

const (
	setAPITokenCommandID           = "codecov.setAPIToken"
	toggleLineDecorationsCommandID = "codecov.toggleLineDecorations"
)

func cloneMap(m map[lsp.DocumentURI]struct{}) map[lsp.DocumentURI]struct{} {
	m2 := make(map[lsp.DocumentURI]struct{}, len(m))
	for k, v := range m {
		m2[k] = v
	}
	return m2
}
