package lspext

import (
	"os"
	"path"
	"time"
)

// FileInfo is the map-based implementation of FileInfo.
type FileInfo struct {
	Name_ string `json:"name"`
	Size_ int64  `json:"size"`
	Dir_  bool   `json:"dir"`
}

func (fi FileInfo) IsDir() bool        { return fi.Dir_ }
func (fi FileInfo) ModTime() time.Time { return time.Time{} }
func (fi FileInfo) Mode() os.FileMode {
	if fi.IsDir() {
		return 0755 | os.ModeDir
	}
	return 0444
}
func (fi FileInfo) Name() string     { return path.Base(fi.Name_) }
func (fi FileInfo) Size() int64      { return int64(fi.Size_) }
func (fi FileInfo) Sys() interface{} { return nil }
