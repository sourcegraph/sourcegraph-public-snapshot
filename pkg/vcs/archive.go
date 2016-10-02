package vcs

import (
	"archive/zip"
	"bytes"
	"context"
	"os"
	"sync"

	"golang.org/x/tools/godoc/vfs"
	"golang.org/x/tools/godoc/vfs/zipfs"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/ctxvfs"
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
