package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"sync"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/sourcegraph/jsonx"
	"github.com/sourcegraph/sourcegraph/cxp"
	"github.com/sourcegraph/sourcegraph/cxp/pkg/cxpmain"
	"github.com/sourcegraph/sourcegraph/xlang/lspext"
	"github.com/sourcegraph/sourcegraph/xlang/uri"
)

//docker:user sourcegraph

// Global data structure that counts hovers on a root URI -> document URI -> line (0-indexed).
var (
	hoverCountsMu sync.Mutex
	hoverCounts   = make(map[uri.URI]map[lsp.DocumentURI]map[int]int)
)

func getHoverCountsByLine(root uri.URI, document lsp.DocumentURI) map[int]int {
	hoverCountsMu.Lock()
	defer hoverCountsMu.Unlock()
	return hoverCounts[root][document]
}

func incrementHoverCount(root uri.URI, document lsp.DocumentURI, line int) {
	hoverCountsMu.Lock()
	defer hoverCountsMu.Unlock()
	rootEntry, ok := hoverCounts[root]
	if !ok {
		rootEntry = map[lsp.DocumentURI]map[int]int{}
		hoverCounts[root] = rootEntry
	}
	documentEntry, ok := rootEntry[document]
	if !ok {
		documentEntry = map[int]int{}
		rootEntry[document] = documentEntry
	}
	documentEntry[line]++
}

func main() {
	cxpmain.Main("cxp-hover-heatmap", func() jsonrpc2.Handler { return jsonrpc2.AsyncHandler(jsonrpc2.HandlerWithError((&handler{}).handle)) })
}

type handler struct {
	mu                      sync.Mutex
	initialized             bool
	rootURI                 *uri.URI // doesn't change after initialize
	settings                extensionSettings
	openDocuments           map[lsp.DocumentURI]struct{}
	registeredContributions bool
}

type extensionSettings struct {
	Hide bool `json:"hide,omitempty"`
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
	initialized := h.initialized
	settings := h.settings
	registeredContributions := h.registeredContributions
	h.mu.Unlock()

	switch req.Method {
	case "initialize":
		if initialized {
			return nil, errors.New("already initialized")
		}

		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		cap, err := cxp.ParseClientCapabilities(*req.Params)
		if err != nil {
			return nil, err
		}
		if cap.Decoration == nil || !cap.Decoration.Dynamic {
			return nil, errors.New("client does not support published decorations")
		}

		var params cxp.InitializeParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}

		rootURI, err := uri.Parse(string(params.OriginalRootURI))
		if err != nil {
			return nil, err
		}
		var settings extensionSettings
		if merged := params.InitializationOptions.Settings.Merged; merged != nil {
			if err := json.Unmarshal(*merged, &settings); err != nil {
				return nil, err
			}
		}
		h.mu.Lock()
		h.initialized = true
		h.rootURI = rootURI
		h.settings = settings
		h.openDocuments = map[lsp.DocumentURI]struct{}{}
		h.mu.Unlock()

		return cxp.InitializeResult{
			Capabilities: cxp.ServerCapabilities{
				ServerCapabilities: lsp.ServerCapabilities{
					HoverProvider: true,
					TextDocumentSync: &lsp.TextDocumentSyncOptionsOrKind{
						Options: &lsp.TextDocumentSyncOptions{OpenClose: true},
					},
					ExecuteCommandProvider: &lsp.ExecuteCommandOptions{
						Commands: []string{
							toggleCommandID,
						},
					},
				},
				DecorationProvider: &cxp.DecorationProviderServerCapabilities{
					DecorationCapabilityOptions: cxp.DecorationCapabilityOptions{Dynamic: true},
				},
			},
		}, nil

	case "initialized":
		if !registeredContributions {
			if err := registerContributions(ctx, conn, settings, registeredContributions); err != nil {
				return nil, errors.WithMessage(err, "publish contributions")
			}
			h.mu.Lock()
			h.registeredContributions = true
			h.mu.Unlock()
		}
		return nil, nil

	case "shutdown", "exit":
		return nil, nil

	case "workspace/didChangeConfiguration":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params struct {
			Settings struct {
				Merged extensionSettings `json:"merged"`
			} `json:"settings"`
		}
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		if reflect.DeepEqual(settings, params.Settings.Merged) {
			// Nothing to do; we already have the latest settings.
			return nil, nil
		}

		h.mu.Lock()
		h.settings = params.Settings.Merged
		h.mu.Unlock()

		if err := h.publishDecorations(ctx, conn, params.Settings.Merged); err != nil {
			return nil, errors.WithMessage(err, "publish decorations")
		}

		if !registeredContributions {
			if err := registerContributions(ctx, conn, params.Settings.Merged, registeredContributions); err != nil {
				return nil, errors.WithMessage(err, "publish contributions")
			}
			h.mu.Lock()
			h.registeredContributions = true
			h.mu.Unlock()
		}

		return nil, nil

	case "workspace/executeCommand":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params lsp.ExecuteCommandParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		switch params.Command {
		case toggleCommandID:
			settings.Hide = !settings.Hide
			if err := h.updateSettings(ctx, conn, settings); err != nil {
				return nil, errors.WithMessage(err, "update settings")
			}
			return nil, nil

		default:
			return nil, fmt.Errorf("command is not executable on server: %q", params.Command)
		}

	case "textDocument/didOpen":
		var params lsp.DidOpenTextDocumentParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		h.mu.Lock()
		h.openDocuments[params.TextDocument.URI] = struct{}{}
		h.mu.Unlock()

		_ = conn.Notify(ctx, "textDocument/publishDecorations", cxp.TextDocumentPublishDecorationsParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: params.TextDocument.URI},
			Decorations:  createDecorations(*h.rootURI, params.TextDocument.URI, settings),
		})

		return nil, nil

	case "textDocument/hover":
		var params lsp.TextDocumentPositionParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		incrementHoverCount(*h.rootURI, params.TextDocument.URI, params.Position.Line)

		// Show immediate visual feedback.
		_ = conn.Notify(ctx, "textDocument/publishDecorations", cxp.TextDocumentPublishDecorationsParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: params.TextDocument.URI},
			Decorations:  createDecorations(*h.rootURI, params.TextDocument.URI, settings),
		})

		return lsp.Hover{
			Contents: []lsp.MarkedString{},
		}, nil
	}

	return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("method not supported: %s", req.Method)}
}

