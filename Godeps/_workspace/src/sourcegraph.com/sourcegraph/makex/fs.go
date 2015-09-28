package makex

import (
	"path/filepath"

	"sourcegraph.com/sourcegraph/rwvfs"
)

// A FileSystem is a file system that can be read, written, and walked. Given an
// existing rwvfs.FileSystem, use NewFileSystem to add the Join method (which
// will use the current OS's filepath.Separator).
type FileSystem interface {
	rwvfs.FileSystem
	Join(elem ...string) string
}

// NewFileSystem returns a FileSystem with Join method (which will use the
// current OS's filepath.Separator).
func NewFileSystem(fs rwvfs.FileSystem) FileSystem {
	return walkableRWVFS{fs}
}

type walkableRWVFS struct{ rwvfs.FileSystem }

func (_ walkableRWVFS) Join(elem ...string) string { return filepath.Join(elem...) }
