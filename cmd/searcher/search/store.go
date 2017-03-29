package search

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"sync"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
)

// Store manages the fetching and storing of git archives. Its main purpose is
// keeping a local disk cache of the fetched archives to help speed up future
// requests for the same archive. As a performance optimization, it is also
// responsible for filtering out files we receive from `git archive` that we
// do not want to search.
//
// Note: The store fetches tarballs but stores zips. We want to be able to
// filter which files we cache, so we need a format that supports streaming
// (tar). We want to be able to support random concurrent access for reading,
// so we store as a zip.
//
// TODO(keegan) remove items from cache.
type Store struct {
	// FetchTar returns an io.ReadCloser to a tar archive. If the error
	// implements "BadRequest() bool", it will be used to determine if the
	// error is a bad request (eg invalid repo).
	FetchTar func(ctx context.Context, repo, commit string) (io.ReadCloser, error)

	// Path is the directory to store the cache
	Path string

	// archives stores all archives in the cache.
	archivesMu sync.Mutex
	archives   map[string]*archive
}

// archive stores the result of fetching an archive.
type archive struct {
	reader *zip.Reader
	err    error
	done   chan struct{} // closed to signal reader and err are set
}

// openReader will open a zip reader to the archive. It will first consult the
// local cache, otherwise will fetch from the network.
func (s *Store) openReader(ctx context.Context, repo, commit string) (zr *zip.Reader, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "OpenReader")
	ext.Component.Set(span, "store")
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
		return nil, errors.Errorf("commit must be resolved (repo=%q, commit=%q)", repo, commit)
	}

	// key is a sha256 hash since we want to use it for the disk name
	h := sha256.Sum256([]byte(repo + " " + commit))
	key := hex.EncodeToString(h[:])
	span.LogKV("key", key)

	s.archivesMu.Lock()
	if s.archives == nil {
		s.archives = make(map[string]*archive)
	}
	arch, ok := s.archives[key]
	if !ok {
		arch = &archive{done: make(chan struct{})}
		s.archives[key] = arch
	}
	s.archivesMu.Unlock()

	// We already have an open reader
	if ok {
		span.SetTag("source", "open")
		<-arch.done
		return arch.reader, arch.err
	}

	// We need to fetch the archive and populate s.archives for future
	// readers.
	defer func() {
		arch.reader, arch.err = zr, err
		close(arch.done)
		// If we failed, remove from archive cache so we try again in
		// the future.
		if err != nil {
			s.archivesMu.Lock()
			delete(s.archives, key)
			s.archivesMu.Unlock()
		}
	}()

	// It may be on disk from a previously running searcher. If we fail to
	// open, we will just fetch from network.
	//
	// Note: We never close the opened file, so we do not store the
	// closer. That is because currently we keep the file open for the
	// lifetime of the process.
	path := filepath.Join(s.Path, key+".zip")
	zrc, err := zip.OpenReader(path)
	if err == nil {
		span.SetTag("source", "disk")
		return &zrc.Reader, nil
	}

	span.SetTag("source", "fetch")
	r, err := s.FetchTar(ctx, repo, commit)
	if err != nil {
		return nil, err
	}
	err = populateDiskCache(path, r)
	r.Close()

	zrc, err = zip.OpenReader(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open archive after fetching")
	}
	return &zrc.Reader, nil
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
