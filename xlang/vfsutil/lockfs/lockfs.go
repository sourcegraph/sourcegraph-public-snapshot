// Package lockfs provides a virtual file system wrapper that locks a
// mutex during operations.
package lockfs

import (
	"fmt"
	"os"
	"sync"

	"golang.org/x/tools/godoc/vfs"
)

// New creates a new locking file system wrapper around fs. The
// contents of files must be immutable, since it has no way of
// synchronizing access to the vfs.ReadSeekCloser from Open after Open
// returns.
func New(mu *sync.Mutex, fs vfs.FileSystem) vfs.FileSystem {
	return &lockFS{mu, fs}
}

type lockFS struct {
	mu *sync.Mutex
	fs vfs.FileSystem
}

func (fs *lockFS) Open(path string) (vfs.ReadSeekCloser, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	return fs.fs.Open(path)
}

func (fs *lockFS) Lstat(path string) (os.FileInfo, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	return fs.fs.Lstat(path)
}

func (fs *lockFS) Stat(path string) (os.FileInfo, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	return fs.fs.Stat(path)
}

func (fs *lockFS) ReadDir(path string) ([]os.FileInfo, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	return fs.fs.ReadDir(path)
}

func (fs *lockFS) String() string {
	return fmt.Sprintf("lockFS(%s)", fs.fs.String())
}
