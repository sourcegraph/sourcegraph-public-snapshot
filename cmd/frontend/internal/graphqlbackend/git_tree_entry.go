package graphqlbackend

import (
	"os"
	"path"
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
	url := prefix + "/-/"
	if r.IsDirectory() {
		url += "tree"
	} else {
		url += "blob"
	}
	return url + "/" + r.path
}
