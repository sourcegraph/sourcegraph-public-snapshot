package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/sourcegraph/jsonx"
	"github.com/sourcegraph/sourcegraph/cxp"
	"github.com/sourcegraph/sourcegraph/cxp/pkg/cxpmain"
	"github.com/sourcegraph/sourcegraph/xlang/lspext"
)

//docker:user sourcegraph

func main() {
	cxpmain.Main("cx-sample-line-colors", func() jsonrpc2.Handler { return jsonrpc2.AsyncHandler(jsonrpc2.HandlerWithError((&handler{}).handle)) })
}

type handler struct {
	mu            sync.Mutex
	initialized   bool
	settings      extensionSettings
	openDocuments map[lsp.DocumentURI]struct{}
}

type extensionSettings struct {
	Colors  []string `json:"colors,omitempty"`
	Animate bool     `json:"animate,omitempty"`
}

func (h *handler) handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
	h.mu.Lock()
	initialized := h.initialized
	settings := h.settings
	openDocuments := cloneMap(h.openDocuments)
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
		if cap.Decoration == nil || (!cap.Decoration.Static && !cap.Decoration.Dynamic) {
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
		h.mu.Lock()
		h.initialized = true
		h.settings = settings
		h.openDocuments = map[lsp.DocumentURI]struct{}{}
		h.mu.Unlock()

		go func() {
			time.Sleep(200 * time.Millisecond)
			_ = conn.Notify(ctx, "window/logMessage", lsp.LogMessageParams{
				Type:    lsp.Info,
				Message: "hello, world!",
			})
		}()

		return cxp.InitializeResult{
			Capabilities: cxp.ServerCapabilities{
				ServerCapabilities: lsp.ServerCapabilities{
					TextDocumentSync: &lsp.TextDocumentSyncOptionsOrKind{
						Options: &lsp.TextDocumentSyncOptions{OpenClose: true},
					},
					ExecuteCommandProvider: &lsp.ExecuteCommandOptions{
						Commands: []string{cycleColorsCommandID, promptColorsCommandID},
					},
				},
				DecorationProvider: &cxp.DecorationProviderServerCapabilities{DecorationCapabilityOptions: cxp.DecorationCapabilityOptions{Static: true, Dynamic: true}},
				Contributions: &cxp.Contributions{
					Commands: []*cxp.CommandContribution{
						{
							Command: cycleColorsCommandID,
							Title:   "Cycle line background colors",
							IconURL: iconURL,
							ExperimentalSettingsAction: &cxp.CommandContributionSettingsAction{
								Path: jsonx.PropertyPath("colors"),
								CycleValues: []interface{}{
									[]string{"red", "white", "blue"},
									[]string{"red", "orange", "yellow", "blue", "green", "indigo", "violet", "white"},
									[]string{"blue", "white"},
									[]string{"red", "green"},
									[]string{"black", "#DD0000", "#FFCE00"},
									[]string{"#000000", "#FFB612", "#007A4D", "#FFFFFF", "#DE3831", "#002395"},
									[]string{},
								},
							},
						},
						{
							Command: promptColorsCommandID,
							Title:   "Set colors...",
						},
					},
					Menus: &cxp.MenuContributions{
						EditorTitle: []*cxp.MenuItemContribution{
							{Command: cycleColorsCommandID},
							{Command: promptColorsCommandID},
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
			Settings struct {
				Merged extensionSettings `json:"merged"`
			} `json:"settings"`
		}
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		h.mu.Lock()
		h.settings = params.Settings.Merged
		h.mu.Unlock()

		_ = publishDecorations(ctx, conn, openDocuments, params.Settings.Merged)
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
		case promptColorsCommandID:
			var result *lsp.MessageActionItem
			if err := conn.Call(ctx, "window/showMessageRequest", lsp.ShowMessageRequestParams{
				Message: "Select line background colors:",
				Type:    lsp.Info,
				Actions: []lsp.MessageActionItem{
					{Title: "#f96316 #b200f8 #00b4f2"},
					{Title: "red green blue"},
				},
			}, &result); err != nil {
				return nil, errors.WithMessage(err, "window/showMessageRequest prompt")
			}
			if result != nil {
				settings.Colors = strings.Fields(result.Title)
				if err := publishDecorations(ctx, conn, openDocuments, settings); err != nil {
					return nil, errors.WithMessage(err, "publish decorations")
				}
				if err := conn.Call(ctx, "configuration/update", cxp.ConfigurationUpdateParams{
					Path:  jsonx.PropertyPath("colors"),
					Value: settings.Colors,
				}, nil); err != nil {
					return nil, errors.WithMessage(err, "configuration/update")
				}
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
		_, wasOpen := h.openDocuments[params.TextDocument.URI]
		h.openDocuments[params.TextDocument.URI] = struct{}{}
		h.mu.Unlock()

		_ = conn.Notify(ctx, "textDocument/publishDecorations", cxp.TextDocumentPublishDecorationsParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: params.TextDocument.URI},
			Decorations:  createDecorations(settings.Colors),
		})

		if !wasOpen && settings.Animate {
			const startLine = 15
			sleep := 500 * time.Millisecond
		loop:
			for {
				select {
				case <-conn.DisconnectNotify():
					break loop
				case <-time.After(100 * time.Millisecond):
					for i := 0; i < 10; i++ {
						h.mu.Lock()
						colors := h.settings.Colors
						h.mu.Unlock()
						decorations := createDecorations(colors)
						var otherColor string
						if len(colors) > 0 {
							otherColor = colors[i%len(colors)]
						} else {
							otherColor = "red"
						}
						decorations = append(decorations, lspext.TextDocumentDecoration{Range: lsp.Range{Start: lsp.Position{Line: startLine + i}, End: lsp.Position{Line: startLine + i}},
							IsWholeLine:     true,
							BackgroundColor: otherColor,
						})
						_ = conn.Notify(ctx, "textDocument/publishDecorations", cxp.TextDocumentPublishDecorationsParams{
							TextDocument: lsp.TextDocumentIdentifier{URI: params.TextDocument.URI},
							Decorations:  decorations,
						})
						time.Sleep(sleep)
					}
				}
			}
		}
		return nil, nil

	case "textDocument/decoration":
		if len(openDocuments) > 0 {
			return nil, nil
		}
		return createDecorations(settings.Colors), nil
	}

	return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("method not supported: %s", req.Method)}
}

func publishDecorations(ctx context.Context, conn *jsonrpc2.Conn, openDocuments map[lsp.DocumentURI]struct{}, settings extensionSettings) error {
	for uri := range openDocuments {
		if err := conn.Notify(ctx, "textDocument/publishDecorations", cxp.TextDocumentPublishDecorationsParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: uri},
			Decorations:  createDecorations(settings.Colors),
		}); err != nil {
			return err
		}
	}
	return nil
}

