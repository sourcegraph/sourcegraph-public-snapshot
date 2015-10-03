package vfsgen

import (
	"compress/gzip"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	pathpkg "path"
	"sort"
	"strconv"
	"text/template"
	"time"

	"github.com/shurcooL/httpfs/vfsutil"
)

// Generate Go code that statically implements input filesystem,
// write the output to a file specified in opt.
func Generate(input http.FileSystem, opt Options) error {
	opt.fillMissing()

	// Create output file.
	f, err := os.Create(opt.Filename)
	if err != nil {
		return err
	}
	defer f.Close()

	err = t.ExecuteTemplate(f, "Header", opt)
	if err != nil {
		return err
	}

	var toc toc
	err = findAndWriteFiles(f, input, &toc)
	if err != nil {
		return err
	}

	err = t.ExecuteTemplate(f, "DirEntries", toc.dirs)
	if err != nil {
		return err
	}

	err = t.ExecuteTemplate(f, "Trailer", toc)
	if err != nil {
		return err
	}

	// Trim any potential excess.
	cur, err := f.Seek(0, os.SEEK_CUR)
	if err != nil {
		return err
	}
	err = f.Truncate(cur)
	if err != nil {
		return err
	}

	return nil
}

type toc struct {
	dirs []*dirInfo

	HasCompressedFile bool // There's at least one compressedFile.
	HasFile           bool // There's at least one uncompressed file.
}

// fileInfo is a definition of a file.
type fileInfo struct {
	Path             string
	Name             string
	ModTime          time.Time
	UncompressedSize int64
}

// dirInfo is a definition of a directory.
type dirInfo struct {
	Path    string
	Name    string
	ModTime time.Time
	Entries []string
}

// findAndWriteFiles recursively finds all the file paths in the given directory tree.
// They are added to the given map as keys. Values will be safe function names
// for each file, which will be used when generating the output code.
func findAndWriteFiles(f *os.File, fs http.FileSystem, toc *toc) error {
	walkFn := func(path string, fi os.FileInfo, r io.ReadSeeker, err error) error {
		if err != nil {
			log.Printf("can't stat file %q: %v\n", path, err)
			return nil
		}

		switch fi.IsDir() {
		case false:
			file := &fileInfo{
				Path:             path,
				Name:             pathpkg.Base(path),
				ModTime:          fi.ModTime().UTC(),
				UncompressedSize: fi.Size(),
			}

			marker, err := f.Seek(0, os.SEEK_CUR)
			if err != nil {
				return err
			}

			// Write _vfsgen_compressedFileInfo.
			err = writeCompressedFileInfo(f, file, r)
			switch err {
			default:
				return err
			case nil:
				toc.HasCompressedFile = true
			// If compressed file is not smaller than original, revert and write original file.
			case errCompressedNotSmaller:
				_, err = r.Seek(0, os.SEEK_SET)
				if err != nil {
					return err
				}

				_, err = f.Seek(marker, os.SEEK_SET)
				if err != nil {
					return err
				}

				// Write _vfsgen_fileInfo.
				err = writeFileInfo(f, file, r)
				if err != nil {
					return err
				}
				toc.HasFile = true
			}
		case true:
			entries, err := readDirPaths(fs, path)
			if err != nil {
				return err
			}

			dir := &dirInfo{
				Path:    path,
				Name:    pathpkg.Base(path),
				ModTime: fi.ModTime().UTC(),
				Entries: entries,
			}

			toc.dirs = append(toc.dirs, dir)

			// Write _vfsgen_dirInfo.
			err = t.ExecuteTemplate(f, "DirInfo", dir)
			if err != nil {
				return err
			}
		}

		return nil
	}

	err := vfsutil.WalkFiles(fs, "/", walkFn)
	if err != nil {
		return err
	}

	return nil
}

// readDirPaths reads the directory named by dirname and returns
// a sorted list of directory paths.
func readDirPaths(fs http.FileSystem, dirname string) ([]string, error) {
	fis, err := vfsutil.ReadDir(fs, dirname)
	if err != nil {
		return nil, err
	}
	paths := make([]string, len(fis))
	for i := range fis {
		paths[i] = pathpkg.Join(dirname, fis[i].Name())
	}
	sort.Strings(paths)
	return paths, nil
}

