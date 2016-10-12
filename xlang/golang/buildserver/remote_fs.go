package buildserver

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/sourcegraph/ctxvfs"
	"github.com/sourcegraph/jsonrpc2"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/lspx"
)

// remoteProxyFS is an implementation of ctxvfs.FileSystem that
// communicates with the LSP proxy server over JSON-RPC to access the
// (virtual) file system.
//
// It caches the results of calls to avoid needless roundtrips. This
// assumes that the file system is immutable.
type remoteProxyFS struct {
	conn *jsonrpc2.Conn

	mu    sync.Mutex
	cache map[string]*fsPathCache
}

type fsPathCache struct {
	readFile    []byte
	readFileErr error
	stat        os.FileInfo
	statErr     error
	lstat       os.FileInfo
	lstatErr    error
	readDir     []os.FileInfo
	readDirErr  error
}

// call sends a request to the LSP proxy with tracing information.
func (fs *remoteProxyFS) call(ctx context.Context, method, path string, result interface{}) (err error) {
	t0 := time.Now()
	span := opentracing.SpanFromContext(ctx)
	defer func() {
		span.LogEventWithPayload(method+": "+path, fmt.Sprintf("took %s, error: %v", time.Since(t0), err))
	}()
	return fs.conn.Call(ctx, method, path, result, addTraceMeta(span))
}

func addTraceMeta(span opentracing.Span) jsonrpc2.CallOption {
	carrier := opentracing.TextMapCarrier{}
	if err := span.Tracer().Inject(span.Context(), opentracing.TextMap, carrier); err != nil {
		panic(err)
	}
	return jsonrpc2.Meta(carrier)
}

func (fs *remoteProxyFS) getCache(path string) *fsPathCache {
	if fs.cache == nil {
		fs.cache = map[string]*fsPathCache{}
	}
	c, ok := fs.cache[path]
	if !ok {
		c = &fsPathCache{}
		fs.cache[path] = c
	}
	return c
}

func (fs *remoteProxyFS) Open(ctx context.Context, path string) (ctxvfs.ReadSeekCloser, error) {
	fs.mu.Lock()
	c := fs.getCache(path)
	if c.readFile != nil {
		fs.mu.Unlock()
		return nopCloser{bytes.NewReader(c.readFile)}, nil
	} else if c.readFileErr != nil {
		fs.mu.Unlock()
		return nil, c.readFileErr
	}
	fs.mu.Unlock()

	var contents []byte
	if err := fs.call(ctx, "fs/readFile", path, &contents); err != nil {
		fs.mu.Lock()
		c.readFileErr = err
		fs.mu.Unlock()
		return nil, err
	}
	fs.mu.Lock()
	c.readFile = contents
	fs.mu.Unlock()
	return nopCloser{bytes.NewReader(contents)}, nil
}

func (fs *remoteProxyFS) Stat(ctx context.Context, path string) (fi os.FileInfo, err error) {
	fs.mu.Lock()
	c := fs.getCache(path)
	if c.stat != nil || c.statErr != nil {
		fs.mu.Unlock()
		return c.stat, c.statErr
	}
	fs.mu.Unlock()
	defer func() {
		fs.mu.Lock()
		c.stat = fi
		c.statErr = err
		fs.mu.Unlock()
	}()

	fi = &lspx.FileInfo{}
	return fi, fs.call(ctx, "fs/stat", path, &fi)
}

func (fs *remoteProxyFS) Lstat(ctx context.Context, path string) (fi os.FileInfo, err error) {
	fs.mu.Lock()
	c := fs.getCache(path)
	if c.lstat != nil || c.lstatErr != nil {
		fs.mu.Unlock()
		return c.lstat, c.lstatErr
	}
	fs.mu.Unlock()
	defer func() {
		fs.mu.Lock()
		c.lstat = fi
		c.lstatErr = err
		fs.mu.Unlock()
	}()

	fi = &lspx.FileInfo{}
	return fi, fs.call(ctx, "fs/lstat", path, &fi)
}

func (fs *remoteProxyFS) ReadDir(ctx context.Context, path string) (fis []os.FileInfo, err error) {
	fs.mu.Lock()
	c := fs.getCache(path)
	if c.readDir != nil || c.readDirErr != nil {
		fs.mu.Unlock()
		return c.readDir, c.readDirErr
	}
	fs.mu.Unlock()
	defer func() {
		fs.mu.Lock()
		c.readDir = fis
		c.readDirErr = err
		fs.mu.Unlock()
	}()

	var fis2 []lspx.FileInfo
	if err := fs.call(ctx, "fs/readDir", path, &fis2); err != nil {
		return nil, err
	}
	fis = make([]os.FileInfo, len(fis2))
	for i, fi := range fis2 {
		fis[i] = fi
	}
	return fis, nil
}

func (fs *remoteProxyFS) String() string {
	return "remoteProxyFS"
}

type nopCloser struct {
	io.ReadSeeker
}

func (nc nopCloser) Close() error { return nil }
