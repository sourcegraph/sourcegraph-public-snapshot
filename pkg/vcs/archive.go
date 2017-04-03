package vcs

import (
	"archive/zip"
	"bytes"
	"context"
	"os"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
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
func ArchiveFileSystem(repo Archiver, treeish string) *ArchiveFS {
	return &ArchiveFS{repo: repo, treeish: treeish}
}

// ArchiveFS is a ctxvfs.FileSystem backed by an Archiver.
type ArchiveFS struct {
	repo    Archiver
	treeish string

	once sync.Once
	err  error // the error encountered during the Archive call (if any)
	zr   *zip.Reader
	fs   vfs.FileSystem // the zipfs virtual file system
}

// fetchOrWait initiates the fetch if it has not yet
// started. Otherwise it waits for it to finish.
func (fs *ArchiveFS) fetchOrWait(ctx context.Context) error {
	fs.once.Do(func() {
		fs.err = fs.fetch(ctx)
	})
	return fs.err
}

func (fs *ArchiveFS) fetch(ctx context.Context) (err error) {
	data, err := fs.repo.Archive(ctx, CommitID(fs.treeish))
	if err != nil {
		return err
	}
	gitserverBytes.Add(float64(len(data)))

	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return err
	}
	fs.zr = zr
	fs.fs = zipfs.New(&zip.ReadCloser{Reader: *zr}, "")
	return nil
}

func (fs *ArchiveFS) Open(ctx context.Context, name string) (ctxvfs.ReadSeekCloser, error) {
	if err := fs.fetchOrWait(ctx); err != nil {
		return nil, err
	}
	return fs.fs.Open(name)
}

func (fs *ArchiveFS) Lstat(ctx context.Context, path string) (os.FileInfo, error) {
	if err := fs.fetchOrWait(ctx); err != nil {
		return nil, err
	}
	return fs.fs.Lstat(path)
}

func (fs *ArchiveFS) Stat(ctx context.Context, path string) (os.FileInfo, error) {
	if err := fs.fetchOrWait(ctx); err != nil {
		return nil, err
	}
	return fs.fs.Stat(path)
}

func (fs *ArchiveFS) ReadDir(ctx context.Context, path string) ([]os.FileInfo, error) {
	if err := fs.fetchOrWait(ctx); err != nil {
		return nil, err
	}
	return fs.fs.ReadDir(path)
}

func (fs *ArchiveFS) ListAllFiles(ctx context.Context) ([]string, error) {
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

func (fs *ArchiveFS) String() string { return "ArchiveFS(" + fs.fs.String() + ")" }

var gitserverBytes = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: "src",
	Subsystem: "vfs",
	Name:      "gitserver_bytes_total",
	Help:      "Total number of bytes read into memory by ArchiveFileSystem.",
})

func init() {
	prometheus.MustRegister(gitserverBytes)
}
