package rwvfs

import (
	"bytes"
	"io"
	"os"
	pathpkg "path"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/tools/godoc/vfs"
	"golang.org/x/tools/godoc/vfs/mapfs"
)

// Map returns a new FileSystem from the provided map. Map keys should be
// forward slash-separated pathnames and not contain a leading slash.
func Map(m map[string]string) FileSystem {
	fs := mapFS{
		m:          m,
		dirs:       map[string]struct{}{"/": struct{}{}},
		mu:         new(sync.RWMutex),
		FileSystem: mapfs.New(m),
	}

	// Create initial dirs.
	for path := range m {
		if err := MkdirAll(fs, filepath.Dir(path)); err != nil {
			panic(err.Error())
		}
	}

	return fs
}

type mapFS struct {
	m    map[string]string
	dirs map[string]struct{}
	mu   *sync.RWMutex
	vfs.FileSystem
}

func (mfs mapFS) Open(path string) (vfs.ReadSeekCloser, error) {
	mfs.mu.RLock()
	defer mfs.mu.RUnlock()

	path = slash(path)
	f, err := mfs.FileSystem.Open(path)
	if err != nil {
		return nil, &os.PathError{Op: "open", Path: path, Err: err}
	}
	return f, nil
}

func (mfs mapFS) Create(path string) (io.WriteCloser, error) {
	mfs.mu.Lock()
	defer mfs.mu.Unlock()

	// Mimic behavior of OS filesystem: truncate to empty string upon creation;
	// immediately update string values with writes.
	path = slash(path)
	mfs.m[noslash(path)] = ""
	return &mapFile{m: mfs.m, path: noslash(path), fsMu: mfs.mu}, nil
}

func noslash(p string) string {
	p = slash(p)
	if p == "/" {
		return "."
	}
	return strings.TrimPrefix(p, "/")
}

// slashdir returns path.Dir(p), but special-cases paths not beginning
// with a slash to be in the root.
func slashdir(p string) string {
	p = filepath.ToSlash(p)
	d := pathpkg.Dir(p)
	if d == "." {
		return "/"
	}
	if strings.HasPrefix(p, "/") {
		return d
	}
	return "/" + d
}

func slash(p string) string {
	if p == "." {
		return "/"
	}
	return "/" + strings.TrimPrefix(filepath.ToSlash(p), "/")
}

type mapFile struct {
	buf  bytes.Buffer
	m    map[string]string
	path string
	fsMu *sync.RWMutex
}

func (f *mapFile) Write(p []byte) (int, error) {
	return f.buf.Write(p)
}

func (f *mapFile) Close() error {
	if f.m == nil {
		// duplicate closes are noop
		return nil
	}

	f.fsMu.Lock()
	defer f.fsMu.Unlock()

	f.m[f.path] = f.buf.String()
	f.buf.Reset()
	f.m = nil
	return nil
}

func (mfs mapFS) lstat(p string) (os.FileInfo, error) {
	mfs.mu.RLock()
	defer mfs.mu.RUnlock()

	// proxy mapfs.mapFS.Lstat to not return errors for empty directories
	// created with Mkdir
	p = slash(p)
	fi, err := mfs.FileSystem.Lstat(p)
	if os.IsNotExist(err) {
		_, ok := mfs.dirs[p]
		if ok {
			return fileInfo{name: pathpkg.Base(p), dir: true}, nil
		}
	}
	return fi, err
}

func (mfs mapFS) Lstat(p string) (os.FileInfo, error) {
	fi, err := mfs.lstat(p)
	if err != nil {
		err = &os.PathError{Op: "lstat", Path: p, Err: err}
	}
	return fi, err
}

func (mfs mapFS) Stat(p string) (os.FileInfo, error) {
	fi, err := mfs.lstat(p)
	if err != nil {
		err = &os.PathError{Op: "stat", Path: p, Err: err}
	}
	return fi, err
}

func (mfs mapFS) ReadDir(p string) ([]os.FileInfo, error) {
	mfs.mu.RLock()
	defer mfs.mu.RUnlock()

	// proxy mapfs.mapFS.ReadDir to not return errors for empty directories
	// created with Mkdir
	p = slash(p)
	fis, err := mfs.FileSystem.ReadDir(p)
	if os.IsNotExist(err) {
		_, ok := mfs.dirs[p]
		if ok {
			// return a list of subdirs and files (the underlying ReadDir impl
			// fails here because it thinks the directories don't exist).
			fis = nil
			for dir, _ := range mfs.dirs {
				if (p != "/" && filepath.Dir(dir) == p) || (p == "/" && filepath.Dir(dir) == "." && dir != "." && dir != "") {
					fis = append(fis, newDirInfo(dir))
				}
			}
			for fn, b := range mfs.m {
				if slashdir(fn) == "/"+p {
					fis = append(fis, newFileInfo(fn, b))
				}
			}
			return fis, nil
		}
	}
	return fis, err
}

func fileInfoNames(fis []os.FileInfo) []string {
	names := make([]string, len(fis))
	for i, fi := range fis {
		names[i] = fi.Name()
	}
	return names
}

func (mfs mapFS) Mkdir(name string) error {
	name = slash(name)
	if slashdir(name) != slash(name) { // don't check for root dir's parent
		if _, err := mfs.Stat(slashdir(name)); err != nil {
			if osErr, ok := err.(*os.PathError); ok && osErr != nil {
				osErr.Op = "mkdir"
				osErr.Path = name
			}
			return err
		}
	}
	fi, _ := mfs.Stat(name)
	if fi != nil {
		return &os.PathError{Op: "mkdir", Path: name, Err: os.ErrExist}
	}

	mfs.mu.Lock()
	defer mfs.mu.Unlock()

	mfs.dirs[slash(name)] = struct{}{}
	return nil
}

func (mfs mapFS) Remove(name string) error {
	mfs.mu.Lock()
	defer mfs.mu.Unlock()

	name = slash(name)
	delete(mfs.dirs, name)
	delete(mfs.m, noslash(name))
	return nil
}
