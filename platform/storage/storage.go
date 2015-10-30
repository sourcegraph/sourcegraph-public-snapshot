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
	fmt.Stringer
	io.Reader
	io.Writer
	io.Seeker
	io.Closer

	// Name returns the name of the file as presented to Open.
	Name() string

	// Truncate changes the size of the file. It does not change the I/O offset.
	Truncate(size int64) error
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
// If the repo is a valid repo URI, storage is considered "local" to that
// repository. Otherwise, storage is considered "global" (i.e. shared across
// all repositories).
func Namespace(ctx context.Context, appName string, repo string) FileSystem {
	return &fileSystem{
		ctx:     ctx,
		client:  sourcegraph.NewClientFromContext(ctx),
		appName: appName,
		repo:    repo,
	}
}

// isStorageError tells if the error is non-nil and non-zero.
func isStorageError(err *sourcegraph.StorageError) bool {
	return err != nil && *err != (sourcegraph.StorageError{})
}

// storageError converts a gRPC StorageError type into it's equivilent Go error
// type. If the err parameter is nil, a nil error is returned.
func storageError(err *sourcegraph.StorageError) error {
	if !isStorageError(err) {
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

// fileInfo wraps a gRPC StorageFileInfo type and provides a os.FileInfo
// implementation.
type fileInfo struct {
	i sourcegraph.StorageFileInfo
}

func (fi fileInfo) Name() string     { return fi.i.Name }
func (fi fileInfo) Size() int64      { return fi.i.Size_ }
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
