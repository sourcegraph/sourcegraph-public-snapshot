package rwvfs

import (
	"errors"
	"io"
	"os"

	"golang.org/x/tools/godoc/vfs"
)

// ReadOnly returns a FileSystem whose write methods (Create, Mkdir,
// Remove) return errors. All other methods pass through to the
// read-only VFS.
func ReadOnly(fs vfs.FileSystem) FileSystem {
	return &roFS{fs}
}

type roFS struct {
	vfs.FileSystem
}

func (s *roFS) Create(path string) (io.WriteCloser, error) {
	return nil, &os.PathError{Op: "create", Path: path, Err: ErrReadOnly}
}

func (s *roFS) Mkdir(name string) error {
	return &os.PathError{Op: "mkdir", Path: name, Err: ErrReadOnly}
}

func (s *roFS) Remove(name string) error {
	return &os.PathError{Op: "remove", Path: name, Err: ErrReadOnly}
}

// ErrReadOnly occurs when a write method (Create, Mkdir, Remove) is
// called on a read-only VFS (i.e., one created by ReadOnly).
var ErrReadOnly = errors.New("read-only VFS")
