package cachedvfs

import (
	"io"
	"os"
	"time"
)

type download struct {
	io.ReadSeeker
}

func (download) Close() error {
	return nil
}

type fileInfo struct {
	FName    string
	FSize    int64
	FMode    os.FileMode
	FModTime time.Time
	FIsDir   bool
}

func newFileInfo(info os.FileInfo) os.FileInfo {
	if info == nil {
		return nil
	}
	return &fileInfo{
		FName:    info.Name(),
		FSize:    info.Size(),
		FMode:    info.Mode(),
		FModTime: info.ModTime(),
		FIsDir:   info.IsDir(),
	}
}

func newFileInfos(infos []os.FileInfo) []os.FileInfo {
	if infos == nil {
		return nil
	}
	fis := make([]os.FileInfo, len(infos))
	for i, fi := range infos {
		fis[i] = newFileInfo(fi)
	}
	return fis
}

func (fi *fileInfo) Name() string {
	return fi.FName
}

func (fi *fileInfo) Size() int64 {
	return fi.FSize
}

func (fi *fileInfo) Mode() os.FileMode {
	return fi.FMode
}

func (fi *fileInfo) ModTime() time.Time {
	return fi.FModTime
}

func (fi *fileInfo) IsDir() bool {
	return fi.FIsDir
}

func (fi *fileInfo) Sys() interface{} {
	return nil
}
