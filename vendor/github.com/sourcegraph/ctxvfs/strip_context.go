package ctxvfs

import (
	"context"
	"os"

	"golang.org/x/tools/godoc/vfs"
)

// StripContext is an adapter for using a ctxvfs.FileSystem (whose
// interface methods take a context.Context parameter) as a
// vfs.FileSystem (whose interface methods DON'T take a
// context.Context parameter). A context.Background() value is passed
// to the underlying ctxvfs.FileSystem.
//
// This should only be used temporarily while you are transitioning
// code to use ctxvfs.FileSystem.
func StripContext(fs FileSystem) vfs.FileSystem {
	return &stripContext{fs}
}

var stripContextCtx = context.Background()

type stripContext struct{ fs FileSystem }

func (fs *stripContext) Open(name string) (vfs.ReadSeekCloser, error) {
	return fs.fs.Open(stripContextCtx, name)
}

func (fs *stripContext) Stat(path string) (os.FileInfo, error) {
	return fs.fs.Stat(stripContextCtx, path)
}

func (fs *stripContext) Lstat(path string) (os.FileInfo, error) {
	return fs.fs.Lstat(stripContextCtx, path)
}

func (fs *stripContext) ReadDir(path string) ([]os.FileInfo, error) {
	return fs.fs.ReadDir(stripContextCtx, path)
}

func (fs *stripContext) RootType(path string) vfs.RootType {
	return ""
}

func (fs *stripContext) String() string { return "stripContext(" + fs.fs.String() + ")" }