// writeCompressedFileInfo writes _vfsgen_compressedFileInfo.
// It returns errCompressedNotSmaller if compressed file is not smaller than original.
func writeCompressedFileInfo(w io.Writer, file *fileInfo, r io.Reader) error {
	err := t.ExecuteTemplate(w, "CompressedFileInfo-Before", file)
	if err != nil {
		return err
	}
	sw := &stringWriter{Writer: w}
	gw := gzip.NewWriter(sw)
	_, err = io.Copy(gw, r)
	if err != nil {
		return err
	}
	err = gw.Close()
	if err != nil {
		return err
	}
	if sw.N >= file.UncompressedSize {
		return errCompressedNotSmaller
	}
	err = t.ExecuteTemplate(w, "CompressedFileInfo-After", file)
	if err != nil {
		return err
	}
	return nil
}

var errCompressedNotSmaller = errors.New("compressed file is not smaller than original")

// Write _vfsgen_fileInfo.
func writeFileInfo(w io.Writer, file *fileInfo, r io.Reader) error {
	err := t.ExecuteTemplate(w, "FileInfo-Before", file)
	if err != nil {
		return err
	}
	sw := &stringWriter{Writer: w}
	_, err = io.Copy(sw, r)
	if err != nil {
		return err
	}
	err = t.ExecuteTemplate(w, "FileInfo-After", file)
	if err != nil {
		return err
	}
	return nil
}

