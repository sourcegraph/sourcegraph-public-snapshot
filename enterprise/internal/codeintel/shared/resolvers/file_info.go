package sharedresolvers

import (
	"io/fs"
	"os"
	"time"
)

type fileInfo struct {
	path  string
	size  int64
	isDir bool
}

func CreateFileInfo(path string, isDir bool) fs.FileInfo {
	return fileInfo{path: path, isDir: isDir}
}

func (f fileInfo) Name() string { return f.path }
func (f fileInfo) Size() int64  { return f.size }
func (f fileInfo) IsDir() bool  { return f.isDir }
func (f fileInfo) Mode() os.FileMode {
	if f.IsDir() {
		return os.ModeDir
	}
	return 0
}
func (f fileInfo) ModTime() time.Time { return time.Now() }
func (f fileInfo) Sys() any           { return any(nil) }
