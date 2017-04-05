package vfsutil

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
)

// ArchiveCacheDir is the location on disk that archives are cached. It is
// configurable so that in production we can point it into CACHE_DIR.
var ArchiveCacheDir = "/tmp/xlang-archive-cache"

// cachedFetch will open a file from the local cache with key. If missing,
// fetcher will fill the cache first. cachedFetch also performs
// single-flighting.
func cachedFetch(ctx context.Context, component, key string, fetcher func(context.Context) (io.ReadCloser, error)) (f *os.File, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Cached Fetch")
	ext.Component.Set(span, component)
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()

	// path uses a sha256 hash of the key since we want to use it for the
	// disk name.
	h := sha256.Sum256([]byte(key))
	path := filepath.Join(ArchiveCacheDir, component, hex.EncodeToString(h[:]))
	span.LogKV("key", key, "path", path)

	// First do a fast-path, assume already on disk
	f, err = os.Open(path)
	if err == nil {
		span.SetTag("source", "fast")
		return f, nil
	}

	// We have to grab the lock for this key, so we can fetch or wait for
	// someone else to finish fetching.
	urlMu := urlMu(component + " " + key)
	urlMu.Lock()
	defer urlMu.Unlock()
	span.LogEvent("urlMu acquired")

	// Since we acquired urlMu, someone else may have put the archive onto
	// the disk.
	f, err = os.Open(path)
	if err == nil {
		span.SetTag("source", "other")
		return f, nil
	}
	// Just in case we failed due to something bad on the FS, remove
	_ = os.Remove(path)

	// Fetch since we still can't open up the file
	span.SetTag("source", "fetch")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, errors.Wrap(err, "could not create archive cache dir")
	}

	// We write to a temporary path to prevent another cachedFetch finding a
	// partialy written file.
	tmpPath := path + ".part"
	f, err = os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create temporary archive cache item")
	}
	defer os.Remove(tmpPath)

	// We are now ready to actually fetch the file. Write it to the
	// partial file and cleanup.
	r, err := fetcher(ctx)
	if err != nil {
		f.Close()
		return nil, errors.Wrap(err, "failed to fetch missing archive cache item")
	}
	err = copyAndClose(f, r)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch missing archive cache item")
	}

	// Put the partially written file in the correct place and open
	err = os.Rename(tmpPath, path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to put cache item in place")
	}
	return os.Open(path)
}

func copyAndClose(dst io.WriteCloser, src io.ReadCloser) error {
	_, err := io.Copy(dst, src)
	if err1 := src.Close(); err == nil {
		err = err1
	}
	if err1 := dst.Close(); err == nil {
		err = err1
	}
	return err
}

func zipNewFileReader(f *os.File) (*zip.Reader, error) {
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	return zip.NewReader(f, fi.Size())
}
