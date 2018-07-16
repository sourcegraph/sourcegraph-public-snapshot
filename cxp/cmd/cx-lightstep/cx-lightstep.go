// Command cx-lightstep is a Sourcegraph extension that decorates text documents with links to the
// relevant LightStep trace view.
//
// It is very primitive now. It just generates LightStep search URLs when it sees a line containing
// "StartSpan" and a string literal (which is used as the query term).
package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"regexp"
	"strings"
	"sync"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/sourcegraph/jsonx"
	"github.com/sourcegraph/sourcegraph/cxp"
	"github.com/sourcegraph/sourcegraph/cxp/pkg/cxpmain"
	"github.com/sourcegraph/sourcegraph/xlang/lspext"
	"github.com/sourcegraph/sourcegraph/xlang/vfsutil"
)

//docker:user sourcegraph

func main() {
	cxpmain.Main("cx-lightstep", func() jsonrpc2.Handler { return jsonrpc2.AsyncHandler(jsonrpc2.HandlerWithError((&handler{}).handle)) })
}

type handler struct {
	mu       sync.Mutex
	settings extensionSettings
}

type extensionSettings struct {
	SpanLinks *bool  `json:"spanLinks,omitempty"`
	Project   string `json:"project,omitempty"`
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
		h.mu.Lock()
		h.settings = settings
		h.mu.Unlock()

		var showHide string
		if settings.SpanLinks == nil || *settings.SpanLinks {
			showHide = "Hide"
		} else {
			showHide = "Show"
		}

		contributions := &cxp.Contributions{
			Commands: []*cxp.CommandContribution{
				{
					Command: toggleSpanLinksCommandID,
					Title:   showHide + " OpenTracing span links",
					IconURL: iconURL,
					ExperimentalSettingsAction: &cxp.CommandContributionSettingsAction{
						Path:        jsonx.PropertyPath("spanLinks"),
						CycleValues: []interface{}{true, false},
					},
				},
				{
					Command: setProjectCommandID,
					Title:   "Set LightStep project name",
					IconURL: iconURL,
					ExperimentalSettingsAction: &cxp.CommandContributionSettingsAction{
						Path:   jsonx.PropertyPath("project"),
						Prompt: "Set LightStep project name (example: mycompany-prod)",
					},
				},
			},
			Menus: &cxp.MenuContributions{
				EditorTitle: []*cxp.MenuItemContribution{
					// Show only 1 of these commands, depending on whether the project name is set.
					{Command: toggleSpanLinksCommandID, Hidden: settings.Project == ""},
					{Command: setProjectCommandID, Hidden: settings.Project != ""},
				},
			},
		}

		return cxp.InitializeResult{
			Capabilities: cxp.ServerCapabilities{
				DecorationProvider: &cxp.DecorationProviderServerCapabilities{DecorationCapabilityOptions: cxp.DecorationCapabilityOptions{Static: true}},
				Contributions:      contributions,
			},
		}, nil

	case "shutdown", "exit":
		return nil, nil

	case "textDocument/decoration":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params lspext.TextDocumentDecorationParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}

		uri, err := url.Parse(string(params.TextDocument.URI))
		if err != nil {
			return nil, err
		}
		path := strings.TrimPrefix(uri.Path, "/")

		if (settings.SpanLinks != nil && !*settings.SpanLinks) && settings.Project != "" {
			return []lspext.TextDocumentDecoration{}, nil
		}

		fs := vfsutil.XRemoteFS{Conn: conn}
		f, err := fs.Open(ctx, "/"+path)
		if err != nil {
			return nil, err
		}
		data, err := ioutil.ReadAll(f)
		if err != nil {
			return nil, err
		}

		spans := findSpans(data)
		decorations := make([]lspext.TextDocumentDecoration, len(spans))
		for i, span := range spans {
			d := lspext.TextDocumentDecoration{
				Range: lsp.Range{
					Start: lsp.Position{Line: span.line},
					End:   lsp.Position{Line: span.line},
				},
				After: &lspext.DecorationAttachmentRenderOptions{
					Color:           "rgba(255, 255, 255, 0.8)",
					BackgroundColor: "#2925ff", // LightStep brand color
				},
			}
			if settings.Project != "" {
				d.After.ContentText = "Live traces (LightStep) » "
				d.After.LinkURL = fmt.Sprintf("https://app.lightstep.com/%s/live?q=%s",
					settings.Project,
					url.QueryEscape(fmt.Sprintf("operation:%q", span.query)),
				)
			} else {
				d.After.ContentText = "Configure LightStep..."
				d.After.HoverMessage = "Press the LightStep icon in the title bar to set your project name."
			}
			decorations[i] = d
		}
		return decorations, nil
	}

	return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("method not supported: %s", req.Method)}
}

type span struct {
	line  int // 0-indexed
	query string
}

var spanPattern = regexp.MustCompile(`(?i)start_?span[^"']+["']([^"']+)["']`)

func findSpans(text []byte) []span {
	var found []span
	for i, line := range bytes.Split(text, []byte("\n")) {
		for _, m := range spanPattern.FindAllSubmatch(line, -1) {
			found = append(found, span{line: i, query: string(m[1])})
		}
	}
	return found
}

const (
	toggleSpanLinksCommandID = "lightstep.spanLinks.toggle"
	setProjectCommandID      = "lightstep.project.set"
)

var iconURL = "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(lightstepLogoSVG))

const lightstepLogoSVG = `<?xml version="1.0" encoding="UTF-8"?>
<svg
   xmlns="http://www.w3.org/2000/svg"
   viewBox="0 0 181.60001 101.6"
   height="101.6"
   width="181.60001">
  <g
     transform="translate(-0.8,-0.1)"
     style="fill:none;fill-rule:evenodd;stroke:none;stroke-width:1"
     id="Page-1">
    <g
       id="logo">
      <g
         id="Group">
        <g
           style="fill-rule:nonzero"
           transform="translate(203,20)"
           id="right-side_light_text_1_" />
        <g
           style="fill:#2d36fb"
           id="dark_logo">
          <path
             id="path4592"
             d="m 28.8,101.7 h 52.3 c 15.6,0 27,-9 35.4,-21.2 5,-7.3 10.2,-13.8 19.9,-13.8 H 168 V 50.1 h -43.8 c -9.7,0 -14.8,6.5 -19.9,13.8 C 95.9,76.1 84.6,85.1 68.9,85.1 H 28.8 Z" />
          <path
             id="path4594"
             d="m 0.8,79.3 h 64.8 c 15.6,0 27,-9 35.4,-21.2 5,-7.3 10.2,-13.8 19.9,-13.8 h 61.5 V 23.6 h -61.5 c -19.1,0 -24.8,8.9 -32.7,19.3 -5.8,7.7 -12,15.7 -22.6,15.7 H 0.8 Z" />
          <path
             id="path4596"
             d="M 16.4,53.1 H 61.5 C 72.1,53.1 78.3,45 84.1,37.4 92,27 97.7,18.1 116.8,18.1 h 34.5 V 0.1 H 108 C 88.9,0.1 83.2,9 75.3,19.4 69.5,27.1 63.3,35.1 52.7,35.1 H 16.4 Z" />
        </g>
      </g>
    </g>
  </g>
</svg>`
