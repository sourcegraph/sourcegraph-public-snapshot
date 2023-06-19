package search

import (
	"archive/zip"
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"os"
	"sort"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// A zipCache is a shared data structure that provides efficient access to a collection of zip files.
// The zero value is usable.
type zipCache struct {
	// Split the cache into many parts, to minimize lock contention.
	// This matters because, for simplicity,
	// we sometimes hold the lock for long-running operations,
	// such as reading a zip file from disk
	// or waiting for all users of a zip file to finish their work.
	// (The latter case should basically never block, since it only
	// occurs when a file is being deleted, and files are deleted
	// when no one has used them for a long time. Nevertheless, take care.)
	shards [64]zipCacheShard
}

type zipCacheShard struct {
	mu sync.Mutex
	m  map[string]*zipFile // path -> zipFile
}

func (c *zipCache) shardFor(path string) *zipCacheShard {
	h := fnv.New32()
	_, _ = io.WriteString(h, path)
	return &c.shards[h.Sum32()%uint32(len(c.shards))]
}

// Get returns a zipFile for the file on disk at path.
// The file MUST be Closed when it is no longer needed.
func (c *zipCache) Get(path string) (*zipFile, error) {
	shard := c.shardFor(path)
	shard.mu.Lock()
	defer shard.mu.Unlock()
	if shard.m == nil {
		shard.m = make(map[string]*zipFile)
	}
	zf, ok := shard.m[path]
	if ok {
		zf.wg.Add(1)
		return zf, nil
	}
	// Cache miss.
	// Reading zip files is fast enough that we can populate the map in-band,
	// which also conveniently provides free single-flighting.
	zf, err := readZipFile(path)
	if err != nil {
		return nil, err
	}
	shard.m[path] = zf
	zf.wg.Add(1)
	return zf, nil
}

func (c *zipCache) delete(path string, trace observation.TraceLogger) {
	shard := c.shardFor(path)
	shard.mu.Lock()
	defer shard.mu.Unlock()
	zf, ok := shard.m[path]
	if !ok {
		// already deleted?!
		return
	}
	// Wait for all clients using this zipFile to complete their work.
	zf.wg.Wait()
	// Mock zipFiles have nil f. Only try to munmap and close f if it is non-nil.
	if zf.f != nil {
		// For now, only log errors here.
		// These calls shouldn't ever fail, and if they do,
		// there's not much to do about it; best to just limp along.
		if err := unmap(zf.Data); err != nil {
			log.Printf("failed to munmap %q: %v", zf.f.Name(), err)
		}
		if err := zf.f.Close(); err != nil {
			log.Printf("failed to close %q: %v", zf.f.Name(), err)
		}
	}
	delete(shard.m, path)
}

// zipFile provides efficient access to a single zip file.
type zipFile struct {
	// Take care with the size of this struct.
	// There are many zipFiles present during typical usage.
	Files  []srcFile
	MaxLen int
	Data   []byte
	f      *os.File
	wg     sync.WaitGroup // ensures underlying file is not munmap'd or closed while in use
}

func readZipFile(path string) (*zipFile, error) {
	// Open zip file at path, prepare to read it.
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	r, err := zip.NewReader(f, fi.Size())
	if err != nil {
		return nil, err
	}

	// Create at populate ZipFile from contents.
	zf := &zipFile{f: f}
	if err := zf.PopulateFiles(r); err != nil {
		return nil, err
	}

	zf.Data, err = mmap(path, f, fi)
	if err != nil {
		return nil, err
	}

	return zf, nil
}

func (f *zipFile) PopulateFiles(r *zip.Reader) error {
	f.Files = make([]srcFile, len(r.File))
	for i, file := range r.File {
		if file.Method != zip.Store {
			return errors.Errorf("file %s stored with compression %v, want %v", file.Name, file.Method, zip.Store)
		}
		off, err := file.DataOffset()
		if err != nil {
			return err
		}
		size := int(file.UncompressedSize64)
		if uint64(size) != file.UncompressedSize64 {
			return errors.Errorf("file %s has size > 2gb: %v", file.Name, size)
		}
		f.Files[i] = srcFile{Name: file.Name, Off: off, Len: int32(size)}
		if size > f.MaxLen {
			f.MaxLen = size
		}
	}

	// We want sequential reads.
	// We wrote this zip file ourselves, in one pass,
	// so r.File should already be ordered by DataOffset.
	// Sort anyway just to make sure.
	sort.Slice(f.Files, func(i, j int) bool { return f.Files[i].Off < f.Files[j].Off })
	return nil
}

// Close allows resources associated with f to be released.
// It MUST be called exactly once for every file retrieved using get.
// Contents from any SrcFile from within f MUST NOT be used after
// Close has been called.
func (f *zipFile) Close() {
	f.wg.Done()
}

// A srcFile is a single file inside a ZipFile.
type srcFile struct {
	// Take care with the size of this struct.
	// There will be *lots* of these in memory.
	// This is why Len is a 32 bit int.
	// (Note that this means that ZipCache cannot
	// handle files inside the zip archive bigger than 2gb.)
	Name string
	Off  int64
	Len  int32
}

// Data returns the contents of s, which is a SrcFile in f.
// The contents MUST NOT be modified.
// It is not safe to use the contents after f has been Closed.
func (f *zipFile) DataFor(s *srcFile) []byte {
	return f.Data[s.Off : s.Off+int64(s.Len)]
}

func (f *srcFile) String() string {
	return fmt.Sprintf("<%s: %d+%d bytes>", f.Name, f.Off, f.Len)
}

// count returns the number of elements in c, assuming c is otherwise unused during the call to c.
// It is intended only for testing.
func (c *zipCache) count() int {
	var n int
	for i := range c.shards {
		shard := &c.shards[i]
		shard.mu.Lock()
		n += len(shard.m)
		shard.mu.Unlock()
	}
	return n
}
