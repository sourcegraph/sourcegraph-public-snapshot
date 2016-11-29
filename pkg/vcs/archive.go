package vcs

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"

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
	err  error // the error encountered during the Archive call (if any)
	zr   *zip.Reader
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
	fs.zr = zr
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

func (fs *archiveFS) ListAllFiles(ctx context.Context) ([]string, error) {
	if err := fs.fetchOrWait(ctx); err != nil {
		return nil, err
	}

	filenames := make([]string, 0, len(fs.zr.File))
	for _, f := range fs.zr.File {
		if f.Mode().IsRegular() {
			filenames = append(filenames, f.Name)
		}
	}
	return filenames, nil
}

func (fs *archiveFS) String() string { return "archiveFS(" + fs.fs.String() + ")" }

// FastVFS wraps a repo and presents a fast VFS interface that calls the repo
// directly until the ArchiveFileSystem is ready, at which point it calls into
// the archive for fast, in-memory access. ctx is required so we can kick of
// background fetching straight away.
func FastVFS(ctx context.Context, repo interface {
	Repository
	Archiver
}, treeish string) (ctxvfs.FileSystem, error) {
	fs := &fastVFS{
		archive: ArchiveFileSystem(repo, treeish).(*archiveFS),
	}
	go func() {
		defer atomic.StoreUint32(&fs.archiveDone, 1)
		fs.archiveErr = fs.archive.fetchOrWait(ctx)
	}()
	// We only set direct now, since we may have to do a blocking resolve
	// of treeish to make it an absolute commit. Absolute commit is
	// required for caching/perf reasons. We also wanted to kick off the
	// archive fetch before this lookup. The length being 40 is the same
	// check used by gitcmd.
	var commit CommitID
	if len(treeish) == 40 {
		commit = CommitID(treeish)
	} else {
		var err error
		commit, err = repo.ResolveRevision(ctx, treeish)
		if err != nil {
			return nil, err
		}
	}
	fs.direct = FileSystem(repo, commit)
	return fs, nil
}

type fastVFS struct {
	archive     *archiveFS
	archiveErr  error
	archiveDone uint32

	direct ctxvfs.FileSystem
}

// chooseImpl returns the direct-to-gitserver VFS implementation until
// the zip archive is fetched. The zip archive is preferred once it's
// available, since it's much faster than roundtripping to gitserver.
func (fs *fastVFS) chooseImpl(ctx context.Context) (ctxvfs.FileSystem, error) {
	if atomic.LoadUint32(&fs.archiveDone) == 1 {
		return fs.archive, fs.archiveErr
	}
	return fs.direct, nil
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
