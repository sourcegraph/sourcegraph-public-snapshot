package langserver

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	opentracing "github.com/opentracing/opentracing-go"

	"github.com/sourcegraph/ctxvfs"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

// IsFileSystemRequest returns if this is an LSP method whose sole
// purpose is modifying the contents of the overlay file system.
func IsFileSystemRequest(method string) bool {
	return method == "textDocument/didOpen" ||
		method == "textDocument/didChange" ||
		method == "textDocument/didClose" ||
		method == "textDocument/didSave"
}

func (h *HandlerShared) HandleFileSystemRequest(ctx context.Context, req *jsonrpc2.Request) error {
	span := opentracing.SpanFromContext(ctx)

	switch req.Method {
	case "textDocument/didOpen":
		var params lsp.DidOpenTextDocumentParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return err
		}
		span.SetTag("uri", params.TextDocument.URI)
		h.addOverlayFile(params.TextDocument.URI, []byte(params.TextDocument.Text))
		return nil

	case "textDocument/didChange":
		var params lsp.DidChangeTextDocumentParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return err
		}
		contents, found := h.readOverlayFile(params.TextDocument.URI)
		if !found {
			return fmt.Errorf("received textDocument/didChange for unknown file %q", params.TextDocument.URI)
		}
		for _, change := range params.ContentChanges {
			switch {
			case change.Range == nil && change.RangeLength == 0:
				contents = []byte(change.Text) // new full content

			default:
				return fmt.Errorf("incremental updates in textDocument/didChange not supported for file %q", params.TextDocument.URI)
			}
		}
		h.addOverlayFile(params.TextDocument.URI, contents)
		return nil

	case "textDocument/didClose":
		var params lsp.DidCloseTextDocumentParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return err
		}
		h.removeOverlayFile(params.TextDocument.URI)
		return nil

	case "textDocument/didSave":
		// no-op
		return nil

	default:
		panic("unexpected file system request method: " + req.Method)
	}
}

func (h *HandlerShared) FilePath(uri string) string {
	path := strings.TrimPrefix(uri, "file://")
	if !strings.HasPrefix(path, "/") {
		panic(fmt.Sprintf("bad uri %q (path %q MUST have leading slash; it can't be relative)", uri, path))
	}
	if strings.Contains(path, ":") {
		panic(fmt.Sprintf("bad uri %q (path %q MUST NOT contain ':')", uri, path))
	}
	if strings.Contains(path, "@") {
		panic(fmt.Sprintf("bad uri %q (path %q MUST NOT contain '@')", uri, path))
	}
	return path
}

func (h *HandlerShared) readFile(ctx context.Context, uri string) ([]byte, error) {
	h.Mu.Lock()
	fs := h.FS
	path := h.FilePath(uri)
	h.Mu.Unlock()
	contents, err := ctxvfs.ReadFile(ctx, fs, path)
	if os.IsNotExist(err) {
		if _, ok := err.(*os.PathError); !ok {
			err = &os.PathError{Op: "Open", Path: path, Err: err}
		}
	}
	return contents, err
}

func uriToOverlayPath(uri string) string {
	return strings.TrimPrefix(uri, "file:///")
}

func (h *HandlerShared) addOverlayFile(uri string, contents []byte) {
	h.Mu.Lock()
	defer h.Mu.Unlock()
	h.overlayFSMu.Lock()
	defer h.overlayFSMu.Unlock()
	h.overlayFS[uriToOverlayPath(uri)] = contents
}

func (h *HandlerShared) removeOverlayFile(uri string) {
	h.Mu.Lock()
	defer h.Mu.Unlock()
	h.overlayFSMu.Lock()
	defer h.overlayFSMu.Unlock()
	delete(h.overlayFS, uriToOverlayPath(uri))
}

func (h *HandlerShared) readOverlayFile(uri string) (contents []byte, found bool) {
	h.Mu.Lock()
	defer h.Mu.Unlock()
	h.overlayFSMu.Lock()
	defer h.overlayFSMu.Unlock()
	contents, found = h.overlayFS[uriToOverlayPath(uri)]
	return
}
