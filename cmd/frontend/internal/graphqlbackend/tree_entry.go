package graphqlbackend

import (
	"context"
	"os"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/graphqlbackend/externallink"
)

func (r *gitTreeEntryResolver) IsDirectory() bool { return r.stat.Mode().IsDir() }

func (r *gitTreeEntryResolver) ExternalURLs(ctx context.Context) ([]*externallink.Resolver, error) {
	return externallink.FileOrDir(ctx, r.commit.repo.repo, r.commit.inputRevOrImmutableRev(), r.path, r.stat.Mode().IsDir())
}

func createFileInfo(path string, isDir bool) os.FileInfo {
	return fileInfo{path: path, isDir: isDir}
}

type fileInfo struct {
	path  string
	isDir bool
}

func (f fileInfo) Name() string { return f.path }
func (f fileInfo) Size() int64  { return 0 }
func (f fileInfo) IsDir() bool  { return f.isDir }
func (f fileInfo) Mode() os.FileMode {
	if f.IsDir() {
		return os.ModeDir
	}
	return 0
}
func (f fileInfo) ModTime() time.Time { return time.Now() }
func (f fileInfo) Sys() interface{}   { return interface{}(nil) }
