package vfsutil

import "golang.org/x/tools/godoc/vfs"

// WalkableFileSystem is a virtual filesystem that is walkable.
//
// The package github.com/kr/fs, among other things, requires a
// walkable FS.
type WalkableFileSystem interface {
	vfs.FileSystem

	// Join joins any number of path elements into a single path,
	// adding a separator if necessary. The result is Cleaned; in
	// particular, all empty strings are ignored.
	//
	// The separator is FileSystem specific.
	Join(elem ...string) string
}

// Walkable wraps fs with a type that implements WalkableFileSystem
// with the given func used as the Join method.
//
// The join func should be filepath.Join if the VFS's underlying FS is
// the host's filesystem (so that, e.g., backslash is used on
// Windows), or path.Join if the VFS uses slash-delimited paths.
func Walkable(fs vfs.FileSystem, join func(elem ...string) string) WalkableFileSystem {
	return walkableFileSystem{fs, join}
}

type walkableFileSystem struct {
	vfs.FileSystem
	join func(elem ...string) string
}

func (fs walkableFileSystem) Join(elem ...string) string { return fs.join(elem...) }
