package vfs

import (
	"io"
	"os"
)

// File represents a single file in the storage system.
type File interface {
	io.Reader
	io.Writer
	io.Seeker
	io.Closer

	// Name returns the name of the file as presented to Open.
	Name() string

	// Truncate changes the size of the file. It does not change the I/O offset.
	Truncate(size int64) error
}

// FileSystem represents the storage system.
type FileSystem interface {
	Create(name string) (File, error)
	Remove(name string) error
	RemoveAll(name string) error
	Open(name string) (File, error)
	Stat(path string) (os.FileInfo, error)
	ReadDir(path string) ([]os.FileInfo, error)
	String() string
}
