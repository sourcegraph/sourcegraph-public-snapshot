package rwvfs

import (
	"os"

	"github.com/kr/fs"
)

// StatAllRecursive recursively stats all files and dirs in fs,
// starting at path and descending. The Name methods of the returned
// FileInfos returns their full path, not just their filename.
func StatAllRecursive(path string, wfs WalkableFileSystem) ([]os.FileInfo, error) {
	var fis []os.FileInfo
	w := fs.WalkFS(path, wfs)
	for w.Step() {
		if err := w.Err(); err != nil {
			return nil, err
		}
		fis = append(fis, treeFileInfo{w.Path(), w.Stat()})
	}
	return fis, nil
}

type treeFileInfo struct {
	path string
	os.FileInfo
}

func (fi treeFileInfo) Name() string { return fi.path }
