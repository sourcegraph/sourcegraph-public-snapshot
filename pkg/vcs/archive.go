package vcs

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/sourcegraph/ctxvfs"
	"golang.org/x/tools/godoc/vfs"
	"golang.org/x/tools/godoc/vfs/zipfs"
)

// An Archiver is a repository that can produce a .zip archive of
// itself at a given commit ID.
type Archiver interface {
	// Archive returns a .zip archive of the repo at the given commit ID.
	Archive(context.Context, CommitID) ([]byte, error)
}

// ArchiveFileSystem returns a virtual file system backed by a .zip
// archive of a Git tree (in the common case, the root tree of a Git
// repository at a specific commit). The treeish is a Git object ID
// that refers to a tree; it can be a commit ID, a tree ID, and so
// on. For consistency, callers should generally use full SHAs, not
// rev specs like branch names, etc.
//
// Unlike vcs.FileSystem (which makes individual execs/calls for each
// FS operation), ArchiveFileSystem fetches the full .zip archive
// initially and then can satisfy FS operations nearly instantly in
// memory.
func ArchiveFileSystem(repo Archiver, treeish string) ctxvfs.FileSystem {
	return &archiveFS{repo: repo, treeish: treeish}
}

type archiveFS struct {
	repo    Archiver
	treeish string

	once sync.Once
	err  error          // the error encountered during the Archive call (if any)
	fs   vfs.FileSystem // the zipfs virtual file system
}

// fetchOrWait initiates the fetch if it has not yet
// started. Otherwise it waits for it to finish.
func (fs *archiveFS) fetchOrWait(ctx context.Context) error {
	fs.once.Do(func() {
		fs.err = fs.fetch(ctx)
	})
	return fs.err
}

func (fs *archiveFS) fetch(ctx context.Context) (err error) {
	data, err := fs.repo.Archive(ctx, CommitID(fs.treeish))
	if err != nil {
		return err
	}

	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return err
	}
	fs.fs = zipfs.New(&zip.ReadCloser{Reader: *zr}, "")
	return nil
}

func (fs *archiveFS) Open(ctx context.Context, name string) (ctxvfs.ReadSeekCloser, error) {
	if err := fs.fetchOrWait(ctx); err != nil {
		return nil, err
	}
	return fs.fs.Open(name)
}

func (fs *archiveFS) Lstat(ctx context.Context, path string) (os.FileInfo, error) {
	if err := fs.fetchOrWait(ctx); err != nil {
		return nil, err
	}
	return fs.fs.Lstat(path)
}

func (fs *archiveFS) Stat(ctx context.Context, path string) (os.FileInfo, error) {
	if err := fs.fetchOrWait(ctx); err != nil {
		return nil, err
	}
	return fs.fs.Stat(path)
}

func (fs *archiveFS) ReadDir(ctx context.Context, path string) ([]os.FileInfo, error) {
	if err := fs.fetchOrWait(ctx); err != nil {
		return nil, err
	}
	return fs.fs.ReadDir(path)
}

func (fs *archiveFS) String() string { return "archiveFS(" + fs.fs.String() + ")" }

// FastVFS wraps a repo and presents a fast VFS interface that calls
// the repo directly until the ArchiveFileSystem is ready, at which
// point it calls into the archive for fast, in-memory access.
func FastVFS(repo interface {
	Repository
	Archiver
}, treeish string) ctxvfs.FileSystem {
	return &fastVFS{
		archive:     ArchiveFileSystem(repo, treeish).(*archiveFS),
		archiveDone: make(chan struct{}),
		direct:      FileSystem(repo, CommitID(treeish)),
	}
}

type fastVFS struct {
	mu             sync.Mutex
	archiveStarted bool
	archive        *archiveFS
	archiveErr     error
	archiveDone    chan struct{}

	direct ctxvfs.FileSystem
}

// chooseImpl returns the direct-to-gitserver VFS implementation until
// the zip archive is fetched. The zip archive is preferred once it's
// available, since it's much faster than roundtripping to gitserver.
func (fs *fastVFS) chooseImpl(ctx context.Context) (ctxvfs.FileSystem, error) {
	fs.mu.Lock()
	if !fs.archiveStarted {
		// Run this once, but don't ever block on it.
		go func() {
			fs.archiveErr = fs.archive.fetchOrWait(ctx)
			close(fs.archiveDone)
		}()
		fs.archiveStarted = true
	}
	fs.mu.Unlock()

	select {
	case <-fs.archiveDone:
		return fs.archive, fs.archiveErr
	default:
		return fs.direct, nil
	}
}

func (fs *fastVFS) Open(ctx context.Context, name string) (ctxvfs.ReadSeekCloser, error) {
	v, err := fs.chooseImpl(ctx)
	if err != nil {
		return nil, err
	}
	return v.Open(ctx, name)
}

func (fs *fastVFS) Lstat(ctx context.Context, name string) (os.FileInfo, error) {
	v, err := fs.chooseImpl(ctx)
	if err != nil {
		return nil, err
	}
	return v.Lstat(ctx, name)
}

func (fs *fastVFS) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	v, err := fs.chooseImpl(ctx)
	if err != nil {
		return nil, err
	}
	return v.Stat(ctx, name)
}

func (fs *fastVFS) ReadDir(ctx context.Context, name string) ([]os.FileInfo, error) {
	v, err := fs.chooseImpl(ctx)
	if err != nil {
		return nil, err
	}
	return v.ReadDir(ctx, name)
}

func (fs *fastVFS) ListAllFiles(ctx context.Context) ([]string, error) {
	v, err := fs.chooseImpl(ctx)
	if err != nil {
		return nil, err
	}
	if v, ok := v.(interface {
		ListAllFiles(context.Context) ([]string, error)
	}); ok {
		return v.ListAllFiles(ctx)
	}

	fis, err := fs.archive.repo.(Repository).ReadDir(ctx, CommitID(fs.archive.treeish), "", true)
	if err != nil {
		return nil, err
	}
	paths := make([]string, 0, len(fis))
	for _, fi := range fis {
		if fi.Mode().IsRegular() {
			paths = append(paths, fi.Name())
		}
	}
	return paths, nil
}

func (fs *fastVFS) String() string {
	return fmt.Sprintf("FastVFS(%s at commit %s)", fs.archive.repo, fs.archive.treeish)
}
