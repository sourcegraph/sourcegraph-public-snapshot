package golang

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/ctxvfs"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/jsonrpc2"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/lspx"
)

// remoteProxyFS is an implementation of ctxvfs.FileSystem that
// communicates with the LSP proxy server over JSON-RPC to access the
// (virtual) file system.
type remoteProxyFS struct {
	conn *jsonrpc2.Conn
}

// call sends a request to the LSP proxy with tracing information.
func (fs *remoteProxyFS) call(ctx context.Context, method, path string, result interface{}) (err error) {
	op := "LSP server: remote VFS call"
	tags := opentracing.Tags{"method": method, "path": path}
	parentSpan := opentracing.SpanFromContext(ctx)
	span := parentSpan.Tracer().StartSpan(op, tags, opentracing.ChildOf(parentSpan.Context()))
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.LogEvent(fmt.Sprintf("error: %v", err))
		}
		span.Finish()
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

func (fs *remoteProxyFS) Open(ctx context.Context, path string) (ctxvfs.ReadSeekCloser, error) {
	var contents []byte
	if err := fs.call(ctx, "fs/readFile", path, &contents); err != nil {
		return nil, err
	}
	return nopCloser{bytes.NewReader(contents)}, nil
}

func (fs *remoteProxyFS) Stat(ctx context.Context, path string) (os.FileInfo, error) {
	var fi lspx.FileInfo
	return &fi, fs.call(ctx, "fs/stat", path, &fi)
}

func (fs *remoteProxyFS) Lstat(ctx context.Context, path string) (os.FileInfo, error) {
	var fi lspx.FileInfo
	return &fi, fs.call(ctx, "fs/lstat", path, &fi)
}

func (fs *remoteProxyFS) ReadDir(ctx context.Context, path string) ([]os.FileInfo, error) {
	var fis []lspx.FileInfo
	if err := fs.call(ctx, "fs/readDir", path, &fis); err != nil {
		return nil, err
	}
	fis2 := make([]os.FileInfo, len(fis))
	for i, fi := range fis {
		fis2[i] = fi
	}
	return fis2, nil
}

func (fs *remoteProxyFS) String() string {
	return "remoteProxyFS"
}

type nopCloser struct {
	io.ReadSeeker
}

func (nc nopCloser) Close() error { return nil }
