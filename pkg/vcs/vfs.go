package vcs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/ctxvfs"
)

type fileSystem struct {
	repo   Repository
	commit CommitID
}

// FileSystem returns a vfs.FileSystem for repo at commit.
func FileSystem(repo Repository, commit CommitID) ctxvfs.FileSystem {
	return fileSystem{repo: repo, commit: commit}
}

func (fs fileSystem) Open(ctx context.Context, name string) (ctxvfs.ReadSeekCloser, error) {
	b, err := fs.repo.ReadFile(ctx, fs.commit, name)
	if err != nil {
		return nil, err
	}
	return nopCloser{ReadSeeker: bytes.NewReader(b)}, nil
}

func (fs fileSystem) Lstat(ctx context.Context, name string) (os.FileInfo, error) {
	return fs.repo.Lstat(ctx, fs.commit, name)
}

func (fs fileSystem) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	return fs.repo.Stat(ctx, fs.commit, name)
}

func (fs fileSystem) ReadDir(ctx context.Context, name string) ([]os.FileInfo, error) {
	return fs.repo.ReadDir(ctx, fs.commit, name, false)
}

func (fs fileSystem) String() string {
	return fmt.Sprintf("%s at commit %s (cmd)", fs.repo, fs.commit)
}

type nopCloser struct {
	io.ReadSeeker
}

func (nc nopCloser) Close() error { return nil }
