package diskcache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
)

// Store is an on disk cache, with items cached via calls to Open.
type Store struct {
	// Dir is the directory to cache items.
	Dir string

	// Component when set is reported to OpenTracing as the component.
	Component string

	// BackgroundTimeout when non-zero will do fetches in the background with
	// a timeout. This means the context passed to fetch will be
	// context.WithTimeout(context.Background(), BackgroundTimeout). When not
	// set fetches are done with the passed in context.
	BackgroundTimeout time.Duration

	// BeforeEvict, when non-nil, is a function to call before evicting a file.
	// It is passed the path to the file to be evicted.
	BeforeEvict func(string)
}

// File is an os.File, but includes the Path
type File struct {
	*os.File

	// The Path on disk for File
	Path string
}

// Fetcher returns a ReadCloser. It is used by Open if the key is not in the
// cache.
type Fetcher func(context.Context) (io.ReadCloser, error)

// FetcherWithPath writes a cache entry to the given file. It is used by Open if the key
// is not in the cache.
type FetcherWithPath func(context.Context, string) error

// Open will open a file from the local cache with key. If missing, fetcher
// will fill the cache first. Open also performs single-flighting for fetcher.
func (s *Store) Open(ctx context.Context, key string, fetcher Fetcher) (file *File, err error) {
	return s.OpenWithPath(ctx, key, func(ctx context.Context, path string) error {
		readCloser, err := fetcher(ctx)
		if err != nil {
			return err
		}
		file, err := os.OpenFile(path, os.O_WRONLY, 0600)
		if err != nil {
			readCloser.Close()
			return errors.Wrap(err, "failed to open temporary archive cache item")
		}
		err = copyAndClose(file, readCloser)
		if err != nil {
			return errors.Wrap(err, "failed to copy and close missing archive cache item")
		}
		return nil
	})
}

// OpenWithPath will open a file from the local cache with key. If missing, fetcher
// will fill the cache first. Open also performs single-flighting for fetcher.
func (s *Store) OpenWithPath(ctx context.Context, key string, fetcher FetcherWithPath) (file *File, err error) {
	span, ctx := ot.StartSpanFromContext(ctx, "Cached Fetch")
	if s.Component != "" {
		ext.Component.Set(span, s.Component)
	}
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		if file != nil {
			// Update modified time. Modified time is used to decide which
			// files to evict from the cache.
			touch(file.Path)
		}
		span.Finish()
	}()

	if s.Dir == "" {
		return nil, errors.New("diskcache.Store.Dir must be set")
	}

	path := s.path(key)
	span.LogKV("key", key, "path", path)

	// First do a fast-path, assume already on disk
	f, err := os.Open(path)
	if err == nil {
		span.SetTag("source", "fast")
		return &File{File: f, Path: path}, nil
	}

	// We (probably) have to fetch
	span.SetTag("source", "fetch")

	// Do the fetch in another goroutine so we can respect ctx cancellation.
	type result struct {
		f   *File
		err error
	}
	ch := make(chan result, 1)
	go func(ctx context.Context) {
		if s.BackgroundTimeout != 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(context.Background(), s.BackgroundTimeout)
			defer cancel()
		}
		f, err := doFetch(ctx, path, fetcher)
		ch <- result{f, err}
	}(ctx)

	select {
	case <-ctx.Done():
		// *os.File sets a finalizer to close the file when no longer used, so
		// we don't need to worry about closing the file in the case of context
		// cancellation.
		return nil, ctx.Err()
	case r := <-ch:
		return r.f, r.err
	}
}

// path returns the path for key.
func (s *Store) path(key string) string {
	// path uses a sha256 hash of the key since we want to use it for the
	// disk name.
	h := sha256.Sum256([]byte(key))
	return filepath.Join(s.Dir, hex.EncodeToString(h[:])) + ".zip"
}

