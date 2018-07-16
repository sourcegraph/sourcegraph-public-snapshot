// Command cx-line-age is a Sourcegraph extension that decorates text documents based on the age of
// each line (since its last commit).
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/url"
	"strings"
	"sync"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/sourcegraph/jsonx"
	"github.com/sourcegraph/sourcegraph/cxp"
	"github.com/sourcegraph/sourcegraph/cxp/pkg/cxpmain"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
	"github.com/sourcegraph/sourcegraph/xlang/lspext"
	"github.com/sourcegraph/sourcegraph/xlang/uri"
)

//docker:user sourcegraph

func main() {
	cxpmain.Main("cx-line-age", func() jsonrpc2.Handler { return jsonrpc2.AsyncHandler(jsonrpc2.HandlerWithError((&handler{}).handle)) })
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
				DecorationProvider: &cxp.DecorationProviderServerCapabilities{DecorationCapabilityOptions: cxp.DecorationCapabilityOptions{Static: true}},
				Contributions: &cxp.Contributions{
					Commands: []*cxp.CommandContribution{
						{
							Command: toggleCommandID,
							Title:   "Line age",
							ExperimentalSettingsAction: &cxp.CommandContributionSettingsAction{
								Path:        jsonx.PropertyPath("hide"),
								CycleValues: []interface{}{false, true},
							},
						},
					},
					Menus: &cxp.MenuContributions{
						EditorTitle: []*cxp.MenuItemContribution{{Command: toggleCommandID}},
					},
				},
			},
		}, nil

	case "shutdown", "exit":
		return nil, nil

	case "textDocument/decorations":
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
		path := strings.TrimPrefix(uri.Path, "/")

		hunks, err := git.BlameFileCmd(ctx, cxp.ExecCmdFunc("git", conn), path, &git.BlameOptions{
			NewestCommit: api.CommitID(rootURI.Rev()),
		})
		if err != nil {
			return nil, err
		}

		var (
			now            = time.Now()
			oldest, newest time.Time
			maxLine        = 0
		)
		for _, hunk := range hunks {
			if hunk.Author.Date.Before(oldest) {
				oldest = hunk.Author.Date
			}
			if hunk.Author.Date.After(newest) {
				newest = hunk.Author.Date
			}
			if hunk.EndLine > maxLine {
				maxLine = hunk.EndLine
			}
		}
		oldestD := now.Sub(oldest)
		newestD := now.Sub(newest)

		// Don't consider any code older than 30 days "completely new".
		if notNew := 60 * 24 * time.Hour; newestD > notNew {
			newestD = notNew
		}

		lineDecorations := make([]lspext.TextDocumentDecoration, maxLine)
		for _, hunk := range hunks {
			for line := hunk.StartLine; line < hunk.EndLine; line++ {
				lineDecorations[line-1] = lspext.TextDocumentDecoration{
					BackgroundColor: colorForAge(now.Sub(hunk.Author.Date), oldestD, newestD),
					Range: lsp.Range{
						Start: lsp.Position{Line: line - 1},
						End:   lsp.Position{Line: line - 1},
					},
					IsWholeLine: true,
				}
			}
		}
		return lineDecorations, nil
	}

	return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("method not supported: %s", req.Method)}
}

func colorForAge(age, oldest, newest time.Duration) string {
	// Exponential decay with half-life of 30 days.
	const λ = float64(30 * 24 * time.Hour)
	agef := λ / (float64(age) + λ)
	oldestf := λ / (float64(oldest) + λ)
	newestf := λ / (float64(newest) + λ)

	// Normalize linearly to [0, 1].
	x := (agef - oldestf) / (newestf - oldestf)

	// Older code is red, newer code is green.
	redHue, greenHue := 0, 116

	// Higher alpha at extremes (oldest and newest).
	alpha := math.Abs(x-0.5) + 0.1

	return fmt.Sprintf("hsla(%d, 100%%, 50%%, %.2f)", redHue+int(x*float64(greenHue-redHue)), alpha)
}

const toggleCommandID = "line-age.toggle"
