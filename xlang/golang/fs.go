package golang

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	opentracing "github.com/opentracing/opentracing-go"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/ctxvfs"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/jsonrpc2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/lsp"
)

// isFileSystemRequest returns if this is an LSP method whose sole
// purpose is modifying the contents of the overlay file system.
func isFileSystemRequest(method string) bool {
	return method == "textDocument/didOpen" ||
		method == "textDocument/didChange" ||
		method == "textDocument/didClose"
}

func (h *handlerShared) handleFileSystemRequest(ctx context.Context, req *jsonrpc2.Request) error {
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

	default:
		panic("unexpected file system request method: " + req.Method)
	}
}

func (h *handlerShared) filePath(uri string) string {
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

func (h *handlerShared) readFile(ctx context.Context, uri string) ([]byte, error) {
	h.mu.Lock()
	fs := h.fs
	path := h.filePath(uri)
	h.mu.Unlock()
	contents, err := ctxvfs.ReadFile(ctx, fs, path)
	if os.IsNotExist(err) {
		if _, ok := err.(*os.PathError); !ok {
			err = &os.PathError{Op: "Open", Path: path, Err: err}
		}
	}
	return contents, err
}

func (h *handlerShared) addOverlayFile(uri string, contents []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.overlayFSMu.Lock()
	defer h.overlayFSMu.Unlock()
	path := h.filePath(uri)
	path = pathTrimPrefix(path, h.overlayMountPath)
	h.overlayFS[path] = contents
}

func (h *handlerShared) removeOverlayFile(uri string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.overlayFSMu.Lock()
	defer h.overlayFSMu.Unlock()
	delete(h.overlayFS, h.filePath(uri))
}

func (h *handlerShared) readOverlayFile(uri string) (contents []byte, found bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.overlayFSMu.Lock()
	defer h.overlayFSMu.Unlock()
	contents, found = h.overlayFS[h.filePath(uri)]
	return
}
