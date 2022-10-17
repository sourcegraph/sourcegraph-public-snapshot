package mapfs

import (
	"io"
	"io/fs"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type mapFSFile struct {
	name string
	size int64
	io.ReadCloser
}

func (f *mapFSFile) Stat() (fs.FileInfo, error) {
	return &mapFSFileEntry{name: f.name, size: f.size}, nil
}

func (d *mapFSFile) ReadDir(count int) ([]fs.DirEntry, error) {
	return nil, &fs.PathError{Op: "read", Path: d.name, Err: errors.New("not a directory")}
}

type mapFSFileEntry struct {
	name string
	size int64
}

func (e *mapFSFileEntry) Name() string       { return e.name }
func (e *mapFSFileEntry) Size() int64        { return e.size }
func (e *mapFSFileEntry) Mode() fs.FileMode  { return fs.ModePerm }
func (e *mapFSFileEntry) ModTime() time.Time { return time.Time{} }
func (e *mapFSFileEntry) IsDir() bool        { return e.Mode().IsDir() }
func (e *mapFSFileEntry) Sys() any           { return nil }
