// Package godocfs implements vfs.FileSystem using a http.FileSystem.
package godocfs

import (
	"net/http"
	"os"

	"github.com/shurcooL/httpfs/vfsutil"
	"golang.org/x/tools/godoc/vfs"
)

// New returns a vfs.FileSystem adapter for the provided http.FileSystem.
func New(fs http.FileSystem) vfs.FileSystem {
	return &godocFS{fs: fs}
}

type godocFS struct {
	fs http.FileSystem
}

func (v *godocFS) Open(name string) (vfs.ReadSeekCloser, error) {
	return v.fs.Open(name)
}

func (v *godocFS) Lstat(path string) (os.FileInfo, error) {
	return v.Stat(path)
}

func (v *godocFS) Stat(path string) (os.FileInfo, error) {
	return vfsutil.Stat(v.fs, path)
}

func (v *godocFS) ReadDir(path string) ([]os.FileInfo, error) {
	return vfsutil.ReadDir(v.fs, path)
}

func (v *godocFS) String() string { return "godocfs" }
