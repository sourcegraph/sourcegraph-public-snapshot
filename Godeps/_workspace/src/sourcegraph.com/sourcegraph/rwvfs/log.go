package rwvfs

import (
	"io"
	"log"
	"os"

	"golang.org/x/tools/godoc/vfs"
)

// Logged creates a new VFS wrapper that logs and times calls to an
// underlying VFS.
func Logged(log *log.Logger, fs FileSystem) FileSystem {
	return loggedFS{log, fs}
}

type loggedFS struct {
	log *log.Logger
	fs  FileSystem
}

func (s loggedFS) Lstat(path string) (os.FileInfo, error) {
	s.log.Printf("Lstat(%q)", path)
	return s.fs.Lstat(path)
}

func (s loggedFS) Stat(path string) (os.FileInfo, error) {
	s.log.Printf("Stat(%q)", path)
	return s.fs.Stat(path)
}

func (s loggedFS) ReadDir(path string) ([]os.FileInfo, error) {
	s.log.Printf("ReadDir(%q)", path)
	return s.fs.ReadDir(path)
}

func (s loggedFS) String() string { return "logged(" + s.fs.String() + ")" }

func (s loggedFS) Open(name string) (vfs.ReadSeekCloser, error) {
	s.log.Printf("Open(%q)", name)
	return s.fs.Open(name)
}

func (s loggedFS) Create(path string) (io.WriteCloser, error) {
	s.log.Printf("Create(%q)", path)
	return s.fs.Create(path)
}

func (s loggedFS) Mkdir(name string) error {
	s.log.Printf("Mkdir(%q)", name)
	return s.fs.Mkdir(name)
}

func (s loggedFS) Remove(name string) error {
	s.log.Printf("Remove(%q)", name)
	return s.fs.Remove(name)
}
