package search

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/golang/groupcache/lru"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
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
// * To evict we need to wait for ongoing searches on the file to finish. We can
//   then close the zip and file.
// * Since we are based on the size of files on disk, we need to add files on disk
//   which are not yet tracked in-memory. These files can exist due to restarts in
//   the process/pod.
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

	// archives stores all archives in the cache.
	archivesMu sync.Mutex
	archives   *lru.Cache

	// once protects Start
	once sync.Once
}

// Start initializes state and starts background goroutines. It can be called
// more than once. It is optional to call, but starting it earlier avoids a
// search request paying the cost of initializing.
func (s *Store) Start() {
	s.once.Do(func() {
		s.archives = lru.New(0)
		s.archives.OnEvicted = func(key lru.Key, value interface{}) {
			// When we Remove or RemoveOldest we hold the lock, so
			// safe to read s.archive.
			cacheSizeLength.Set(float64(s.archives.Len()))
			evictions.Inc()
			go value.(*archive).onEvicted()
		}
		err := s.openOnDisk()
		if err != nil {
			log.Println("failed to open pre-existing cached items: ", err)
		}
		go s.watchAndEvict()
	})
}

// archive stores the result of fetching an archive.
type archive struct {
	reader *zip.ReadCloser
	err    error
	done   chan struct{} // closed to signal reader and err are set

	path string         // stored so we can os.Remove when evicted
	wg   sync.WaitGroup // tracks open readers (ongoing searches)
}

// open returns a new reader to the underlying zip file. It must be closed so
// that we can cleanly evict the archive.
func (a *archive) open(ctx context.Context) (*archiveReadCloser, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-a.done:
		break
	}
	if a.err != nil {
		return nil, a.err
	}
	a.wg.Add(1)
	openReaders.Inc()
	return &archiveReadCloser{
		archive: a,
	}, nil
}

func (a *archive) onEvicted() {
	<-a.done
	// Best effort removal. We can remove the file and open readers will
	// still work.
	_ = os.Remove(a.path)
	if a.err != nil {
		return
	}
	a.wg.Wait()
	a.reader.Close()
}

// archiveReadCloser exposes an interface like zip.ReadCloser. However, we use
// our own close logic.
type archiveReadCloser struct {
	archive *archive
	closed  uint32
}

func (a *archiveReadCloser) Reader() *zip.Reader {
	return &a.archive.reader.Reader
}

func (a *archiveReadCloser) Close() error {
	if !atomic.CompareAndSwapUint32(&a.closed, 0, 1) {
		return errors.New("already closed")
	}
	a.archive.wg.Done()
	openReaders.Dec()
	return nil
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
		if ar != nil {
			span.LogKV("numfiles", len(ar.Reader().File))
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
	path := filepath.Join(s.Path, key+".zip")
	span.LogKV("key", key)

	s.archivesMu.Lock()
	value, hasArchive := s.archives.Get(key)
	if !hasArchive {
		value = &archive{
			path: path,
			done: make(chan struct{}),
		}
		s.archives.Add(key, value)
		cacheSizeLength.Set(float64(s.archives.Len()))
		added.Inc()
	}
	s.archivesMu.Unlock()
	arch := value.(*archive)

	if !hasArchive {
		// We need to fetch the archive and populate s.archives. We do
		// it in the background since other readers may be affected.
		go func() {
			arch.reader, arch.err = s.fetch(repo, commit, path)
			close(arch.done)
			// If we failed, remove from archive cache so we try again in
			// the future.
			if arch.err != nil {
				s.archivesMu.Lock()
				s.archives.Remove(key)
				s.archivesMu.Unlock()
			}
		}()
	}

	return arch.open(ctx)
}

// fetch fetches an archive from the network and stores it on disk. It does
// not populate the in-memory cache. You should probably be calling
// openReader.
func (s *Store) fetch(repo, commit, path string) (zr *zip.ReadCloser, err error) {
	// Background context since can be assocaited with more than 1
	// concurrent request.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	span, ctx := opentracing.StartSpanFromContext(ctx, "Fetch")
	ext.Component.Set(span, "store")
	span.SetTag("repo", repo)
	span.SetTag("commit", commit)
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		if zr != nil {
			span.LogKV("numfiles", len(zr.File))
		}
		span.Finish()
	}()

	r, err := s.FetchTar(ctx, repo, commit)
	if err != nil {
		return nil, err
	}
	err = populateDiskCache(path, r)
	r.Close()
	if err != nil {
		return nil, err
	}
	zr, err = zip.OpenReader(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open archive after fetching")
	}
	return zr, nil
}

