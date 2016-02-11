package vcs

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"golang.org/x/tools/godoc/vfs"
)

type fileSystem struct {
	repo   Repository
	commit CommitID
}

// FileSystem returns a vfs.FileSystem for repo at commit.
func FileSystem(repo Repository, commit CommitID) vfs.FileSystem {
	return fileSystem{repo: repo, commit: commit}
}

func (fs fileSystem) Open(name string) (vfs.ReadSeekCloser, error) {
	b, err := fs.repo.ReadFile(fs.commit, name)
	if err != nil {
		return nil, err
	}
	return nopCloser{ReadSeeker: bytes.NewReader(b)}, nil
}

func (fs fileSystem) Lstat(name string) (os.FileInfo, error) {
	return fs.repo.Lstat(fs.commit, name)
}

func (fs fileSystem) Stat(name string) (os.FileInfo, error) {
	return fs.repo.Stat(fs.commit, name)
}

func (fs fileSystem) ReadDir(name string) ([]os.FileInfo, error) {
	return fs.repo.ReadDir(fs.commit, name, false)
}

func (fs fileSystem) String() string {
	return fmt.Sprintf("git repository %s commit %s (cmd)", fs.repo.GitRootDir(), fs.commit)
}

type nopCloser struct {
	io.ReadSeeker
}

func (nc nopCloser) Close() error { return nil }
