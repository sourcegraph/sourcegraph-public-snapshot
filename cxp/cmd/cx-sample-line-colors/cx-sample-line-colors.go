package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"log"
	"math/rand"
	"reflect"
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
	mu                      sync.Mutex
	initialized             bool
	settings                extensionSettings
	openDocuments           map[lsp.DocumentURI]struct{}
	registeredContributions bool
}

type extensionSettings struct {
	Colors []string `json:"lineColors.colors,omitempty"`
}

func (h *handler) handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
	h.mu.Lock()
	initialized := h.initialized
	settings := h.settings
	openDocuments := cloneMap(h.openDocuments)
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
						Commands: []string{
							randomizeColorsCommandID,
							promptColorsCommandID,
							clearCommandID,
						},
					},
				},
				DecorationProvider: &cxp.DecorationProviderServerCapabilities{
					DecorationCapabilityOptions: cxp.DecorationCapabilityOptions{Static: true, Dynamic: true},
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

		if err := publishDecorations(ctx, conn, openDocuments, params.Settings.Merged); err != nil {
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
		case randomizeColorsCommandID:
			settings.Colors = randomColors()
			if err := h.updateSettings(ctx, conn, openDocuments, settings); err != nil {
				return nil, errors.WithMessage(err, "update settings")
			}
			return nil, nil

		case promptColorsCommandID:
			var result *string
			if err := conn.Call(ctx, "window/showInput", cxp.ShowInputParams{
				Message:      "Enter line background colors (space-separated):",
				DefaultValue: strings.Join(settings.Colors, " "),
			}, &result); err != nil {
				return nil, errors.WithMessage(err, "window/showInput")
			}
			if result != nil {
				settings.Colors = strings.Fields(*result)
				if err := h.updateSettings(ctx, conn, openDocuments, settings); err != nil {
					return nil, errors.WithMessage(err, "update settings")
				}
			}
			return nil, nil

		case clearCommandID:
			settings.Colors = nil
			if err := h.updateSettings(ctx, conn, openDocuments, settings); err != nil {
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
			Decorations:  createDecorations(settings.Colors),
		})

		return nil, nil

	case "textDocument/decoration":
		if len(openDocuments) > 0 {
			return nil, nil
		}
		return createDecorations(settings.Colors), nil
	}

	return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("method not supported: %s", req.Method)}
}

func (h *handler) updateSettings(ctx context.Context, conn *jsonrpc2.Conn, openDocuments map[lsp.DocumentURI]struct{}, newSettings extensionSettings) error {
	if err := publishDecorations(ctx, conn, openDocuments, newSettings); err != nil {
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
			Path:  jsonx.MakePath("lineColors.colors"),
			Value: newSettings.Colors,
		}, nil); err != nil {
			log.Println("configuration/update error:", err)
		}
	}()
	return nil
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

func registerContributions(ctx context.Context, conn *jsonrpc2.Conn, settings extensionSettings, unregister bool) error {
	if err := conn.Call(ctx, "client/registerCapability", cxp.RegistrationParams{
		Registrations: []cxp.Registration{
			{
				ID:     "main",
				Method: "window/contribution",
				RegisterOptions: &cxp.Contributions{
					Commands: []*cxp.CommandContribution{
						{
							Command:  randomizeColorsCommandID,
							Title:    "Randomize theme",
							Category: "Line colors",
							ToolbarItem: &cxp.CommandContributionToolbarItem{
								Description:     "Randomize theme",
								IconURL:         iconURL(settings.Colors),
								IconDescription: fmt.Sprintf("vertical stripes of the following colors: %s", strings.Join(settings.Colors, " ")),
							},
						},
						{
							Command:  promptColorsCommandID,
							Title:    "Set theme...",
							Category: "Line colors",
						},
						{
							Command:  clearCommandID,
							Title:    "Clear",
							Category: "Line colors",
						},
					},
					Menus: &cxp.MenuContributions{
						EditorTitle: []*cxp.MenuItemContribution{
							{Command: randomizeColorsCommandID},
						},
						CommandPalette: []*cxp.MenuItemContribution{
							{Command: promptColorsCommandID},
							{Command: clearCommandID},
						},
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

const (
	randomizeColorsCommandID = "sample-line-colors.randomize"
	promptColorsCommandID    = "sample-line-colors.prompt"
	clearCommandID           = "sample-line-colors.clear"
)

func iconURL(colors []string) string {
	if len(colors) == 0 {
		colors = []string{"red", "orange", "yellow", "green", "blue", "indigo", "violet"}
	}

	svgRects := make([]string, len(colors))
	const totalWidth = float64(100)
	for i, color := range colors {
		width := totalWidth / float64(len(colors))
		svgRects[i] = fmt.Sprintf(`<rect x="%.2f" width="%.2f" height="100" fill="%s"/>`, width*float64(i), width, html.EscapeString(color))
		log.Println(color)
	}
	svg := `<?xml version="1.0" encoding="utf-8" standalone="yes"?><!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN" "http://www.w3.org/Graphics/SVG/1.1/DTD/svg11.dtd"><svg xmlns="http://www.w3.org/2000/svg" version="1.1" width="100" height="100">` + strings.Join(svgRects, "") + `</svg>`
	return "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(svg))
}

func cloneMap(m map[lsp.DocumentURI]struct{}) map[lsp.DocumentURI]struct{} {
	m2 := make(map[lsp.DocumentURI]struct{}, len(m))
	for k, v := range m {
		m2[k] = v
	}
	return m2
}

var assortedColors = []string{
	"hotpink",
	"red",
	"blue",
	"green",
	"purple",
	"orange",
	"violet",
	"cyan",
	"khaki",
	"lavender",
	"fuchsia",
	"limegreen",
	"royalblue",
	"yellow",
}

func randomColors() []string {
	count := 3 + rand.Intn(3)
	colors := make([]string, count)
	for i := range colors {
		colors[i] = assortedColors[rand.Intn(len(assortedColors))]
	}
	return colors
}
