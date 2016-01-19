package gopherjs_http

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	pathpkg "path"
	"strings"
	"time"
)

// NewFS returns an http.FileSystem that is exactly like source, except all Go packages are compiled to JavaScript with GopherJS.
//
// For example:
//
// 	/mypkg/foo.go
// 	/mypkg/bar.go
//
// Become replaced with:
//
// 	/mypkg/mypkg.js
//
// Where mypkg.js is the result of building mypkg with GopherJS.
func NewFS(source http.FileSystem) http.FileSystem {
	return &gopherJSFS{source: source}
}

type gopherJSFS struct {
	source http.FileSystem
}

func (fs *gopherJSFS) Open(path string) (http.File, error) {
	switch dir, file := pathpkg.Split(path); {
	case file == pathpkg.Base(dir)+".js":
		return fs.compileGoPackage(dir)
	default:
		return fs.openSource(path)
	}
}

func (fs *gopherJSFS) openSource(path string) (http.File, error) {
	f, err := fs.source.Open(path)
	if err != nil {
		return nil, err
	}

	fi, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}

	switch {
	// Files with .go and ".inc.js" extensions are consumed and no longer exist
	// in output filesystem.
	case !fi.IsDir() && pathpkg.Ext(fi.Name()) == ".go":
		fallthrough
	case !fi.IsDir() && strings.HasSuffix(fi.Name(), ".inc.js"):
		f.Close()
		return nil, &os.PathError{Op: "open", Path: path, Err: os.ErrNotExist}
	case !fi.IsDir():
		return f, nil
	}
	defer f.Close()

	fis, err := f.Readdir(0)
	if err != nil {
		return nil, err
	}

	// Include all subfolders, non-.go files.
	var entries []os.FileInfo
	var haveGo []os.FileInfo
	for _, fi := range fis {
		switch {
		case !fi.IsDir() && pathpkg.Ext(fi.Name()) == ".go":
			haveGo = append(haveGo, fi)
		case !fi.IsDir() && strings.HasSuffix(fi.Name(), ".inc.js"):
			// TODO: Handle ".inc.js" files correctly.
			entries = append(entries, fi)
		default:
			entries = append(entries, fi)
		}
	}

	// If it has any .go files, present the Go package compiled with GopherJS as an additional virtual file.
	if len(haveGo) > 0 {
		entries = append(entries, &file{
			name:    fi.Name() + ".js",
			size:    0,           // TODO.
			modTime: time.Time{}, // TODO.
		})
	}

	return &dir{
		name:    fi.Name(),
		entries: entries,
		modTime: fi.ModTime(),
	}, nil
}

func (fs *gopherJSFS) compileGoPackage(dir string) (http.File, error) {
	f, err := fs.source.Open(dir)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	if !fi.IsDir() {
		return nil, fmt.Errorf("%s is not a dir", dir)
	}

	fis, err := f.Readdir(0)
	if err != nil {
		return nil, err
	}

	var goFiles []os.FileInfo
	for _, f := range fis {
		if !f.IsDir() && pathpkg.Ext(f.Name()) == ".go" {
			goFiles = append(goFiles, f)
		}
	}
	if len(goFiles) == 0 {
		return nil, fmt.Errorf("%s has no .go files", dir)
	}

	// TODO: Clean this up.
	{
		name := pathpkg.Base(dir) + ".js"

		var names []string
		var goReaders []io.Reader
		var goClosers []io.Closer
		for _, goFile := range goFiles {
			file, err := fs.source.Open(pathpkg.Join(dir, goFile.Name()))
			if err != nil {
				return nil, err
			}
			names = append(names, goFile.Name())
			goReaders = append(goReaders, file)
			goClosers = append(goClosers, file)
		}

		fmt.Printf("REBUILDING SOURCE for: %s using %+v\n", name, names)
		content := []byte(handleJsError(goReadersToJs(names, goReaders)))

		for _, closer := range goClosers {
			closer.Close()
		}

		return &file{
			name:    name,
			size:    int64(len(content)),
			modTime: time.Now(),
			Reader:  bytes.NewReader(content),
		}, nil
	}
}

// file is an opened file instance.
type file struct {
	name    string
	modTime time.Time
	size    int64
	*bytes.Reader
}

func (f *file) Readdir(count int) ([]os.FileInfo, error) {
	return nil, fmt.Errorf("cannot Readdir from file %s", f.name)
}
func (f *file) Stat() (os.FileInfo, error) { return f, nil }

func (f *file) Name() string       { return f.name }
func (f *file) Size() int64        { return f.size }
func (f *file) Mode() os.FileMode  { return 0444 }
func (f *file) ModTime() time.Time { return f.modTime }
func (f *file) IsDir() bool        { return false }
func (f *file) Sys() interface{}   { return nil }

func (f *file) Close() error {
	return nil
}

// dir is an opened dir instance.
type dir struct {
	name    string
	modTime time.Time
	entries []os.FileInfo
	pos     int // Position within entries for Seek and Readdir.
}

func (d *dir) Read([]byte) (int, error) {
	return 0, fmt.Errorf("cannot Read from directory %s", d.name)
}
func (d *dir) Close() error               { return nil }
func (d *dir) Stat() (os.FileInfo, error) { return d, nil }

func (d *dir) Name() string       { return d.name }
func (d *dir) Size() int64        { return 0 }
func (d *dir) Mode() os.FileMode  { return 0755 | os.ModeDir }
func (d *dir) ModTime() time.Time { return d.modTime }
func (d *dir) IsDir() bool        { return true }
func (d *dir) Sys() interface{}   { return nil }

func (d *dir) Seek(offset int64, whence int) (int64, error) {
	if offset == 0 && whence == os.SEEK_SET {
		d.pos = 0
		return 0, nil
	}
	return 0, fmt.Errorf("unsupported Seek in directory %s", d.name)
}

func (d *dir) Readdir(count int) ([]os.FileInfo, error) {
	if d.pos >= len(d.entries) && count > 0 {
		return nil, io.EOF
	}
	if count <= 0 || count > len(d.entries)-d.pos {
		count = len(d.entries) - d.pos
	}
	e := d.entries[d.pos : d.pos+count]
	d.pos += count
	return e, nil
}
