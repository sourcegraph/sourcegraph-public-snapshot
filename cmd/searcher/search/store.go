package search

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/diskcache"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/lazyzip"
)

// Store manages the fetching and storing of git archives. Its main purpose is
// keeping a local disk cache of the fetched archives to help speed up future
// requests for the same archive. As a performance optimization, it is also
// responsible for filtering out files we receive from `git archive` that we
// do not want to search.
//
// We use an LRU to do cache eviction:
// * When to evict is based on the total size of *.zip on disk.
// * What to evict uses the LRU algorithm.
// * We touch files when opening them, so can do LRU based on file
//   modification times.
//
// Note: The store fetches tarballs but stores zips. We want to be able to
// filter which files we cache, so we need a format that supports streaming
// (tar). We want to be able to support random concurrent access for reading,
// so we store as a zip.
type Store struct {
	// FetchTar returns an io.ReadCloser to a tar archive. If the error
	// implements "BadRequest() bool", it will be used to determine if the
	// error is a bad request (eg invalid repo).
	FetchTar func(ctx context.Context, repo, commit string) (io.ReadCloser, error)

	// Path is the directory to store the cache
	Path string

	// MaxCacheSizeBytes is the maximum size of the cache in bytes. Note:
	// We can temporarily be larger than MaxCacheSizeBytes. When we go
	// over MaxCacheSizeBytes we trigger delete files until we get below
	// MaxCacheSizeBytes.
	MaxCacheSizeBytes int64

	// once protects Start
	once sync.Once

	// cache is the disk backed cache.
	cache *diskcache.Store
}

// Start initializes state and starts background goroutines. It can be called
// more than once. It is optional to call, but starting it earlier avoids a
// search request paying the cost of initializing.
func (s *Store) Start() {
	s.once.Do(func() {
		s.cache = &diskcache.Store{
			Dir:               s.Path,
			Component:         "store",
			BackgroundTimeout: 2 * time.Minute,
		}
		go s.watchAndEvict()
	})
}

// archiveReadCloser is like zip.ReadCloser. We need it since we can't use
// zip.OpenReader.
type archiveReadCloser struct {
	r *lazyzip.Reader
	f *os.File
}

func (ar *archiveReadCloser) Reader() *lazyzip.Reader {
	return ar.r
}

// Close closes the file for the archive.
func (ar *archiveReadCloser) Close() error {
	return ar.f.Close()
}

// openReader will open a zip reader to the archive. It will first consult the
// local cache, otherwise will fetch from the network.
func (s *Store) openReader(ctx context.Context, repo, commit string) (ar *archiveReadCloser, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "OpenReader")
	ext.Component.Set(span, "store")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()

	// Ensure we have initialized
	s.Start()

	// We already validate commit is absolute in ServeHTTP, but since we
	// rely on it for caching we check again.
	if len(commit) != 40 {
		return nil, errors.Errorf("commit must be resolved (repo=%q, commit=%q)", repo, commit)
	}

	// key is a sha256 hash since we want to use it for the disk name
	h := sha256.Sum256([]byte(repo + " " + commit))
	key := hex.EncodeToString(h[:])
	span.LogKV("key", key)

	// Our fetch can take a long time, and the frontend aggressively cancels
	// requests. So we open in the background to give it extra time.
	resC := make(chan struct {
		File *diskcache.File
		Err  error
	})
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		f, err := s.cache.Open(ctx, key, func(ctx context.Context) (io.ReadCloser, error) {
			return s.fetch(ctx, repo, commit)
		})
		resC <- struct {
			File *diskcache.File
			Err  error
		}{f, err}
	}()

	select {
	case <-ctx.Done():
		// We are fetching in the background, so must remember to close the
		// file even though we won't read it.
		go func() {
			res := <-resC
			if res.File != nil {
				res.File.Close()
			}
		}()
		return nil, ctx.Err()

	case res := <-resC:
		if res.Err != nil {
			return nil, res.Err
		}
		fi, err := res.File.Stat()
		if err != nil {
			res.File.Close()
			return nil, err
		}
		zr, err := lazyzip.NewReader(res.File, fi.Size())
		if err != nil {
			res.File.Close()
			return nil, err
		}
		return &archiveReadCloser{r: zr, f: res.File.File}, nil
	}
}

// fetch fetches an archive from the network and stores it on disk. It does
// not populate the in-memory cache. You should probably be calling
// openReader.
func (s *Store) fetch(ctx context.Context, repo, commit string) (rc io.ReadCloser, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Fetch")
	ext.Component.Set(span, "store")
	span.SetTag("repo", repo)
	span.SetTag("commit", commit)
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()

	r, err := s.FetchTar(ctx, repo, commit)
	if err != nil {
		return nil, err
	}

	pr, pw := io.Pipe()

	// Write tr to zw. Return the first error encountered, but clean up if
	// we encounter an error.
	go func() {
		defer r.Close()
		tr := tar.NewReader(r)
		zw := zip.NewWriter(pw)
		err := copySearchable(tr, zw)
		if err1 := zw.Close(); err == nil {
			err = err1
		}
		pw.CloseWithError(err)
	}()

	return pr, nil
}

// copySearchable copies searchable files from tr to zw. A searchable file is
// any file that is a candidate for being searched (under size limit and
// non-binary).
func copySearchable(tr *tar.Reader, zw *zip.Writer) error {
	// 32*1024 is the same size used by io.Copy
	buf := make([]byte, 32*1024)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		// We only care about files
		if hdr.Typeflag != tar.TypeReg && hdr.Typeflag != tar.TypeRegA {
			continue
		}
		// We do not search large files
		if hdr.Size > maxFileSize {
			continue
		}
		// Heuristic: Assume file is binary if first 256 bytes contain a
		// 0x00. Best effort, so ignore err
		n, err := tr.Read(buf)
		if n > 0 && bytes.IndexByte(buf[:n], 0x00) >= 0 {
			continue
		}
		if err == io.EOF {
			// tar.Reader.Read guarantees n == 0 if err ==
			// io.EOF. So we do not have to write anything to zr
			// for an empty file.
			continue
		}
		if err != nil {
			return err
		}

		// We are happy with the file, so we can write it to zw.
		w, err := zw.CreateHeader(&zip.FileHeader{
			Name:   hdr.Name,
			Method: zip.Store,
		})
		if err != nil {
			return err
		}

		// First write the data already read into buf
		nw, err := w.Write(buf[:n])
		if err != nil {
			return err
		}
		if nw != n {
			return io.ErrShortWrite
		}

		_, err = io.CopyBuffer(w, tr, buf)
		if err != nil {
			return err
		}
	}
}

// watchAndEvict is a loop which periodically checks the size of the cache and
// evicts/deletes items if the store gets too large.
func (s *Store) watchAndEvict() {
	if s.MaxCacheSizeBytes == 0 {
		return
	}

	for {
		time.Sleep(10 * time.Second)
		stats, err := s.cache.Evict(s.MaxCacheSizeBytes)
		if err != nil {
			log.Printf("failed to Evict: %s", err)
			continue
		}
		cacheSizeBytes.Set(float64(stats.CacheSize))
		evictions.Add(float64(stats.Evicted))
	}
}

var (
	cacheSizeBytes = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "searcher",
		Subsystem: "store",
		Name:      "cache_size_bytes",
		Help:      "The total size of items in the on disk cache.",
	})
	evictions = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "searcher",
		Subsystem: "store",
		Name:      "evictions",
		Help:      "The total number of items evicted from the cache.",
	})
)

func init() {
	prometheus.MustRegister(cacheSizeBytes)
	prometheus.MustRegister(evictions)
}
