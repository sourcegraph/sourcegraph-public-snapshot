package graphqlbackend

import (
	"context"
	"os"
	"path"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/graphqlbackend/externallink"
)

// gitTreeEntryResolver resolves an entry in a Git tree in a repository. The entry can be any Git
// object type that is valid in a tree.
type gitTreeEntryResolver struct {
	commit *gitCommitResolver

	path string      // this tree entry's path (relative to the root)
	stat os.FileInfo // this tree entry's file info

	isRecursive bool // whether entries is populated recursively (otherwise just current level of hierarchy)
}

func (r *gitTreeEntryResolver) Path() string { return r.path }
func (r *gitTreeEntryResolver) Name() string { return path.Base(r.path) }

func (r *gitTreeEntryResolver) ToGitTree() (*gitTreeEntryResolver, bool) { return r, true }
func (r *gitTreeEntryResolver) ToGitBlob() (*gitTreeEntryResolver, bool) { return r, true }

func (r *gitTreeEntryResolver) Commit() *gitCommitResolver { return r.commit }

func (r *gitTreeEntryResolver) Repository() *repositoryResolver { return r.commit.repo }

func (r *gitTreeEntryResolver) IsRecursive() bool { return r.isRecursive }

func (r *gitTreeEntryResolver) URL() string { return r.urlPath(r.commit.repoRevURL()) }

func (r *gitTreeEntryResolver) CanonicalURL() string {
	return r.urlPath(r.commit.canonicalRepoRevURL())
}

func (r *gitTreeEntryResolver) urlPath(prefix string) string {
	if r.IsRoot() {
		return prefix
	}

	url := prefix + "/-/"
	if r.IsDirectory() {
		url += "tree"
	} else {
		url += "blob"
	}
	return url + "/" + r.path
}

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
