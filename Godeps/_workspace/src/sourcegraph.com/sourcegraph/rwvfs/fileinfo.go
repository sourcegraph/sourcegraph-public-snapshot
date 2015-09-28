package rwvfs

import (
	"encoding/json"
	"os"
	"path"
	"time"
)

func newDirInfo(name string) os.FileInfo {
	return fileInfo{name: path.Base(name), dir: true}
}

func newSymlinkInfo(name string) os.FileInfo {
	return fileInfo{name: path.Base(name), symlink: true}
}

func newFileInfo(name, contents string) os.FileInfo {
	return fileInfo{name: path.Base(name), size: int64(len(contents))}
}

// fileInfo implements os.FileInfo.
type fileInfo struct {
	name    string
	size    int64
	dir     bool
	modTime time.Time
	symlink bool
}

func (fi fileInfo) IsDir() bool        { return fi.dir }
func (fi fileInfo) ModTime() time.Time { return fi.modTime }
func (fi fileInfo) Mode() os.FileMode {
	if fi.IsDir() {
		return 0755 | os.ModeDir
	}
	if fi.symlink {
		return 0755 | os.ModeSymlink
	}
	return 0444
}
func (fi fileInfo) Name() string     { return path.Base(fi.name) }
func (fi fileInfo) Size() int64      { return fi.size }
func (fi fileInfo) Sys() interface{} { return nil }

type fileInfoJSON struct{ os.FileInfo }

func (fi fileInfoJSON) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"name":    fi.Name(),
		"size":    fi.Size(),
		"dir":     fi.Mode().IsDir(),
		"modTime": fi.ModTime(),
		"symlink": fi.Mode()&os.ModeSymlink > 0,
	})
}

func (fi *fileInfoJSON) UnmarshalJSON(b []byte) error {
	var fi2 struct {
		Name    string    `json:"name"`
		Size    int64     `json:"size"`
		Dir     bool      `json:"dir"`
		ModTime time.Time `json:"modTime"`
		Symlink bool      `json:"symlink"`
	}
	if err := json.Unmarshal(b, &fi2); err != nil {
		return err
	}
	*fi = fileInfoJSON{
		fileInfo{
			name:    fi2.Name,
			size:    fi2.Size,
			dir:     fi2.Dir,
			modTime: fi2.ModTime,
			symlink: fi2.Symlink,
		},
	}
	return nil
}