func doFetch(ctx context.Context, path string, fetcher FetcherWithPath) (file *File, err error) {
	// We have to grab the lock for this key, so we can fetch or wait for
	// someone else to finish fetching.
	urlMu := urlMu(path)
	urlMu.Lock()
	defer urlMu.Unlock()

	// Since we acquired the lock we may have timed out.
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Since we acquired urlMu, someone else may have put the archive onto
	// the disk.
	f, err := os.Open(path)
	if err == nil {
		return &File{File: f, Path: path}, nil
	}
	// Just in case we failed due to something bad on the FS, remove
	_ = os.Remove(path)

	// Fetch since we still can't open up the file
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, errors.Wrap(err, "could not create archive cache dir")
	}

	// We write to a temporary path to prevent another Open finding a
	// partially written file. We ensure the file is writeable and truncate
	// it.
	tmpPath := path + ".part"
	f, err = os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create temporary archive cache item")
	}
	f.Close()
	defer os.Remove(tmpPath)

	// We are now ready to actually fetch the file.
	err = fetcher(ctx, tmpPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch missing archive cache item")
	}

	// Sync the contents to disk. If we crash we don't want to leave behind
	// invalid zip files due to unwritten OS buffers.
	if err := fsync(tmpPath); err != nil {
		return nil, errors.Wrap(err, "failed to sync cache item to disk")
	}

	// Put the partially written file in the correct place and open
	err = os.Rename(tmpPath, path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to put cache item in place")
	}

	// Sync the directory. We need to ensure the rename is recorded to disk.
	if err := fsync(filepath.Dir(path)); err != nil {
		return nil, errors.Wrap(err, "failed to sync cache directory to disk")
	}

	f, err = os.Open(path)
	if err != nil {
		return nil, err
	}
	return &File{File: f, Path: path}, nil
}

// EvictStats is information gathered during Evict.
type EvictStats struct {
	// CacheSize is the size of the cache before evicting.
	CacheSize int64

	// Evicted is the number of items evicted.
	Evicted int
}

// Evict will remove files from Store.Dir until it is smaller than
// maxCacheSizeBytes. It evicts files with the oldest modification time first.
func (s *Store) Evict(maxCacheSizeBytes int64) (stats EvictStats, err error) {
	isZip := func(fi os.FileInfo) bool {
		return strings.HasSuffix(fi.Name(), ".zip")
	}

	list, err := ioutil.ReadDir(s.Dir)
	if err != nil {
		if os.IsNotExist(err) {
			return EvictStats{
				CacheSize: 0,
				Evicted:   0,
			}, nil
		}
		return stats, errors.Wrapf(err, "failed to ReadDir %s", s.Dir)
	}

	// Sum up the total size of all zips
	var size int64
	for _, fi := range list {
		if isZip(fi) {
			size += fi.Size()
		}
	}
	stats.CacheSize = size

	// Nothing to evict
	if size <= maxCacheSizeBytes {
		return stats, nil
	}

	// Keep removing files until we are under the cache size. Remove the
	// oldest first.
	sort.Slice(list, func(i, j int) bool {
		return list[i].ModTime().Before(list[j].ModTime())
	})
	for _, fi := range list {
		if size <= maxCacheSizeBytes {
			break
		}
		if !isZip(fi) {
			continue
		}
		path := filepath.Join(s.Dir, fi.Name())
		if s.BeforeEvict != nil {
			s.BeforeEvict(path)
		}
		err = os.Remove(path)
		if err != nil {
			log.Printf("failed to remove %s: %s", path, err)
			continue
		}
		stats.Evicted++
		size -= fi.Size()
	}

	return stats, nil
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

// touch updates the modified time to time.Now(). It is best-effort, and will
// log if it fails.
func touch(path string) {
	t := time.Now()
	if err := os.Chtimes(path, t, t); err != nil {
		log.Printf("failed to touch %s: %s", path, err)
	}
}

func fsync(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	err = f.Sync()
	if err1 := f.Close(); err == nil {
		err = err1
	}
	return err
}
