package graphqlbackend

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/externallink"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/conf/reposource"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
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

func (r *gitTreeEntryResolver) URL() string {
	if submodule := r.Submodule(); submodule != nil {
		repoURI, err := cloneURLToURI(submodule.URL())
		if err != nil {
			log15.Error("Failed to resolve submodule repository URI from clone URL", "cloneURL", submodule.URL())
			return ""
		}
		return "/" + repoURI + "@" + submodule.Commit()
	}
	return r.urlPath(r.commit.repoRevURL())
}

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

func (r *gitTreeEntryResolver) Submodule() *gitSubmoduleResolver {
	if submoduleInfo, ok := r.stat.Sys().(git.Submodule); ok {
		return &gitSubmoduleResolver{submodule: submoduleInfo}
	}
	return nil
}

func cloneURLToURI(cloneURL string) (string, error) {
	repoURI, err := reposource.CloneURLToRepoURI(cloneURL)
	if err != nil {
		return "", err
	}
	if repoURI == "" {
		return "", fmt.Errorf("No matching code host found for %s", cloneURL)
	}
	return string(repoURI), nil
}

func createFileInfo(path string, isDir bool) os.FileInfo {
	return fileInfo{path: path, isDir: isDir}
}

func (r *gitTreeEntryResolver) IsSingleChild(ctx context.Context, args *gitTreeEntryConnectionArgs) (bool, error) {
	if !r.IsDirectory() {
		return false, nil
	}
	entries, err := git.ReadDir(ctx, backend.CachedGitRepo(r.commit.repo.repo), api.CommitID(r.commit.oid), filepath.Dir(r.path), false)
	if err != nil {
		return false, err
	}
	return len(entries) == 1, nil
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
