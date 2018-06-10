package graphqlbackend

import (
	"context"
	"os"
	"path"
)

// gitTreeEntryResolver resolves an entry in a Git tree in a repository. The entry can be any Git
// object type that is valid in a tree.
type gitTreeEntryResolver struct {
	commit *gitCommitResolver

	path string // this tree entry's path (relative to the root)

	// stat is populated by the creator of this gitTreeEntryResolver if it has this
	// information available. Not all creators will have the stat info; in
	// that case, some gitTreeEntryResolver methods have to look up the information
	// on their own.
	stat os.FileInfo

	isRecursive bool // whether entries is populated recursively (otherwise just current level of hierarchy)
}

func (r *gitTreeEntryResolver) Path() string { return r.path }
func (r *gitTreeEntryResolver) Name() string { return path.Base(r.path) }

func (r *gitTreeEntryResolver) ToGitTree() (*gitTreeEntryResolver, bool) { return r, true }
func (r *gitTreeEntryResolver) ToGitBlob() (*gitTreeEntryResolver, bool) { return r, true }

func (r *gitTreeEntryResolver) Commit() *gitCommitResolver { return r.commit }

func (r *gitTreeEntryResolver) Repository(ctx context.Context) (*repositoryResolver, error) {
	return r.commit.Repository(ctx)
}

func (r *gitTreeEntryResolver) IsRecursive() bool { return r.isRecursive }

func (r *gitTreeEntryResolver) URL(ctx context.Context) (string, error) {
	url := r.commit.repoRevURL() + "/-/"

	isDir, err := r.IsDirectory(ctx)
	if err != nil {
		return "", err
	}
	if isDir {
		url += "tree"
	} else {
		url += "blob"
	}
	return url + "/" + r.path, nil
}
