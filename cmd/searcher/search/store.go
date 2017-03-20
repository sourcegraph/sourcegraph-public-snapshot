package search

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// Store manages the fetching and storing of git archives. Its main purpose is
// keeping a local disk cache of the fetched archives, to help speed up future
// requests for the same archive. As a performance optimization, it is also
// responsible for filtering out files we receive from `git archive` that we
// do not want to search.
//
// TODO:
// * Remove items from cache.
// * Experiment with removing file but keep fd open.
// * Experiment with using tar instead of zip. For fetching a tar is likely
// more performant. For searching concurrent random access of zips may be more
// useful.
// * Experiment with filtering at the git archive layer (current archiving is
// very simple).
type Store struct {
	// FetchZip returns a []byte to a zip archive. If the error implements
	// "BadRequest() bool", it will be used to determine if the error is a
	// bad request (eg invalid repo).
	//
	// NOTE: gitcmd.Open.Archive returns the bytes in memory. However, we
	// only need to be able to stream it in. Update to io.ReadCloser once
	// we have a nice way to stream in the archive.
	FetchZip func(ctx context.Context, repo, commit string) ([]byte, error)

	// Path is the directory to store the cache
	Path string

	// fetches tracks in progress FetchZip operations. Its main use is to
	// single-flight fetches.
	fetchesMu sync.Mutex
	fetches   map[string]*fetchResult
}

// fetchResult stores the result of FetchZip
type fetchResult struct {
	// b is the fetched content. Set when done is closed
	b []byte
	// err is the error that occurred while fetching. Set when done is closed
	err error
	// done is closed when the fetch is complete
	done chan struct{}
}

// resolve returns the zip once it has been fetched.
func (r *fetchResult) resolve() (*zip.Reader, func() error, error) {
	<-r.done
	if r.err != nil {
		return nil, nil, r.err
	}
	rAt := bytes.NewReader(r.b)
	zr, err := zip.NewReader(rAt, int64(len(r.b)))
	nopCloser := func() error { return nil }
	return zr, nopCloser, err
}

// openReader will open a zip reader and closer to the archive. It will first
// consult the local cache, otherwise will fetch from the network.
func (s *Store) openReader(ctx context.Context, repo, commit string) (zr *zip.Reader, closer func() error, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Searcher OpenReader")
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

	// We already validate commit is absolute in ServeHTTP, but since we
	// rely on it for caching we check again.
	if len(commit) != 40 {
		return nil, nil, fmt.Errorf("commit must be resolved (repo=%q, commit=%q)", repo, commit)
	}

	// key is a sha256 hash since we want to use it for the disk name
	h := sha256.Sum256([]byte(repo + " " + commit))
	key := hex.EncodeToString(h[:])
	path := filepath.Join(s.Path, key+".zip")
	span.LogKV("key", key)

	// Fast path: check disk without a lock
	r, err := zip.OpenReader(path)
	if err == nil {
		span.SetTag("source", "disk-fast")
		return &r.Reader, r.Close, nil
	}
	if !os.IsNotExist(err) {
		// TODO do we want to do best effort here in case the local
		// disk cache is misbehaving.
		return nil, nil, err
	}

	// Now check while holding the lock. We try the inflight map first,
	// then check disk again. If both those fail, we fetch the archive
	// from gitserver.
	s.fetchesMu.Lock()
	if s.fetches == nil {
		s.fetches = make(map[string]*fetchResult)
	}
	fetch := s.fetches[key]
	if fetch != nil {
		// An in flight fetch is occurring
		s.fetchesMu.Unlock()
		span.SetTag("source", "inflight")
		span.LogEvent("resolve")
		return fetch.resolve()
	}
	// We do not use zip.OpenReader since that does a lot more work (parse
	// zip header, etc) and we are currently holding fetchesMu. A simple
	// os.Stat is also not used, since the following os.OpenReader would
	// race with the cache invalidation.
	fd, err := os.Open(path)
	if err == nil {
		// Since we last checked the disk, another fetch completed
		s.fetchesMu.Unlock()
		span.SetTag("source", "disk-slow")
		return zipReadCloser(fd)
	}
	// Nothing on disk or in-memory, we will now fetch.
	fetch = &fetchResult{done: make(chan struct{})}
	s.fetches[key] = fetch
	s.fetchesMu.Unlock()
	span.SetTag("source", "fetch")

	go func() {
		fetch.b, fetch.err = s.FetchZip(ctx, repo, commit)
		close(fetch.done)
		defer func() {
			s.fetchesMu.Lock()
			delete(s.fetches, key)
			s.fetchesMu.Unlock()
		}()
		if fetch.err == nil {
			populateDiskCache(path, fetch.b)
		}
	}()

	span.LogEvent("resolve")
	return fetch.resolve()
}

// populateDiskCache puts the item into the disk cache at path atomically
func populateDiskCache(path string, b []byte) {
	// TODO remove cache items if we store too many items
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		log.Printf("failed to create dir for disk cache %s: %s", filepath.Dir(path), err)
		return
	}
	tmpPath := path + ".part"
	err := ioutil.WriteFile(tmpPath, b, 0600)
	if err != nil {
		log.Printf("failed to populate %d bytes to disk cache at %s: %s", len(b), tmpPath, err)
		return
	}
	err = os.Rename(tmpPath, path)
	if err != nil {
		log.Printf("failed to move partial cache into place at %s: %s", tmpPath, err)
		err = os.Remove(tmpPath)
		if err != nil {
			log.Printf("failed to cleanup failed partial move at %s: %s", tmpPath, err)
		}
		return
	}
}

// zipReadCloser is based on zip.OpenReader, except takes an already open
// os.File.
func zipReadCloser(f *os.File) (*zip.Reader, func() error, error) {
	fi, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, nil, err
	}
	zr, err := zip.NewReader(f, fi.Size())
	if err != nil {
		f.Close()
		return nil, nil, err
	}
	return zr, f.Close, nil
}