var t = template.Must(template.New("").Funcs(template.FuncMap{
	"quote": func(s string) string {
		return strconv.Quote(s)
	},
	"quoteBytes": func(b []byte) string {
		return strconv.Quote(string(b))
	},
}).Delims("⦗⦗", "⦘⦘").Parse(`⦗⦗define "Header"⦘⦘// generated by vfsgen; DO NOT EDIT

⦗⦗with .BuildTags⦘⦘// +build ⦗⦗.⦘⦘

⦗⦗end⦘⦘package ⦗⦗.PackageName⦘⦘

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	pathpkg "path"
	"time"
)

// ⦗⦗.VariableName⦘⦘ statically implements the virtual filesystem given to vfsgen as input.
var ⦗⦗.VariableName⦘⦘ = func() http.FileSystem {
	mustUnmarshalTextTime := func(text string) time.Time {
		var t time.Time
		err := t.UnmarshalText([]byte(text))
		if err != nil {
			panic(err)
		}
		return t
	}

	fs := _vfsgen_fs{
⦗⦗end⦘⦘



⦗⦗define "CompressedFileInfo-Before"⦘⦘		⦗⦗quote .Path⦘⦘: &_vfsgen_compressedFileInfo{
			name:              ⦗⦗quote .Name⦘⦘,
			modTime:           mustUnmarshalTextTime(⦗⦗quoteBytes .ModTime.MarshalText⦘⦘),
			compressedContent: []byte("⦗⦗end⦘⦘⦗⦗define "CompressedFileInfo-After"⦘⦘"),
			uncompressedSize:  ⦗⦗.UncompressedSize⦘⦘,
		},
⦗⦗end⦘⦘



⦗⦗define "FileInfo-Before"⦘⦘		⦗⦗quote .Path⦘⦘: &_vfsgen_fileInfo{
			name:    ⦗⦗quote .Name⦘⦘,
			modTime: mustUnmarshalTextTime(⦗⦗quoteBytes .ModTime.MarshalText⦘⦘),
			content: []byte("⦗⦗end⦘⦘⦗⦗define "FileInfo-After"⦘⦘"),
		},
⦗⦗end⦘⦘



⦗⦗define "DirInfo"⦘⦘		⦗⦗quote .Path⦘⦘: &_vfsgen_dirInfo{
			name:    ⦗⦗quote .Name⦘⦘,
			modTime: mustUnmarshalTextTime(⦗⦗quoteBytes .ModTime.MarshalText⦘⦘),
		},
⦗⦗end⦘⦘



⦗⦗define "DirEntries"⦘⦘	}

⦗⦗range .⦘⦘⦗⦗if .Entries⦘⦘	fs[⦗⦗quote .Path⦘⦘].(*_vfsgen_dirInfo).entries = []os.FileInfo{⦗⦗range .Entries⦘⦘
		fs[⦗⦗quote .⦘⦘].(os.FileInfo),⦗⦗end⦘⦘
	}
⦗⦗end⦘⦘⦗⦗end⦘⦘
	return fs
}()
⦗⦗end⦘⦘



⦗⦗define "Trailer"⦘⦘
type _vfsgen_fs map[string]interface{}

func (fs _vfsgen_fs) Open(path string) (http.File, error) {
	path = pathpkg.Clean("/" + path)
	f, ok := fs[path]
	if !ok {
		return nil, &os.PathError{Op: "open", Path: path, Err: os.ErrNotExist}
	}

	switch f := f.(type) {⦗⦗if .HasCompressedFile⦘⦘
	case *_vfsgen_compressedFileInfo:
		gr, err := gzip.NewReader(bytes.NewReader(f.compressedContent))
		if err != nil {
			// This should never happen because we generate the gzip bytes such that they are always valid.
			panic("unexpected error reading own gzip compressed bytes: " + err.Error())
		}
		return &_vfsgen_compressedFile{
			_vfsgen_compressedFileInfo: f,
			gr: gr,
		}, nil⦗⦗end⦘⦘⦗⦗if .HasFile⦘⦘
	case *_vfsgen_fileInfo:
		return &_vfsgen_file{
			_vfsgen_fileInfo: f,
			Reader:           bytes.NewReader(f.content),
		}, nil⦗⦗end⦘⦘
	case *_vfsgen_dirInfo:
		return &_vfsgen_dir{
			_vfsgen_dirInfo: f,
		}, nil
	default:
		// This should never happen because we generate only the above types.
		panic(fmt.Sprintf("unexpected type %T", f))
	}
}
⦗⦗if .HasCompressedFile⦘⦘
// _vfsgen_compressedFileInfo is a static definition of a gzip compressed file.
type _vfsgen_compressedFileInfo struct {
	name              string
	modTime           time.Time
	compressedContent []byte
	uncompressedSize  int64
}

func (f *_vfsgen_compressedFileInfo) Readdir(count int) ([]os.FileInfo, error) {
	return nil, fmt.Errorf("cannot Readdir from file %s", f.name)
}
func (f *_vfsgen_compressedFileInfo) Stat() (os.FileInfo, error) { return f, nil }

func (f *_vfsgen_compressedFileInfo) GzipBytes() []byte {
	return f.compressedContent
}

func (f *_vfsgen_compressedFileInfo) Name() string       { return f.name }
func (f *_vfsgen_compressedFileInfo) Size() int64        { return f.uncompressedSize }
func (f *_vfsgen_compressedFileInfo) Mode() os.FileMode  { return 0444 }
func (f *_vfsgen_compressedFileInfo) ModTime() time.Time { return f.modTime }
func (f *_vfsgen_compressedFileInfo) IsDir() bool        { return false }
func (f *_vfsgen_compressedFileInfo) Sys() interface{}   { return nil }

// _vfsgen_compressedFile is an opened compressedFile instance.
type _vfsgen_compressedFile struct {
	*_vfsgen_compressedFileInfo
	gr      *gzip.Reader
	grPos   int64 // Actual gr uncompressed position.
	seekPos int64 // Seek uncompressed position.
}

func (f *_vfsgen_compressedFile) Read(p []byte) (n int, err error) {
	if f.grPos > f.seekPos {
		// Rewind to beginning.
		err = f.gr.Reset(bytes.NewReader(f._vfsgen_compressedFileInfo.compressedContent))
		if err != nil {
			return 0, err
		}
		f.grPos = 0
	}
	if f.grPos < f.seekPos {
		// Fast-forward.
		_, err = io.ReadFull(f.gr, make([]byte, f.seekPos-f.grPos))
		if err != nil {
			return 0, err
		}
		f.grPos = f.seekPos
	}
	n, err = f.gr.Read(p)
	f.grPos += int64(n)
	f.seekPos = f.grPos
	return n, err
}
func (f *_vfsgen_compressedFile) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case os.SEEK_SET:
		f.seekPos = 0 + offset
	case os.SEEK_CUR:
		f.seekPos += offset
	case os.SEEK_END:
		f.seekPos = f._vfsgen_compressedFileInfo.uncompressedSize + offset
	}
	return f.seekPos, nil
}
func (f *_vfsgen_compressedFile) Close() error {
	return f.gr.Close()
}
⦗⦗else⦘⦘
// We already imported "compress/gzip", but ended up not using it. Avoid unused import error.
var _ = gzip.Reader
⦗⦗end⦘⦘⦗⦗if .HasFile⦘⦘
// _vfsgen_fileInfo is a static definition of an uncompressed file (because it's not worth gzip compressing).
type _vfsgen_fileInfo struct {
	name    string
	modTime time.Time
	content []byte
}

func (f *_vfsgen_fileInfo) Readdir(count int) ([]os.FileInfo, error) {
	return nil, fmt.Errorf("cannot Readdir from file %s", f.name)
}
func (f *_vfsgen_fileInfo) Stat() (os.FileInfo, error) { return f, nil }

func (f *_vfsgen_fileInfo) NotWorthGzipCompressing() {}

func (f *_vfsgen_fileInfo) Name() string       { return f.name }
func (f *_vfsgen_fileInfo) Size() int64        { return int64(len(f.content)) }
func (f *_vfsgen_fileInfo) Mode() os.FileMode  { return 0444 }
func (f *_vfsgen_fileInfo) ModTime() time.Time { return f.modTime }
func (f *_vfsgen_fileInfo) IsDir() bool        { return false }
func (f *_vfsgen_fileInfo) Sys() interface{}   { return nil }

// _vfsgen_file is an opened file instance.
type _vfsgen_file struct {
	*_vfsgen_fileInfo
	*bytes.Reader
}

func (f *_vfsgen_file) Close() error {
	return nil
}
⦗⦗end⦘⦘
// _vfsgen_dirInfo is a static definition of a directory.
type _vfsgen_dirInfo struct {
	name    string
	modTime time.Time
	entries []os.FileInfo
}

func (d *_vfsgen_dirInfo) Read([]byte) (int, error) {
	return 0, fmt.Errorf("cannot Read from directory %s", d.name)
}
func (d *_vfsgen_dirInfo) Close() error               { return nil }
func (d *_vfsgen_dirInfo) Stat() (os.FileInfo, error) { return d, nil }

func (d *_vfsgen_dirInfo) Name() string       { return d.name }
func (d *_vfsgen_dirInfo) Size() int64        { return 0 }
func (d *_vfsgen_dirInfo) Mode() os.FileMode  { return 0755 | os.ModeDir }
func (d *_vfsgen_dirInfo) ModTime() time.Time { return d.modTime }
func (d *_vfsgen_dirInfo) IsDir() bool        { return true }
func (d *_vfsgen_dirInfo) Sys() interface{}   { return nil }

// _vfsgen_dir is an opened dir instance.
type _vfsgen_dir struct {
	*_vfsgen_dirInfo
	pos int // Position within entries for Seek and Readdir.
}

func (d *_vfsgen_dir) Seek(offset int64, whence int) (int64, error) {
	if offset == 0 && whence == os.SEEK_SET {
		d.pos = 0
		return 0, nil
	}
	return 0, fmt.Errorf("unsupported Seek in directory %s", d._vfsgen_dirInfo.name)
}

func (d *_vfsgen_dir) Readdir(count int) ([]os.FileInfo, error) {
	if d.pos >= len(d._vfsgen_dirInfo.entries) && count > 0 {
		return nil, io.EOF
	}
	if count <= 0 || count > len(d._vfsgen_dirInfo.entries)-d.pos {
		count = len(d._vfsgen_dirInfo.entries) - d.pos
	}
	e := d._vfsgen_dirInfo.entries[d.pos : d.pos+count]
	d.pos += count
	return e, nil
}
⦗⦗end⦘⦘`))
