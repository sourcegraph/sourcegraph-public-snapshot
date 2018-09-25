package ctxvfs

import (
	"bytes"
	"context"
	"os"
	pathpkg "path"
)

func SingleFileOverlay(fs FileSystem, path string, contents []byte) FileSystem {
	return &singleFileOverlay{
		FileSystem: fs,
		path:       pathpkg.Clean(path),
		name:       pathpkg.Base(path),
		pathDir:    pathpkg.Dir(path),
		contents:   contents,
	}
}

// singleFileOverlay wraps a VFS and adds a single file.
type singleFileOverlay struct {
	FileSystem
	path     string
	name     string
	pathDir  string
	contents []byte
}

func (fs *singleFileOverlay) Open(ctx context.Context, path string) (ReadSeekCloser, error) {
	if path == fs.path {
		return nopCloser{bytes.NewReader(fs.contents)}, nil
	}
	return fs.FileSystem.Open(ctx, path)
}

func (fs *singleFileOverlay) Stat(ctx context.Context, path string) (os.FileInfo, error) {
	if path == fs.path {
		return fileInfo{fs.name, int64(len(fs.contents))}, nil
	}
	return fs.FileSystem.Stat(ctx, path)
}

func (fs *singleFileOverlay) ReadDir(ctx context.Context, path string) ([]os.FileInfo, error) {
	fis, err := fs.FileSystem.ReadDir(ctx, path)
	if err == nil && path == pathpkg.Dir(fs.path) {
		fis = append(fis, fileInfo{fs.name, int64(len(fs.contents))})
	}
	return fis, err
}

func (fs *singleFileOverlay) Lstat(ctx context.Context, path string) (os.FileInfo, error) {
	return fs.Stat(ctx, path)
}