// populateDiskCache puts the item into the disk cache at path atomically.
func populateDiskCache(path string, r io.Reader) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return errors.Wrap(err, "could not create cache dir")
	}

	// We write to a temporary path to prevent openReader finding a
	// partialy written zip.
	tmpPath := path + ".part"
	f, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return errors.Wrap(err, "failed to create temporary cache item")
	}
	defer os.Remove(tmpPath)

	// Write tr to zw. Return the first error encountered, but clean up if
	// we encounter an error.
	tr := tar.NewReader(r)
	zw := zip.NewWriter(f)
	err = copySearchable(tr, zw)
	if err1 := zw.Close(); err == nil {
		err = err1
	}
	if err1 := f.Close(); err == nil {
		err = err1
	}
	if err != nil {
		return err
	}

	return os.Rename(tmpPath, path)
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

	cacheSize := func() int64 {
		list, err := ioutil.ReadDir(s.Path)
		if err != nil {
			log.Printf("failed to ReadDir(%s): %s", s.Path, err)
			return 0
		}

		var size int64
		for _, fi := range list {
			if !strings.HasSuffix(fi.Name(), ".zip") {
				continue
			}
			size += fi.Size()
		}
		cacheSizeBytes.Set(float64(size))
		return size
	}

	for {
		if cacheSize() <= s.MaxCacheSizeBytes {
			time.Sleep(10 * time.Second)
			continue
		}

		s.archivesMu.Lock()
		s.archives.RemoveOldest()
		s.archivesMu.Unlock()

		// Give some time to settle
		time.Sleep(time.Second)
	}
}

// openOnDisk opens all files on disk but not yet in s.archives.
func (s *Store) openOnDisk() error {
	// We are initializing s.archives
	s.archivesMu.Lock()
	defer s.archivesMu.Unlock()

	if err := os.MkdirAll(s.Path, 0700); err != nil {
		return err
	}

	onDisk, err := ioutil.ReadDir(s.Path)
	if err != nil {
		return err
	}

	for _, fi := range onDisk {
		path := filepath.Join(s.Path, fi.Name())
		if !strings.HasSuffix(fi.Name(), ".zip") {
			// We should only have cache items in path. Most
			// common non-cache item is .zip.part files
			err = os.Remove(path)
			if err != nil {
				log.Printf("failed to remove %s: %s", path, err)
			}
		}

		zr, err := zip.OpenReader(path)
		if err != nil {
			// rather not have it in the cache if we can't
			// open it.
			if !os.IsNotExist(err) {
				err = os.Remove(path)
				if err != nil {
					log.Printf("failed to remove %s: %s", path, err)
				}
			}
			continue
		}

		key := strings.TrimSuffix(fi.Name(), ".zip")
		arch := &archive{
			path:   path,
			reader: zr,
			done:   make(chan struct{}),
		}
		close(arch.done)
		s.archives.Add(key, arch)
		cacheSizeLength.Set(float64(s.archives.Len()))
		added.Inc()
	}

	return nil
}

var (
	cacheSizeLength = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "searcher",
		Subsystem: "store",
		Name:      "cache_size_length",
		Help:      "The number of items in the cache.",
	})
	cacheSizeBytes = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "searcher",
		Subsystem: "store",
		Name:      "cache_size_bytes",
		Help:      "The total size of items in the on disk cache.",
	})
	openReaders = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "searcher",
		Subsystem: "store",
		Name:      "open_readers",
		Help:      "The total number of open readers.",
	})
	added = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "searcher",
		Subsystem: "store",
		Name:      "added",
		Help:      "The total number of items added to the cache.",
	})
	evictions = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "searcher",
		Subsystem: "store",
		Name:      "evictions",
		Help:      "The total number of items evicted from the cache.",
	})
)

func init() {
	prometheus.MustRegister(cacheSizeLength)
	prometheus.MustRegister(cacheSizeBytes)
	prometheus.MustRegister(openReaders)
	prometheus.MustRegister(added)
	prometheus.MustRegister(evictions)
}