func (h *handler) updateSettings(ctx context.Context, conn *jsonrpc2.Conn, newSettings extensionSettings) error {
	if err := h.publishDecorations(ctx, conn, newSettings); err != nil {
		return errors.WithMessage(err, "publish decorations")
	}
	if err := registerContributions(ctx, conn, newSettings, true); err != nil {
		return errors.WithMessage(err, "publish contributions")
	}
	h.mu.Lock()
	h.settings = newSettings
	h.mu.Unlock()

	// Run async because we are currently handling a client request, and we would deadlock otherwise.
	go func() {
		if err := conn.Call(ctx, "configuration/update", cxp.ConfigurationUpdateParams{
			Path:  jsonx.Path{},
			Value: newSettings,
		}, nil); err != nil {
			log.Println("configuration/update error:", err)
		}
	}()
	return nil
}

func (h *handler) publishDecorations(ctx context.Context, conn *jsonrpc2.Conn, settings extensionSettings) error {
	h.mu.Lock()
	openDocuments := cloneMap(h.openDocuments)
	h.mu.Unlock()
	for uri := range openDocuments {
		if err := conn.Notify(ctx, "textDocument/publishDecorations", cxp.TextDocumentPublishDecorationsParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: uri},
			Decorations:  createDecorations(*h.rootURI, uri, settings),
		}); err != nil {
			return err
		}
	}
	return nil
}

func createDecorations(root uri.URI, document lsp.DocumentURI, settings extensionSettings) []lspext.TextDocumentDecoration {
	decorations := []lspext.TextDocumentDecoration{}
	if settings.Hide {
		return decorations
	}

	counts := getHoverCountsByLine(root, document)
	var maxCount int
	for _, count := range counts {
		if count > maxCount {
			maxCount = count
		}
	}
	for line, count := range counts {
		decorations = append(decorations, lspext.TextDocumentDecoration{
			Range:           lsp.Range{Start: lsp.Position{Line: line}, End: lsp.Position{Line: line}},
			IsWholeLine:     true,
			BackgroundColor: getColorByRelativeFrequency(count, maxCount),
		})
	}
	return decorations
}

func registerContributions(ctx context.Context, conn *jsonrpc2.Conn, settings extensionSettings, unregister bool) error {
	var showHide string
	if settings.Hide {
		showHide = "Show"
	} else {
		showHide = "Hide"
	}
	if err := conn.Call(ctx, "client/registerCapability", cxp.RegistrationParams{
		Registrations: []cxp.Registration{
			{
				ID:     "main",
				Method: "window/contribution",
				RegisterOptions: &cxp.Contributions{
					Commands: []*cxp.CommandContribution{
						{
							Command: toggleCommandID,
							Title:   "Heatmap",
							Detail:  showHide + " hover heatmap",
						},
					},
					Menus: &cxp.MenuContributions{
						CommandPalette: []*cxp.MenuItemContribution{{Command: toggleCommandID}},
					},
				},
				OverwriteExisting: unregister,
			},
		},
	}, nil); err != nil {
		return errors.WithMessage(err, "client/unregisterCapability")
	}
	return nil
}

const toggleCommandID = "hover-heatmap.toggle"

// getColorByRelativeFrequency determines the background color for a given line. `maxCount` is the
// highest number of hovers that a line in this file has. We calculate the line color by the
// relative frequency of hovers, with `maxCount` being the denominator, meaning `maxCount` will
// always be the darkest red in a file.
func getColorByRelativeFrequency(count, maxCount int) string {
	redHue, greenHue := 0, 116
	x := float64(count)
	if maxCount > 0 {
		x = float64(count) / float64(maxCount)
	}

	const alpha = 1.0
	return fmt.Sprintf("hsla(%d, 100%%, 50%%, %.2f)", greenHue-int(x*float64(greenHue-redHue)), alpha)
}

func cloneMap(m map[lsp.DocumentURI]struct{}) map[lsp.DocumentURI]struct{} {
	m2 := make(map[lsp.DocumentURI]struct{}, len(m))
	for k, v := range m {
		m2[k] = v
	}
	return m2
}
