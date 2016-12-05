package vfsutil

import (
	"context"
	"fmt"
	"os"
	pathpkg "path"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/neelance/parallel"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/sourcegraph/ctxvfs"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	lspext2 "github.com/sourcegraph/go-langserver/pkg/lspext"
	"github.com/sourcegraph/jsonrpc2"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/lspext"
)

// XRemoteFS is an implementation of ctxvfs.FileSystem that is backed by an
// LSP proxy server that implements the Files LSP extension
// (https://github.com/sourcegraph/language-server-protocol/pull/4).
//
// It caches the results of calls (both the initial workspace/xfiles call and
// all textDocument/xcontent calls) to avoid needless roundtrips. This assumes
// that the file system is immutable.
type XRemoteFS struct {
	Conn *jsonrpc2.Conn

	mu           sync.Mutex
	fileContents map[string]string

	once     sync.Once
	paths    sortedPaths
	pathsErr error
}

// call sends a request to the LSP proxy with tracing information.
func (fs *XRemoteFS) call(ctx context.Context, method, label string, params, result interface{}) (err error) {
	t0 := time.Now()
	span := opentracing.SpanFromContext(ctx)
	defer func() {
		span.LogEventWithPayload(method+": "+label, fmt.Sprintf("took %s, error: %v", time.Since(t0), err))
	}()
	return fs.Conn.Call(ctx, method, params, result, addTraceMeta(span))
}

func (fs *XRemoteFS) Open(ctx context.Context, path string) (ctxvfs.ReadSeekCloser, error) {
	fs.mu.Lock()
	if fs.fileContents == nil {
		fs.fileContents = map[string]string{}
	}
	text, ok := fs.fileContents[path]
	fs.mu.Unlock()
	if ok {
		return nopCloser{strings.NewReader(text)}, nil
	}

	// Precache all files in the requested path's directory to improve
	// performance across subsequent requests.
	dir := pathpkg.Dir(path)
	fis, err := fs.ReadDir(ctx, dir)
	if err != nil {
		return nil, err
	}
	par := parallel.NewRun(50)
	var resText string
	for i, fi := range fis {
		if fi.Mode().IsRegular() {
			fiPath := pathpkg.Join(dir, fi.Name())

			// Only block on prefetching a certain number, to avoid
			// slowing down on huge directories (we're assuming that
			// we won't need most of these files). For those files we
			// prefetch but don't block on, we will cache their
			// contents for subsequent operations. We might do some
			// duplicate work, but it's cheap.
			const maxPrefetch = 25
			block := i <= maxPrefetch || fiPath == path
			if block {
				par.Acquire()
			}

			go func(path2 string) {
				if block {
					defer par.Release()
				}
				text, err := fs.openSingleFile(ctx, path2)
				if path == path2 {
					resText = text
					if err != nil {
						par.Error(err)
					}
				}
			}(fiPath)
		}
	}
	if err := par.Wait(); err != nil {
		return nil, err
	}
	return nopCloser{strings.NewReader(resText)}, nil
}

func (fs *XRemoteFS) openSingleFile(ctx context.Context, path string) (string, error) {
	fs.mu.Lock()
	text, ok := fs.fileContents[path]
	fs.mu.Unlock()
	if !ok {
		params := lspext2.ContentParams{TextDocument: lsp.TextDocumentIdentifier{URI: "file://" + path}}
		var res lsp.TextDocumentItem
		if err := fs.call(ctx, "textDocument/xcontent", path, params, &res); err != nil {
			// TODO(sqs): cache error responses
			return "", err
		}
		text = res.Text

		fs.mu.Lock()
		fs.fileContents[path] = res.Text
		fs.mu.Unlock()
	}
	return text, nil
}

func (fs *XRemoteFS) fetchPaths(ctx context.Context) error {
	fs.once.Do(func() {
		params := lspext2.FilesParams{}
		var res []lsp.TextDocumentIdentifier
		fs.pathsErr = fs.call(ctx, "workspace/xfiles", "", &params, &res)
		if fs.pathsErr == nil {
			fs.paths = make(sortedPaths, len(res))
			for i, res := range res {
				fs.paths[i] = strings.TrimPrefix(res.URI, "file://")
			}
			sort.Strings(fs.paths)
		}
	})
	return fs.pathsErr
}

func (fs *XRemoteFS) Stat(ctx context.Context, path string) (os.FileInfo, error) {
	if path == "/" {
		return &lspext.FileInfo{Name_: "/", Dir_: true}, nil
	}

	if err := fs.fetchPaths(ctx); err != nil {
		return nil, err
	}

	return fs.paths.Stat(path)
}

func (fs *XRemoteFS) Lstat(ctx context.Context, path string) (os.FileInfo, error) {
	return fs.Stat(ctx, path)
}

func (fs *XRemoteFS) ReadDir(ctx context.Context, path string) ([]os.FileInfo, error) {
	if err := fs.fetchPaths(ctx); err != nil {
		return nil, err
	}

	return fs.paths.ReadDir(path)
}

func (fs *XRemoteFS) String() string {
	return "XRemoteFS"
}

type sortedPaths []string

func (paths sortedPaths) Stat(path string) (os.FileInfo, error) {
	if path == "/" {
		return &lspext.FileInfo{Name_: "/", Dir_: true}, nil
	}

	i := sort.SearchStrings(paths, path)
	if i >= len(paths) {
		return nil, &os.PathError{Op: "Stat", Path: path, Err: os.ErrNotExist}
	}
	if paths[i] == path {
		return &lspext.FileInfo{Name_: pathpkg.Base(path)}, nil
	}
	// '-' is before '/' so we need to check over all strings with path as
	// a prefix to try and find path + "/". This technically is O(N), but
	// in practice we expect to only do one iteration of this loop.
	for ; i < len(paths); i++ {
		if !strings.HasPrefix(paths[i], path) {
			break
		}
		if strings.HasPrefix(paths[i], path+"/") {
			return &lspext.FileInfo{Name_: pathpkg.Base(path), Dir_: true}, nil
		}
	}
	return nil, &os.PathError{Op: "Stat", Path: path, Err: os.ErrNotExist}
}

func (paths sortedPaths) ReadDir(path string) ([]os.FileInfo, error) {
	prefix := path
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	i := sort.SearchStrings(paths, prefix)
	if i >= len(paths) {
		return nil, &os.PathError{Op: "ReadDir", Path: path, Err: os.ErrNotExist}
	}
	if paths[i] == path {
		return nil, &os.PathError{Op: "ReadDir", Path: path, Err: os.ErrInvalid} // is file, not dir
	}

	var fis []os.FileInfo
	for _, path := range paths[i:] {
		if strings.HasPrefix(path, prefix) {
			rest := strings.TrimPrefix(path, prefix)

			var name string
			c := strings.Index(rest, "/")
			if c == -1 {
				name = rest
			} else {
				name = rest[:c]
			}

			if len(fis) == 0 || fis[len(fis)-1].Name() != name {
				fis = append(fis, &lspext.FileInfo{Name_: name, Dir_: c != -1})
			}
		} else {
			break
		}
	}
	if len(fis) == 0 {
		return nil, &os.PathError{Op: "ReadDir", Path: path, Err: os.ErrNotExist}
	}
	return fis, nil
}
