package vfsutil

import (
	"archive/zip"
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/pkg/diskcache"
)

// ArchiveCacheDir is the location on disk that archives are cached. It is
// configurable so that in production we can point it into CACHE_DIR.
var ArchiveCacheDir = "/tmp/xlang-archive-cache"

// Evicter implements Evict
type Evicter interface {
	// Evict evicts an item from a cache.
	Evict()
}

type cachedFile struct {
	// File is an open FD to the fetched data
	File *os.File

	// path is the disk path for File
	path string
}

// Evict will remove the file from the cache. It does not close File. It also
// does not protect against other open readers or concurrent fetches.
func (f *cachedFile) Evict() {
	// Best-effort. Ignore error
	_ = os.Remove(f.path)
	cachedFileEvict.Inc()
}

// cachedFetch will open a file from the local cache with key. If missing,
// fetcher will fill the cache first. cachedFetch also performs
// single-flighting.
func cachedFetch(ctx context.Context, component, key string, fetcher func(context.Context) (io.ReadCloser, error)) (ff *cachedFile, err error) {
	s := &diskcache.Store{
		// Dir uses component as a subdir to prevent conflicts between
		// components with the same key.
		Dir:       filepath.Join(ArchiveCacheDir, component),
		Component: component,
	}
	f, err := s.Open(ctx, key, fetcher)
	if err != nil {
		return nil, err
	}
	return &cachedFile{
		File: f.File,
		path: f.Path,
	}, nil
}

func zipNewFileReader(f *os.File) (*zip.Reader, error) {
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	return zip.NewReader(f, fi.Size())
}

var cachedFileEvict = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: "xlang",
	Subsystem: "vfs",
	Name:      "cached_file_evict",
	Help:      "Total number of evictions to cachedFetch archives.",
})

func init() {
	prometheus.MustRegister(cachedFileEvict)
}
