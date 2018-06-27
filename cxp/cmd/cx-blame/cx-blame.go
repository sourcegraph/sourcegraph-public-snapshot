// Command cx-blame is a Sourcegraph extension that decorates text documents with Git blame
// information.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"

	humanize "github.com/dustin/go-humanize"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/sourcegraph/sourcegraph/cxp"
	"github.com/sourcegraph/sourcegraph/cxp/pkg/cxpmain"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
	"github.com/sourcegraph/sourcegraph/xlang/lspext"
	"github.com/sourcegraph/sourcegraph/xlang/uri"
)

//docker:user sourcegraph

func main() {
	cxpmain.Main("cx-blame", func() jsonrpc2.Handler { return jsonrpc2.AsyncHandler(jsonrpc2.HandlerWithError((&handler{}).handle)) })
}

type handler struct {
	mu      sync.Mutex
	rootURI *uri.URI
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

		var params lspext.InitializeParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		rootURI, err := uri.Parse(string(params.OriginalRootURI))
		if err != nil {
			return nil, err
		}
		h.mu.Lock()
		h.rootURI = rootURI
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
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params lspext.TextDocumentDecorationsParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
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

		truncate := func(max int, str, omission string) string {
			if len(str) > max {
				return str[:max] + omission
			}
			return str
		}

		decorations := make([]lspext.TextDocumentDecoration, len(hunks))
		for i, hunk := range hunks {
			decorations[i] = lspext.TextDocumentDecoration{
				// Alternate design: add a border at the top of each hunk.
				//
				// Background: "linear-gradient(to bottom, var(--info) -20%, transparent 2px)",
				// Border:      "solid",
				// BorderWidth: "1px 0 0 0",
				// BorderColor: "var(--info)",
				Range: lsp.Range{
					Start: lsp.Position{Line: hunk.StartLine - 1},
					End:   lsp.Position{Line: hunk.EndLine - 1},
				},
				IsWholeLine: true,
				After: &lspext.DecorationAttachmentRenderOptions{
					ContentText: fmt.Sprintf("%s, %s • %s %s",
						hunk.Author.Name,
						humanize.Time(hunk.Author.Date),
						truncate(80, hunk.Message, "…"),
						truncate(7, string(hunk.CommitID), ""),
					),
					// Alternate design: show the blame as white on teal (much more noticeable, too
					// distracting to keep always enabled).
					//
					// Color:           "white",
					// BackgroundColor: "var(--info)",
					Color:           "var(--text-muted)",
					BackgroundColor: "rgba(127, 127, 127, 0.1)",
					// TODO(extensions): Find a way to not need to hardcode our URL structure.
					LinkURL: fmt.Sprintf("/%s/-/commit/%s", rootURI.Repo(), hunk.CommitID),
				},
			}
		}
		return decorations, nil
	}

	return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("method not supported: %s", req.Method)}
}
