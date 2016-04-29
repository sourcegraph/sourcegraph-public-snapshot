package graphstoreutil

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"time"

	"github.com/golang/groupcache/lru"
	"github.com/prometheus/client_golang/prometheus"

	"golang.org/x/tools/godoc/vfs"
	"sourcegraph.com/sourcegraph/rwvfs"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/cache"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/cachedvfs"
)

var (
	// godoc has around 100k repos. __versions is immutable, so only cache
	// to increase speed of repeated requests
	versionCache = cache.TTL(cache.Sync(lru.New(100000)), time.Minute)
	versionRe    = regexp.MustCompile(`^.*/__versions$`)

	// idx files have the potential of being large, so a slightly
	// conservative cache size. Also they are _mostly_ immutable, but we
	// add a large TTL just in case.
	idxCache = cache.TTL(cache.Sync(lru.New(10000)), time.Hour)
	idxRe    = regexp.MustCompile(`^.*\.idx$`)
)

// withVFSCache returns a graphstore which caches commonly accessed
// objects. It is targeted at the cloudstoragevfs.
func withVFSCache(rwfs rwvfs.FileSystem) rwvfs.FileSystem {
	cfs := cachedvfs.New(rwfs, versionCache, versionRe)
	cfs = cachedvfs.New(cfs, idxCache, idxRe)
	return &mergedFS{fs: cfs, rwfs: rwfs}
}

// mergedFS will use fs for the read operations, otherwise will use rwfs for
// the additional operations.
//
// This wrapper only satisfies the rwvfs.FileSystem interface, but quite often
// implementors of rwvfs.FileSystem satisfy a few extra interfaces (like
// MkdirAllOverrider). We just target the interfaces that cloudStorageVFS
// targets since that is what we use in production.
type mergedFS struct {
	fs   vfs.FileSystem
	rwfs rwvfs.FileSystem
}

func (f *mergedFS) Open(name string) (vfs.ReadSeekCloser, error) { return f.fs.Open(name) }
func (f *mergedFS) Lstat(path string) (os.FileInfo, error)       { return f.fs.Lstat(path) }
func (f *mergedFS) Stat(path string) (os.FileInfo, error)        { return f.fs.Stat(path) }
func (f *mergedFS) ReadDir(path string) ([]os.FileInfo, error)   { return f.fs.ReadDir(path) }
func (f *mergedFS) String() string {
	return fmt.Sprintf("mergedFS(%s, %s)", f.fs.String(), f.rwfs.String())
}

func (f *mergedFS) Create(path string) (io.WriteCloser, error) { return f.rwfs.Create(path) }
func (f *mergedFS) Mkdir(name string) error                    { return f.rwfs.Mkdir(name) }
func (f *mergedFS) Remove(name string) error                   { return f.rwfs.Remove(name) }

// Register some cache hit metrics
func init() {
	c := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "graphstore",
		Name:      "cache_hit",
		Help:      "Counts cache hits and misses for graphstore cache.",
	}, []string{"cache", "type"})
	prometheus.MustRegister(c)
	versionCache = cache.Hook(
		versionCache,
		c.WithLabelValues("version", "hit").Inc,
		c.WithLabelValues("version", "miss").Inc,
	)
	idxCache = cache.Hook(
		idxCache,
		c.WithLabelValues("idx", "hit").Inc,
		c.WithLabelValues("idx", "miss").Inc,
	)
}
