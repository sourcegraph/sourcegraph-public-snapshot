package fileutil

import (
	"io/fs"
	"os"
	"sort"
	"time"
)

// FileInfo implements fs.FileInfo.
type FileInfo struct {
	Name_    string
	Mode_    os.FileMode
	Size_    int64
	ModTime_ time.Time
	Sys_     any
}

func (fi *FileInfo) Name() string       { return fi.Name_ }
func (fi *FileInfo) Size() int64        { return fi.Size_ }
func (fi *FileInfo) Mode() os.FileMode  { return fi.Mode_ }
func (fi *FileInfo) ModTime() time.Time { return fi.ModTime_ }
func (fi *FileInfo) IsDir() bool        { return fi.Mode().IsDir() }
func (fi *FileInfo) Sys() any           { return fi.Sys_ }

// SortFileInfosByName sorts fis by name, alphabetically.
func SortFileInfosByName(fis []fs.FileInfo) {
	sort.Sort(fileInfosByName(fis))
}

type fileInfosByName []fs.FileInfo

func (v fileInfosByName) Len() int           { return len(v) }
func (v fileInfosByName) Less(i, j int) bool { return v[i].Name() < v[j].Name() }
func (v fileInfosByName) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }
