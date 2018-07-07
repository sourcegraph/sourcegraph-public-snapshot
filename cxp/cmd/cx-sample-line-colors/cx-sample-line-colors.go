package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/sourcegraph/jsonx"
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
	settings extensionSettings
}

type extensionSettings struct {
	Colors []string `json:"colors,omitempty"`
}

func (h *handler) handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
	h.mu.Lock()
	settings := h.settings
	h.mu.Unlock()

	switch req.Method {
	case "initialize":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}

		cap, err := cxp.ParseExperimentalClientCapabilities(*req.Params)
		if err != nil {
			return nil, err
		}
		if !cap.Exec {
			return nil, errors.New("client does not support exec")
		}
		if !cap.Decorations {
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
		h.settings = settings
		h.mu.Unlock()

		return lsp.InitializeResult{
			Capabilities: lsp.ServerCapabilities{
				Experimental: cxp.ExperimentalServerCapabilities{
					DecorationsProvider: true,
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
						},
						Menus: &cxp.MenuContributions{
							EditorTitle: []*cxp.MenuItemContribution{{Command: cycleColorsCommandID}},
						},
					},
				},
			},
		}, nil

	case "shutdown", "exit":
		return nil, nil

	case "textDocument/decorations":
		if len(settings.Colors) == 0 {
			return []lspext.TextDocumentDecoration{}, nil
		}
		maxLines := 10
		if len(settings.Colors) > maxLines {
			maxLines = len(settings.Colors)
		}
		decorations := make([]lspext.TextDocumentDecoration, maxLines)
		for i := 0; i < maxLines; i++ {
			decorations[i] = lspext.TextDocumentDecoration{
				Range:           lsp.Range{Start: lsp.Position{Line: i}, End: lsp.Position{Line: i}},
				IsWholeLine:     true,
				BackgroundColor: settings.Colors[i%len(settings.Colors)],
			}
		}
		return decorations, nil
	}

	return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("method not supported: %s", req.Method)}
}

const cycleColorsCommandID = "sample-line-colors.cycle"

var iconURL = "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(usFlagSVG))

const usFlagSVG = `<?xml version="1.0" encoding="UTF-8"?><svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="1235" height="650" viewBox="0 0 7410 3900"><rect width="7410" height="3900" fill="#b22234"/><path d="M0,450H7410m0,600H0m0,600H7410m0,600H0m0,600H7410m0,600H0" stroke="#fff" stroke-width="300"/><rect width="2964" height="2100" fill="#3c3b6e"/><g fill="#fff"><g id="s18"><g id="s9"><g id="s5"><g id="s4"><path id="s" d="M247,90 317.534230,307.082039 132.873218,172.917961H361.126782L176.465770,307.082039z"/><use xlink:href="#s" y="420"/><use xlink:href="#s" y="840"/><use xlink:href="#s" y="1260"/></g><use xlink:href="#s" y="1680"/></g><use xlink:href="#s4" x="247" y="210"/></g><use xlink:href="#s9" x="494"/></g><use xlink:href="#s18" x="988"/><use xlink:href="#s9" x="1976"/><use xlink:href="#s5" x="2470"/></g></svg>`
