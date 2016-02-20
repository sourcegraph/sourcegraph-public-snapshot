package rwvfs

import (
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/tools/godoc/vfs"
)

// Sub returns an implementation of FileSystem mounted at prefix on the
// underlying fs. If fs doesn't have an existing directory at prefix, you can
// can call Mkdir("/") on the new filesystem to create it.
func Sub(fs FileSystem, prefix string) FileSystem {
	subfs := subFS{fs, prefix}
	switch ufs := fs.(type) {
	case LinkFS:
		return subLinkFS{subfs}
	case FetcherOpener:
		return subFetcherOpenerFS{subfs}
	case walkableFS:
		return Sub(ufs.FileSystem, prefix)
	case walkableFetcherOpenerFS:
		return Sub(ufs.FileSystem, prefix)
	case loggedFS:
		return Sub(ufs.fs, prefix)
	default:
		return subfs
	}
}

type subFS struct {
	fs     FileSystem
	prefix string
}

var _ FileSystem = subFS{}

type subLinkFS struct{ subFS }

var _ LinkFS = subLinkFS{}

type subFetcherOpenerFS struct{ subFS }

var _ FetcherOpener = subFetcherOpenerFS{}

func (s subFS) resolve(path string) string {
	return filepath.ToSlash(filepath.Join(s.prefix, strings.TrimPrefix(filepath.ToSlash(path), "/")))
}

func (s subFS) stripPrefix(path string) string {
	return "/" + strings.TrimPrefix(strings.TrimPrefix(strings.TrimPrefix(filepath.ToSlash(path), "/"), strings.TrimPrefix(filepath.ToSlash(s.prefix), "/")), "/")
}

func (s subFS) Lstat(path string) (os.FileInfo, error) {
	fi, err := s.fs.Lstat(s.resolve(path))
	if err != nil {
		return nil, s.resolvePathError(err)
	}
	return fi, nil
}

func (s subFS) Stat(path string) (os.FileInfo, error) {
	fi, err := s.fs.Stat(s.resolve(path))
	if err != nil {
		return nil, s.resolvePathError(err)
	}
	return fi, nil
}

func (s subLinkFS) ReadLink(name string) (string, error) {
	dst, err := s.fs.(LinkFS).ReadLink(s.resolve(name))
	if err != nil {
		return dst, s.resolvePathError(err)
	}
	return filepath.Rel(s.prefix, dst)
}

func (s subLinkFS) Symlink(oldname, newname string) error {
	if err := s.fs.(LinkFS).Symlink(s.resolve(oldname), s.resolve(newname)); err != nil {
		return s.resolvePathError(err)
	}
	return nil
}

func (s subFS) ReadDir(path string) ([]os.FileInfo, error) {
	entries, err := s.fs.ReadDir(s.resolve(path))
	if err != nil {
		return nil, s.resolvePathError(err)
	}
	return entries, nil
}

func (s subFS) String() string { return "sub(" + s.fs.String() + ", " + s.prefix + ")" }

func (s subFS) Open(name string) (vfs.ReadSeekCloser, error) {
	f, err := s.fs.Open(s.resolve(name))
	if err != nil {
		return nil, s.resolvePathError(err)
	}
	return f, nil
}

func (s subFetcherOpenerFS) OpenFetcher(name string) (vfs.ReadSeekCloser, error) {
	f, err := s.fs.(FetcherOpener).OpenFetcher(s.resolve(name))
	if err != nil {
		return nil, s.resolvePathError(err)
	}
	return f, nil
}

func (s subFS) Create(path string) (io.WriteCloser, error) {
	f, err := s.fs.Create(s.resolve(path))
	if err != nil {
		return nil, s.resolvePathError(err)
	}
	return f, nil
}

func (s subFS) Mkdir(name string) error {
	err := s.mkdir(name)
	if os.IsNotExist(err) {
		// Automatically create subFS's prefix dirs they don't exist.
		if osErr, ok := err.(*os.PathError); ok && path.Dir(osErr.Path) == "/" {
			if err := MkdirAll(s.fs, s.prefix); err != nil {
				return s.resolvePathError(err)
			}
			return s.mkdir(name)
		}
	}
	return err
}

func (s subFS) mkdir(name string) error {
	if err := s.fs.Mkdir(s.resolve(name)); err != nil {
		return s.resolvePathError(err)
	}
	return nil
}

func (s subFS) Remove(name string) error {
	if err := s.fs.Remove(s.resolve(name)); err != nil {
		return s.resolvePathError(err)
	}
	return nil
}

func (s subFS) resolvePathError(err error) error {
	if err == nil {
		return nil
	}
	if perr, ok := err.(*os.PathError); ok && perr != nil {
		perr.Path = s.stripPrefix(perr.Path)
	}
	return err
}
