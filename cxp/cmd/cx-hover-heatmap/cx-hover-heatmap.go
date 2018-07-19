package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/sourcegraph/jsonx"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"

	"github.com/sourcegraph/sourcegraph/xlang/lspext"
	"github.com/sourcegraph/sourcegraph/xlang/uri"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/sourcegraph/cxp"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/sourcegraph/sourcegraph/cxp/pkg/cxpmain"
)

// Global in-memory store of line of code to number of hovers
// Key: unique file identifier repoPath/filePath@revision
// Value: map with line number -> number of hovers
var store = make(map[string]map[int]int)

//docker:user sourcegraph

func main() {
	cxpmain.Main("cxp-hover-heatmap", func() jsonrpc2.Handler { return jsonrpc2.AsyncHandler(jsonrpc2.HandlerWithError((&handler{}).handle)) })
}

type handler struct {
	mu       sync.Mutex
	rootURI  *uri.URI
	settings extensionSettings
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
	rootURI := h.rootURI
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
		if !cap.Exec {
			return nil, errors.New("client does not support exec")
		}
		if cap.Decoration == nil || !cap.Decoration.Static {
			return nil, errors.New("client does not support decorations")
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
		h.rootURI = rootURI
		h.settings = settings
		h.mu.Unlock()

		return cxp.InitializeResult{
			Capabilities: cxp.ServerCapabilities{
				ServerCapabilities: lsp.ServerCapabilities{
					HoverProvider: true,
				},
				DecorationProvider: &cxp.DecorationProviderServerCapabilities{
					DecorationCapabilityOptions: cxp.DecorationCapabilityOptions{Dynamic: true},
				},
				Contributions: &cxp.Contributions{
					Commands: []*cxp.CommandContribution{
						{
							Command: toggleHoverHeatmapID,
							Title:   "Toggle hover heatmap",
							ExperimentalSettingsAction: &cxp.CommandContributionSettingsAction{
								Path:        jsonx.PropertyPath("hide"),
								CycleValues: []interface{}{false, true},
							},
						},
					},
					Menus: &cxp.MenuContributions{
						EditorTitle: []*cxp.MenuItemContribution{{Command: toggleHoverHeatmapID}},
					},
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
		uri, err := url.Parse(string(params.TextDocument.URI))
		if err != nil {
			return nil, err
		}
		// Construct a unique identifier for each line of code. repoPath/filePath@revision
		rawPath := fmt.Sprintf("%s%s@%s", string(h.rootURI.Repo()), uri.EscapedPath(), h.rootURI.Rev())
		line := params.Position.Line
		if store[rawPath] != nil {
			store[rawPath][line] = store[rawPath][line] + 1
		} else {
			store[rawPath] = map[int]int{}
			store[rawPath][line] = 1
		}

		return lsp.Hover{
			Contents: []lsp.MarkedString{},
		}, nil

	case "textDocument/decoration":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params lspext.TextDocumentDecorationParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}

		if settings.Hide {
			return []lspext.TextDocumentDecoration{}, nil
		}

		uri, err := url.Parse(string(params.TextDocument.URI))
		if err != nil {
			return nil, err
		}

		rawPath := fmt.Sprintf("%s%s@%s", string(h.rootURI.Repo()), uri.EscapedPath(), h.rootURI.Rev())
		path := strings.TrimPrefix(uri.Path, "/")

		hunks, err := git.BlameFileCmd(ctx, cxp.ExecCmdFunc("git", conn), path, &git.BlameOptions{
			NewestCommit: api.CommitID(rootURI.Rev()),
		})
		if err != nil {
			return nil, err
		}

		var (
			highest    int
			lineCounts = map[int]int{}
		)

		lineDecorations := make([]lspext.TextDocumentDecoration, hunks[len(hunks)-1].EndLine)
		for _, hunk := range hunks {
			for line := hunk.StartLine; line < hunk.EndLine; line++ {
				count := store[rawPath][line-1]
				// relativeFrequency := getRelativeFrequency(rawPath, line, total)
				if count > highest {
					highest = count
				}
				lineCounts[line] = count
			}
		}

		for line, count := range lineCounts {

			lineDecorations[line-1] = lspext.TextDocumentDecoration{
				BackgroundColor: getColorByRelativeFrequency(count, highest),
				Range: lsp.Range{
					Start: lsp.Position{Line: line - 1},
					End:   lsp.Position{Line: line - 1},
				},
				IsWholeLine: true,
			}
		}

		return lineDecorations, nil

	}

	return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("method not supported: %s", req.Method)}
}

// Determines the background color for a given line.
// `highest` is the highest number of hovers that a line in this file has.
// We calculate the line color by the relative frequency of hovers, with `highest`
// being the denominator, meaning `highest` will always be the darkest red in a file.
func getColorByRelativeFrequency(count, highest int) string {
	redHue, greenHue := 0, 116
	x := float64(count)
	if highest > 0 {
		x = float64(count) / float64(highest)
	}

	const alpha = 1.0
	return fmt.Sprintf("hsla(%d, 100%%, 50%%, %.2f)", greenHue-int(x*float64(greenHue-redHue)), alpha)
}

const toggleHoverHeatmapID = "hover-heatmap.toggle"