func createDecorations(colors []string) []lspext.TextDocumentDecoration {
	if len(colors) == 0 {
		return []lspext.TextDocumentDecoration{}
	}
	maxLines := 10
	if len(colors) > maxLines {
		maxLines = len(colors)
	}
	decorations := make([]lspext.TextDocumentDecoration, maxLines)
	for i := 0; i < maxLines; i++ {
		decorations[i] = lspext.TextDocumentDecoration{
			Range:           lsp.Range{Start: lsp.Position{Line: i}, End: lsp.Position{Line: i}},
			IsWholeLine:     true,
			BackgroundColor: colors[i%len(colors)],
		}
	}
	return decorations
}

const (
	cycleColorsCommandID  = "sample-line-colors.cycle"
	promptColorsCommandID = "sample-line-colors.prompt"
)

var iconURL = "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(usFlagSVG))

const usFlagSVG = `<?xml version="1.0" encoding="UTF-8"?><svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="1235" height="650" viewBox="0 0 7410 3900"><rect width="7410" height="3900" fill="#b22234"/><path d="M0,450H7410m0,600H0m0,600H7410m0,600H0m0,600H7410m0,600H0" stroke="#fff" stroke-width="300"/><rect width="2964" height="2100" fill="#3c3b6e"/><g fill="#fff"><g id="s18"><g id="s9"><g id="s5"><g id="s4"><path id="s" d="M247,90 317.534230,307.082039 132.873218,172.917961H361.126782L176.465770,307.082039z"/><use xlink:href="#s" y="420"/><use xlink:href="#s" y="840"/><use xlink:href="#s" y="1260"/></g><use xlink:href="#s" y="1680"/></g><use xlink:href="#s4" x="247" y="210"/></g><use xlink:href="#s9" x="494"/></g><use xlink:href="#s18" x="988"/><use xlink:href="#s9" x="1976"/><use xlink:href="#s5" x="2470"/></g></svg>`

func cloneMap(m map[lsp.DocumentURI]struct{}) map[lsp.DocumentURI]struct{} {
	m2 := make(map[lsp.DocumentURI]struct{}, len(m))
	for k, v := range m {
		m2[k] = v
	}
	return m2
}
