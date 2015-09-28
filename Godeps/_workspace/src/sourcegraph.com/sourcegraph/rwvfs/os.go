package rwvfs

import (
	"errors"
	"fmt"
	"io"
	"os"
	pathpkg "path"
	"path/filepath"
	"strings"

	"golang.org/x/tools/godoc/vfs"
)

// OS returns an implementation of FileSystem reading from the tree rooted at
// root.
func OS(root string) FileSystem {
	return OSPerm(root, 0666, 0755)
}

func OSPerm(root string, filePerm, dirPerm os.FileMode) FileSystem {
	return &osFS{
		root:       root,
		filePerm:   filePerm,
		dirPerm:    dirPerm,
		FileSystem: vfs.OS(root),
	}
}

type osFS struct {
	root     string
	filePerm os.FileMode
	dirPerm  os.FileMode
	parents  bool
	vfs.FileSystem
}

// CreateParentDirs makes this FS use os.MkdirAll instead of os.Mkdir,
// which means that parent dirs are created as needed and calling
// Mkdir on an existing dir does not return an error.
func (fs *osFS) CreateParentDirs(v bool) {
	fs.parents = v
}

// resolve is from golang.org/x/tools/godoc/vfs.
func (fs *osFS) resolve(path string) string {
	// Clean the path so that it cannot possibly begin with ../.
	// If it did, the result of filepath.Join would be outside the
	// tree rooted at root.  We probably won't ever see a path
	// with .. in it, but be safe anyway.
	path = pathpkg.Clean("/" + path)

	return filepath.Join(string(fs.root), path)
}

func (fs *osFS) ReadLink(name string) (string, error) {
	dst, err := os.Readlink(fs.resolve(name))
	if err != nil {
		return "", err
	}
	if !filepath.IsAbs(dst) {
		dst = filepath.Join(fs.resolve(name), dst)
	}
	if !strings.HasPrefix(dst, fs.root+string(os.PathSeparator)) && dst != fs.root {
		return dst, ErrOutsideRoot
	}
	return filepath.Rel(fs.root, dst)
}

func (fs *osFS) Symlink(oldname, newname string) error {
	return os.Symlink(fs.resolve(oldname), fs.resolve(newname))
}

// ErrOutsideRoot occurs when a symlink refers to a path that is not in the current VFS.
var ErrOutsideRoot = errors.New("link destination is outside of filesystem")

// Create opens the file at path for writing, creating the file if it doesn't
// exist and truncating it otherwise.
func (fs *osFS) Create(path string) (io.WriteCloser, error) {
	f, err := os.OpenFile(fs.resolve(path), os.O_RDWR|os.O_CREATE|os.O_TRUNC, fs.filePerm)
	if err != nil {
		return nil, err
	}

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if fi.IsDir() {
		return nil, fmt.Errorf("Open: %s is a directory", path)
	}

	return f, nil
}

func (fs *osFS) Mkdir(name string) error {
	if fs.parents {
		return os.MkdirAll(fs.resolve(name), fs.dirPerm)
	}
	return os.Mkdir(fs.resolve(name), fs.dirPerm)
}

func (fs *osFS) Remove(name string) error {
	return os.Remove(fs.resolve(name))
}
