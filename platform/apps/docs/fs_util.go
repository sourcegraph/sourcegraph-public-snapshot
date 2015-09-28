package docs

import (
	"os"
	"strings"
	"time"

	"github.com/spf13/afero"
	"golang.org/x/tools/godoc/vfs"
)

// aferoVFS is a VFS-to-afero-VFS adapter for filesystems.
type aferoVFS struct{ vfs.FileSystem }

func trim(path string) string { return strings.TrimPrefix(path, "/") }

func (fs aferoVFS) Open(name string) (afero.File, error) {
	f, err := fs.FileSystem.Open(trim(name))
	if err != nil {
		return nil, err
	}
	return aferoFile{fs, trim(name), f}, nil
}

func (aferoVFS) Name() string                                 { return "aferoVFS" }
func (aferoVFS) Create(name string) (afero.File, error)       { return nil, nil }
func (aferoVFS) Mkdir(path string, perm os.FileMode) error    { return nil }
func (aferoVFS) MkdirAll(path string, perm os.FileMode) error { return nil }
func (aferoVFS) Remove(path string) error                     { return nil }
func (aferoVFS) RemoveAll(path string) error                  { return nil }
func (aferoVFS) Rename(oldname, newname string) error         { return nil }
func (aferoVFS) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	panic("OpenFile")
}
func (aferoVFS) Chmod(name string, mode os.FileMode) error                   { return nil }
func (aferoVFS) Chtimes(name string, atime time.Time, mtime time.Time) error { return nil }

// aferoFile is a VFS-to-afero-VFS adapter for files.
type aferoFile struct {
	fs   aferoVFS
	name string
	vfs.ReadSeekCloser
}

func (f aferoFile) Stat() (os.FileInfo, error) {
	return f.fs.Stat(trim(f.name))
}

func (f aferoFile) Readdir(count int) ([]os.FileInfo, error) {
	if count > 0 {
		panic("count > 0")
	}
	return f.fs.ReadDir(trim(f.name))
}

func (f aferoFile) Readdirnames(n int) ([]string, error) {
	fis, err := f.Readdir(n)
	if err != nil {
		return nil, err
	}
	names := make([]string, len(fis))
	for i, fi := range fis {
		names[i] = fi.Name()
	}
	return names, nil
}

func (aferoFile) Name() string                                   { return "aferoFile" }
func (aferoFile) ReadAt(p []byte, off int64) (n int, err error)  { panic("ReadAt") }
func (aferoFile) Write(p []byte) (n int, err error)              { return 0, nil }
func (aferoFile) WriteAt(p []byte, off int64) (n int, err error) { return 0, nil }
func (aferoFile) WriteString(s string) (ret int, err error)      { return 0, nil }
func (aferoFile) Truncate(size int64) error                      { return nil }
