package vfsutil

import (
	"context"
	"io"
	"net/http"

	"golang.org/x/net/context/ctxhttp"

	"github.com/pkg/errors"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// NewZipVFS downloads a zip archive from a URL (or fetches from the local cache
// on disk) and returns a new VFS backed by that zip archive.
func NewZipVFS(url, cacheKey, rootDirInZip string, onFetchStart, onFetchFailed func()) (*ArchiveFS, error) {
	fetch := func(ctx context.Context) (ar *archiveReader, err error) {
		span, ctx := opentracing.StartSpanFromContext(ctx, "zip Fetch")
		ext.Component.Set(span, "zipvfs")
		span.SetTag("url", url)
		defer func() {
			if err != nil {
				ext.Error.Set(span, true)
				span.SetTag("err", err)
			}
			span.Finish()
		}()

		ff, err := cachedFetch(ctx, "zipvfs", cacheKey, func(ctx context.Context) (io.ReadCloser, error) {
			onFetchStart()
			request, err := http.NewRequest("GET", url, nil)
			request.Header.Add("Accept", "application/zip")
			resp, err := ctxhttp.Do(ctx, nil, request)
			if err != nil {
				return nil, errors.Errorf("failed to fetch zip archive from %s: %s", url, err)
			}
			if resp.StatusCode != http.StatusOK {
				resp.Body.Close()
				return nil, errors.Errorf("zip URL %s returned HTTP %d", url, resp.StatusCode)
			}
			return resp.Body, nil
		})
		if err != nil {
			onFetchFailed()
			return nil, errors.Errorf("failed to fetch/write/open zip archive from %s: %s", url, err)
		}
		f := ff.File

		zr, err := zipNewFileReader(f)
		if err != nil {
			f.Close()
			return nil, errors.Errorf("failed to read zip archive associated with local disk cache key %s: %s", cacheKey, err)
		}

		if len(zr.File) == 0 {
			f.Close()
			return nil, errors.Errorf("zip archive %s is empty", cacheKey)
		}

		return &archiveReader{
			Reader: zr,
			Closer: f,
			Prefix: rootDirInZip,
		}, nil
	}

	// TODO(chris) don't eagerly fetch here. Instead, make a composite ArchiveFS
	fs := &ArchiveFS{fetch: fetch}
	err := fs.fetchOrWait(context.Background())
	return fs, err
}
