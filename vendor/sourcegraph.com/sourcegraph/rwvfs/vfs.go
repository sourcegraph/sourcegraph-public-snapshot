// Package rwvfs augments vfs to support write operations.
package rwvfs

import (
	"io"
	"os"
	"path"
	"path/filepath"
	"syscall"

	"github.com/kr/fs"

	"golang.org/x/tools/godoc/vfs"
)

type FileSystem interface {
	vfs.FileSystem

	// Create creates the named file, truncating it if it already exists.
	Create(path string) (io.WriteCloser, error)

	// Mkdir creates a new directory. If name is already a directory, Mkdir
	// returns an error (that can be detected using os.IsExist).
	Mkdir(name string) error

	// Remove removes the named file or directory.
	Remove(name string) error
}

// MkdirAllOverrider can be implemented by VFSs for which MkdirAll
// requires special behavior (e.g., VFSs without any discrete notion
// of a directory, such as Amazon S3). When MkdirAll is called on
// MkdirAllOverrider implementations, it calls the interface's method
// and passes along the error (if any).
type MkdirAllOverrider interface {
	MkdirAll(path string) error
}

func isMkdirAllOverrider(fs FileSystem) (MkdirAllOverrider, bool) {
	switch fs := fs.(type) {
	case MkdirAllOverrider:
		return fs, true
	case loggedFS:
		return isMkdirAllOverrider(fs.fs)
	case subFS:
		return isMkdirAllOverrider(fs.fs)
	case subLinkFS:
		return isMkdirAllOverrider(fs.fs)
	case subFetcherOpenerFS:
		return isMkdirAllOverrider(fs.fs)
	case walkableFS:
		return isMkdirAllOverrider(fs.FileSystem)
	case walkableLinkFS:
		return isMkdirAllOverrider(fs.FileSystem)
	}
	return nil, false
}

// MkdirAll creates a directory named path, along with any necessary parents. If
// path is already a directory, MkdirAll does nothing and returns nil.
func MkdirAll(fs FileSystem, path string) error {
	// adapted from os/MkdirAll

	path = filepath.ToSlash(path)

	if fs, ok := isMkdirAllOverrider(fs); ok {
		return fs.MkdirAll(path)
	}

	dir, err := fs.Stat(path)
	if err == nil {
		if dir.IsDir() {
			return nil
		}
		return &os.PathError{"mkdir", path, syscall.ENOTDIR}
	}

	i := len(path)
	for i > 0 && path[i-1] == '/' {
		i--
	}

	j := i
	for j > 0 && path[j-1] != '/' {
		j--
	}

	if j > 1 {
		err = MkdirAll(fs, path[0:j-1])
		if err != nil {
			return err
		}
	}

	err = fs.Mkdir(path)
	if err != nil {
		dir, err1 := fs.Lstat(path)
		if err1 == nil && dir.IsDir() {
			return nil
		}
		return err
	}
	return nil
}

// Glob returns the names of all files under prefix matching pattern or nil if
// there is no matching file. The syntax of patterns is the same as in
// path/filepath.Match.
func Glob(wfs WalkableFileSystem, prefix, pattern string) (matches []string, err error) {
	walker := fs.WalkFS(filepath.Clean(prefix), wfs)
	for walker.Step() {
		p := filepath.ToSlash(walker.Path())
		matched, err := path.Match(pattern, p)
		if err != nil {
			return nil, err
		}
		if matched {
			matches = append(matches, p)
		}
	}
	return
}

type WalkableFileSystem interface {
	FileSystem
	Join(elem ...string) string
}

// Walkable creates a walkable VFS by wrapping fs.
func Walkable(fs FileSystem) WalkableFileSystem {
	wfs := walkableFS{fs}
	switch fs.(type) {
	case LinkFS:
		return walkableLinkFS{wfs}
	case FetcherOpener:
		return walkableFetcherOpenerFS{wfs}
	default:
		return wfs
	}
}

type walkableFS struct{ FileSystem }

func (_ walkableFS) Join(elem ...string) string { return path.Join(elem...) }

type walkableLinkFS struct{ walkableFS }

func (f walkableLinkFS) ReadLink(name string) (string, error) {
	return f.FileSystem.(LinkFS).ReadLink(name)
}

func (f walkableLinkFS) Symlink(oldname, newname string) error {
	return f.FileSystem.(LinkFS).Symlink(oldname, newname)
}

var _ LinkFS = walkableLinkFS{}

type walkableFetcherOpenerFS struct{ walkableFS }

func (f walkableFetcherOpenerFS) OpenFetcher(name string) (vfs.ReadSeekCloser, error) {
	return f.FileSystem.(FetcherOpener).OpenFetcher(name)
}

var _ FetcherOpener = walkableFetcherOpenerFS{}

// A LinkFS is a filesystem that supports creating and dereferencing
// symlinks.
type LinkFS interface {
	// Symlink creates newname as a symbolic link to oldname.
	Symlink(oldname, newname string) error

	// ReadLink returns the destination of the named symbolic link.
	ReadLink(name string) (string, error)
}
