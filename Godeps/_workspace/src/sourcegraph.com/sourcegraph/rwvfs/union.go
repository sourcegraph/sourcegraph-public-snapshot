package rwvfs

import (
	"fmt"
	"io"
	"os"
	pathpkg "path"

	"golang.org/x/tools/godoc/vfs"
)

// Union returns a new FileSystem which is the union of the provided file systems.
// For read operations, vfs.NameSpace is used and its behavior is inherited.
// Write operations are applied to the first file system which contains the parent directory.
// The union file system is not thread-safe. Concurrent access to itself and/or
// its underlying file systems requires synchronization.
func Union(fileSystems ...FileSystem) FileSystem {
	if len(fileSystems) == 0 {
		return ReadOnly(Map(nil))
	}
	if len(fileSystems) == 1 {
		return fileSystems[0]
	}

	ns := vfs.NameSpace{}
	for _, fs := range fileSystems {
		ns.Bind("/", fs, "/", vfs.BindAfter)
	}
	return &unionFS{
		NameSpace:   ns,
		fileSystems: fileSystems,
	}
}

type unionFS struct {
	vfs.NameSpace
	fileSystems []FileSystem
}

func (fs *unionFS) resolve(path string) FileSystem {
	parent := pathpkg.Dir(path)
	for _, fs2 := range fs.fileSystems {
		if _, err := fs2.Stat(parent); err != nil && os.IsNotExist(err) {
			continue
		}
		return fs2
	}
	return fs.fileSystems[0]
}

// Create creates the named file, truncating it if it already exists.
func (fs *unionFS) Create(path string) (io.WriteCloser, error) {
	return fs.resolve(path).Create(path)
}

// Mkdir creates a new directory. If name is already a directory, Mkdir
// returns an error (that can be detected using os.IsExist).
func (fs *unionFS) Mkdir(name string) error {
	return fs.resolve(name).Mkdir(name)
}

// Remove removes the named file or directory.
func (fs *unionFS) Remove(name string) error {
	return fs.resolve(name).Remove(name)
}

func (fs *unionFS) String() string {
	return fmt.Sprintf("union(%v)", fs.fileSystems)
}
