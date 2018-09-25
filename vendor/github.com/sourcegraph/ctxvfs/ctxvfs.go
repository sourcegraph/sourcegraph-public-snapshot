// Package ctxvfs defines a virtual file system interface whose
// methods accept a context.Context parameter. It is otherwise similar
// to golang.org/x/tools/godoc/vfs.
package ctxvfs

import (
	"context"
	"io"
	"io/ioutil"
	"os"
)

// The FileSystem interface specifies the methods used to access the
// file system.
type FileSystem interface {
	Open(ctx context.Context, name string) (ReadSeekCloser, error)
	Lstat(ctx context.Context, path string) (os.FileInfo, error)
	Stat(ctx context.Context, path string) (os.FileInfo, error)
	ReadDir(ctx context.Context, path string) ([]os.FileInfo, error)
	String() string
}

// A ReadSeekCloser can Read, Seek, and Close.
type ReadSeekCloser interface {
	io.Reader
	io.Seeker
	io.Closer
}

// ReadFile reads the file named by path from fs and returns the contents.
func ReadFile(ctx context.Context, fs FileSystem, path string) ([]byte, error) {
	rc, err := fs.Open(ctx, path)
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return ioutil.ReadAll(rc)
}
