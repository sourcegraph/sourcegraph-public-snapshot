package vcsclient

import (
	"os"
	"time"
)

// Stat returns the FileInfo structure describing the tree entry.
func (e *TreeEntry) Stat() (os.FileInfo, error) {
	// We can't just make TreeEntry implement os.FileInfo, because then we'd
	// have to rename its fields that conflict with FileInfo's method names
	// (Name and Size).

	var mode os.FileMode
	switch e.Type {
	case DirEntry:
		mode |= os.ModeDir
	case SymlinkEntry:
		mode |= os.ModeSymlink
	}

	return &fileInfo{
		name:  e.Name,
		mode:  mode,
		size:  int64(e.Size_),
		mtime: e.ModTime.Time(),
	}, nil
}

type fileInfo struct {
	name  string
	mode  os.FileMode
	size  int64
	mtime time.Time
}

func (fi *fileInfo) Name() string       { return fi.name }
func (fi *fileInfo) Size() int64        { return fi.size }
func (fi *fileInfo) Mode() os.FileMode  { return fi.mode }
func (fi *fileInfo) ModTime() time.Time { return fi.mtime }
func (fi *fileInfo) IsDir() bool        { return fi.Mode().IsDir() }
func (fi *fileInfo) Sys() interface{}   { return nil }

type TreeEntriesByTypeByName []*TreeEntry

func (v TreeEntriesByTypeByName) Len() int      { return len(v) }
func (v TreeEntriesByTypeByName) Swap(i, j int) { v[i], v[j] = v[j], v[i] }
func (v TreeEntriesByTypeByName) Less(i, j int) bool {
	// Sort dirs before everything else.
	if v[i].Type == DirEntry && v[j].Type != DirEntry {
		return true
	}
	if v[i].Type != DirEntry && v[j].Type == DirEntry {
		return false
	}
	return v[i].Name < v[j].Name
}
