// Command cx-blame is a Sourcegraph extension that decorates text documents with Git blame
// information.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"reflect"
	"strings"
	"sync"

	humanize "github.com/dustin/go-humanize"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
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
	cxpmain.Main("cx-blame", func() jsonrpc2.Handler { return jsonrpc2.AsyncHandler(jsonrpc2.HandlerWithError((&handler{}).handle)) })
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
		if !cap.Exec {
			return nil, errors.New("client does not support exec")
		}
		if cap.Decoration == nil || !cap.Decoration.Dynamic {
			return nil, errors.New("client does not support published decorations")
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
		h.initialized = true
		h.rootURI = rootURI
		h.settings = settings
		h.openDocuments = map[lsp.DocumentURI]struct{}{}
		h.mu.Unlock()

		return cxp.InitializeResult{
			Capabilities: cxp.ServerCapabilities{
				ServerCapabilities: lsp.ServerCapabilities{
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

		decorations, err := createDecorations(ctx, conn, *h.rootURI, params.TextDocument.URI, settings)
		if err != nil {
			return nil, err
		}
		_ = conn.Notify(ctx, "textDocument/publishDecorations", cxp.TextDocumentPublishDecorationsParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: params.TextDocument.URI},
			Decorations:  decorations,
		})
		return nil, nil
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
		decorations, err := createDecorations(ctx, conn, *h.rootURI, uri, settings)
		if err != nil {
			return err
		}
		if err := conn.Notify(ctx, "textDocument/publishDecorations", cxp.TextDocumentPublishDecorationsParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: uri},
			Decorations:  decorations,
		}); err != nil {
			return err
		}
	}
	return nil
}

func createDecorations(ctx context.Context, conn *jsonrpc2.Conn, root uri.URI, document lsp.DocumentURI, settings extensionSettings) ([]lspext.TextDocumentDecoration, error) {
	decorations := []lspext.TextDocumentDecoration{}
	if settings.Hide {
		return decorations, nil
	}

	uri, err := url.Parse(string(document))
	if err != nil {
		return nil, err
	}
	path := strings.TrimPrefix(uri.Path, "/")

	hunks, err := git.BlameFileCmd(ctx, cxp.ExecCmdFunc("git", conn), path, &git.BlameOptions{
		NewestCommit: api.CommitID(root.Rev()),
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

	decorations = make([]lspext.TextDocumentDecoration, len(hunks))
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
				LinkURL: fmt.Sprintf("/%s/-/commit/%s", root.Repo(), hunk.CommitID),
			},
		}
	}
	return decorations, nil
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
							Title:   "Blame",
							Detail:  showHide + " blame annotations",
						},
					},
					Menus: &cxp.MenuContributions{
						CommandPalette: []*cxp.MenuItemContribution{{Command: toggleCommandID}},
						EditorTitle:    []*cxp.MenuItemContribution{{Command: toggleCommandID}},
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

const toggleCommandID = "blame.toggle"

func cloneMap(m map[lsp.DocumentURI]struct{}) map[lsp.DocumentURI]struct{} {
	m2 := make(map[lsp.DocumentURI]struct{}, len(m))
	for k, v := range m {
		m2[k] = v
	}
	return m2
}
