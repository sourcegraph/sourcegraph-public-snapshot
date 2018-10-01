package ctxvfs

import (
	"context"
	"fmt"
	"os"
	"sync"
)

// Sync creates a new file system wrapper around fs that locks a mutex
// during its operations.
//
// The contents of files must be immutable, since it has no way of
// synchronizing access to the ReadSeekCloser from Open after Open
// returns.
func Sync(mu *sync.Mutex, fs FileSystem) FileSystem {
	return &syncFS{mu, fs}
}

type syncFS struct {
	mu *sync.Mutex
	fs FileSystem
}

func (fs *syncFS) Open(ctx context.Context, name string) (ReadSeekCloser, error) {
	fs.mu.Lock()
	f, err := fs.fs.Open(ctx, name)
	fs.mu.Unlock()
	return f, err
}

func (fs *syncFS) Lstat(ctx context.Context, path string) (os.FileInfo, error) {
	fs.mu.Lock()
	fi, err := fs.fs.Lstat(ctx, path)
	fs.mu.Unlock()
	return fi, err
}

func (fs *syncFS) Stat(ctx context.Context, path string) (os.FileInfo, error) {
	fs.mu.Lock()
	fi, err := fs.fs.Stat(ctx, path)
	fs.mu.Unlock()
	return fi, err
}

func (fs *syncFS) ReadDir(ctx context.Context, path string) ([]os.FileInfo, error) {
	fs.mu.Lock()
	fi, err := fs.fs.ReadDir(ctx, path)
	fs.mu.Unlock()
	return fi, err
}

func (fs *syncFS) String() string {
	return fmt.Sprintf("lock(%s)", fs.fs.String())
}
