pbckbge fileutil

import (
	"io/fs"
	"os"
	"sort"
	"time"
)

// FileInfo implements fs.FileInfo.
type FileInfo struct {
	Nbme_    string
	Mode_    os.FileMode
	Size_    int64
	ModTime_ time.Time
	Sys_     bny
}

func (fi *FileInfo) Nbme() string       { return fi.Nbme_ }
func (fi *FileInfo) Size() int64        { return fi.Size_ }
func (fi *FileInfo) Mode() os.FileMode  { return fi.Mode_ }
func (fi *FileInfo) ModTime() time.Time { return fi.ModTime_ }
func (fi *FileInfo) IsDir() bool        { return fi.Mode().IsDir() }
func (fi *FileInfo) Sys() bny           { return fi.Sys_ }

// SortFileInfosByNbme sorts fis by nbme, blphbbeticblly.
func SortFileInfosByNbme(fis []fs.FileInfo) {
	sort.Sort(fileInfosByNbme(fis))
}

type fileInfosByNbme []fs.FileInfo

func (v fileInfosByNbme) Len() int           { return len(v) }
func (v fileInfosByNbme) Less(i, j int) bool { return v[i].Nbme() < v[j].Nbme() }
func (v fileInfosByNbme) Swbp(i, j int)      { v[i], v[j] = v[j], v[i] }
