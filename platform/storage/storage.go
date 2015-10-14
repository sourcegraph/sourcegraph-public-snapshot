package storage

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

// File represents a single file in the storage system.
type File interface {
	Name() string
	fmt.Stringer
	io.Reader
	io.Writer
	io.Seeker
	io.Closer
}

// FileSystem represents the storage system.
type FileSystem interface {
	fmt.Stringer
	Create(name string) (File, error)
	Remove(name string) error
	RemoveAll(name string) error
	Open(name string) (File, error)
	Lstat(path string) (os.FileInfo, error)
	ReadDir(path string) ([]os.FileInfo, error)
}

// Namespace returns a storage system for the given namespace. The returned
// filesystem cannot read/write outside of the namespace provided here.
//
// appName is the name of the application whose data you are trying to
// read/write, applications may read and write to eachother's data assuming the
// admin has not restricted such access.
//
// If the RepoSpec is nil, storage is considered "global" (i.e. shared across
// all repositories, instead of local to the specified one).
func Namespace(ctx context.Context, appName string, repo *sourcegraph.RepoSpec) FileSystem {
	return &fileSystem{
		ctx:     ctx,
		client:  sourcegraph.NewClientFromContext(ctx),
		appName: appName,
		repo:    repo,
	}
}

func storageError(err *sourcegraph.StorageError) error {
	if err == nil {
		return nil
	}
	switch err.Code {
	case sourcegraph.StorageError_EOF:
		return io.EOF
	case sourcegraph.StorageError_NOT_EXIST:
		return os.ErrNotExist
	case sourcegraph.StorageError_PERMISSION:
		return os.ErrPermission
	default:
		return errors.New(err.Message)
	}
}

type fileInfo struct {
	i sourcegraph.StorageFileInfo
}

func (fi fileInfo) Name() string     { return fi.i.Name }
func (fi fileInfo) Size() int64      { return fi.i.Size }
func (fi fileInfo) IsDir() bool      { return fi.i.IsDir }
func (fi fileInfo) Sys() interface{} { return nil }

func (fi fileInfo) Mode() os.FileMode {
	if fi.i.IsDir {
		return os.ModeDir
	}
	return os.ModePerm
}

func (fi fileInfo) ModTime() time.Time {
	return fi.i.ModTime.Time()
}
