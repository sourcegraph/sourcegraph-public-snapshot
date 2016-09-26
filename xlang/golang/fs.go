package golang

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	pathpkg "path"
	"strings"
	"time"

	"golang.org/x/tools/godoc/vfs"

	opentracing "github.com/opentracing/opentracing-go"

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

func (h *handlerShared) readFile(uri string) ([]byte, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	contents, err := vfs.ReadFile(h.fs, h.filePath(uri))
	if os.IsNotExist(err) {
		if _, ok := err.(*os.PathError); !ok {
			err = &os.PathError{Op: "Open", Path: h.filePath(uri), Err: err}
		}
	}
	return contents, err
}

func (h *handlerShared) addOverlayFile(uri string, contents []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	path := h.filePath(uri)
	path = "/" + pathTrimPrefix(path, h.overlayMountPath)
	h.overlayFS[path] = contents
}

func (h *handlerShared) removeOverlayFile(uri string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.overlayFS, h.filePath(uri))
}

func (h *handlerShared) readOverlayFile(uri string) (contents []byte, found bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	contents, found = h.overlayFS[h.filePath(uri)]
	return
}

// namespaceFS makes Stat and Lstat produce results when called on an
// ancestor of a mount point. For example, if you've mounted at /x/y/z
// in a vfs.NameSpace and call Stat("/x/y"), it will return
// os.ErrNotExist. This wrapper type makes it return a directory
// FileInfo.
type namespaceFS struct{ vfs.NameSpace }

func (fs namespaceFS) Stat(path string) (os.FileInfo, error) {
	fi, err := fs.NameSpace.Stat(path)
	if !os.IsNotExist(err) {
		return fi, err
	}

	_, err = fs.NameSpace.ReadDir(path)
	if err == nil {
		return dirInfo{}, nil
	}
	return nil, err
}

func (fs namespaceFS) Lstat(path string) (os.FileInfo, error) { return fs.Stat(path) }

type dirInfo struct{ name string }

func (e dirInfo) Name() string { return e.name }

func (e dirInfo) Size() int64 { return 0 }

func (e dirInfo) Mode() os.FileMode { return os.ModeDir | os.ModePerm }

func (e dirInfo) ModTime() time.Time { return time.Time{} }

func (e dirInfo) IsDir() bool { return true }

func (e dirInfo) Sys() interface{} { return nil }

func newWithFileOverlaid(fs vfs.FileSystem, path string, contents []byte) vfs.FileSystem {
	return &withFileOverlaid{
		FileSystem: fs,
		path:       pathpkg.Clean(path),
		name:       pathpkg.Base(path),
		pathDir:    pathpkg.Dir(path),
		contents:   contents,
	}
}

// withFileOverlaid wraps a VFS and adds a single file.
type withFileOverlaid struct {
	vfs.FileSystem
	path     string
	name     string
	pathDir  string
	contents []byte
}

func (fs *withFileOverlaid) Open(path string) (vfs.ReadSeekCloser, error) {
	if path == fs.path {
		return nopCloser{bytes.NewReader(fs.contents)}, nil
	}
	return fs.FileSystem.Open(path)
}

func (fs *withFileOverlaid) Stat(path string) (os.FileInfo, error) {
	if path == fs.path {
		return fileInfo{fs.name, int64(len(fs.contents))}, nil
	}
	return fs.FileSystem.Stat(path)
}

func (fs *withFileOverlaid) ReadDir(path string) ([]os.FileInfo, error) {
	fis, err := fs.FileSystem.ReadDir(path)
	if err == nil && path == pathpkg.Dir(fs.path) {
		fis = append(fis, fileInfo{fs.name, int64(len(fs.contents))})
	}
	return fis, err
}

func (fs *withFileOverlaid) Lstat(path string) (os.FileInfo, error) { return fs.Stat(path) }

type fileInfo struct {
	name string
	size int64
}

func (e fileInfo) Name() string { return e.name }

func (e fileInfo) Size() int64 { return e.size }

func (e fileInfo) Mode() os.FileMode { return os.ModePerm }

func (e fileInfo) ModTime() time.Time { return time.Time{} }

func (e fileInfo) IsDir() bool { return false }

func (e fileInfo) Sys() interface{} { return nil }

type nopCloser struct {
	io.ReadSeeker
}

func (nc nopCloser) Close() error { return nil }
