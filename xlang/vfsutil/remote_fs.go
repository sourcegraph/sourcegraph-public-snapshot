package vfsutil

import (
	"bytes"
	"context"
	"io"
	"os"
	"strconv"
	"sync"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/sourcegraph/ctxvfs"
	"github.com/sourcegraph/jsonrpc2"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/lspext"
)

// RemoteFS selects the currently active remote FS protocol (based on
// env vars) and returns it for conn.
func RemoteFS(conn *jsonrpc2.Conn) ctxvfs.FileSystem {
	// Temporary experiment. Soon it will always be XRemoteFS
	if v, _ := strconv.ParseBool(os.Getenv("X_REMOTE_FS_ENABLE")); v {
		return &XRemoteFS{Conn: conn}
	}
	return &RemoteProxyFS{Conn: conn}
}

// RemoteProxyFS is an implementation of ctxvfs.FileSystem that
// communicates with the LSP proxy server over JSON-RPC to access the
// (virtual) file system.
//
// It caches the results of calls to avoid needless roundtrips. This
// assumes that the file system is immutable.
type RemoteProxyFS struct {
	Conn *jsonrpc2.Conn

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
func (fs *RemoteProxyFS) call(ctx context.Context, method, path string, result interface{}) (err error) {
	span := opentracing.SpanFromContext(ctx)
	return fs.Conn.Call(ctx, method, path, result, addTraceMeta(span))
}

func addTraceMeta(span opentracing.Span) jsonrpc2.CallOption {
	carrier := opentracing.TextMapCarrier{}
	if err := span.Tracer().Inject(span.Context(), opentracing.TextMap, carrier); err != nil {
		panic(err)
	}
	return jsonrpc2.Meta(carrier)
}

func (fs *RemoteProxyFS) getCache(path string) *fsPathCache {
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

func (fs *RemoteProxyFS) Open(ctx context.Context, path string) (ctxvfs.ReadSeekCloser, error) {
	fs.mu.Lock()
	c := fs.getCache(path)
	if c.readFile != nil {
		readFile := c.readFile
		fs.mu.Unlock()
		return nopCloser{bytes.NewReader(readFile)}, nil
	} else if c.readFileErr != nil {
		readFileErr := c.readFileErr
		fs.mu.Unlock()
		return nil, readFileErr
	}
	fs.mu.Unlock()

	var contents map[string][]byte
	if err := fs.call(ctx, "fs/readDirFiles", path, &contents); err != nil {
		fs.mu.Lock()
		c.readFileErr = err
		fs.mu.Unlock()
		return nil, err
	}
	fs.mu.Lock()
	// Precache all files in the requested path's directory to improve
	// performance across subsequent requests.
	for p, f := range contents {
		c := fs.getCache(p)
		if c.readFile == nil {
			c.readFile = f
		}
	}
	fs.mu.Unlock()
	return nopCloser{bytes.NewReader(contents[path])}, nil
}

func (fs *RemoteProxyFS) Stat(ctx context.Context, path string) (fi os.FileInfo, err error) {
	fs.mu.Lock()
	c := fs.getCache(path)
	if c.stat != nil || c.statErr != nil {
		stat := c.stat
		statErr := c.statErr
		fs.mu.Unlock()
		return stat, statErr
	}
	fs.mu.Unlock()
	defer func() {
		fs.mu.Lock()
		c.stat = fi
		c.statErr = err
		fs.mu.Unlock()
	}()

	fi = &lspext.FileInfo{}
	return fi, fs.call(ctx, "fs/stat", path, &fi)
}

func (fs *RemoteProxyFS) Lstat(ctx context.Context, path string) (fi os.FileInfo, err error) {
	fs.mu.Lock()
	c := fs.getCache(path)
	if c.lstat != nil || c.lstatErr != nil {
		lstat := c.lstat
		lstatErr := c.lstatErr
		fs.mu.Unlock()
		return lstat, lstatErr
	}
	fs.mu.Unlock()
	defer func() {
		fs.mu.Lock()
		c.lstat = fi
		c.lstatErr = err
		fs.mu.Unlock()
	}()

	fi = &lspext.FileInfo{}
	return fi, fs.call(ctx, "fs/lstat", path, &fi)
}

func (fs *RemoteProxyFS) ReadDir(ctx context.Context, path string) (fis []os.FileInfo, err error) {
	fs.mu.Lock()
	c := fs.getCache(path)
	if c.readDir != nil || c.readDirErr != nil {
		readDir := c.readDir
		readDirErr := c.readDirErr
		fs.mu.Unlock()
		return readDir, readDirErr
	}
	fs.mu.Unlock()
	defer func() {
		fs.mu.Lock()
		c.readDir = fis
		c.readDirErr = err
		fs.mu.Unlock()
	}()

	var fis2 []lspext.FileInfo
	if err := fs.call(ctx, "fs/readDir", path, &fis2); err != nil {
		return nil, err
	}
	fis = make([]os.FileInfo, len(fis2))
	for i, fi := range fis2 {
		fis[i] = fi
	}
	return fis, nil
}

func (fs *RemoteProxyFS) String() string {
	return "RemoteProxyFS"
}

type nopCloser struct {
	io.ReadSeeker
}

func (nc nopCloser) Close() error { return nil }
