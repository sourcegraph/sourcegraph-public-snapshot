package sharedresolvers

import (
	"io/fs"
)

// GitTreeEntryResolver resolves an entry in a Git tree in a repository. The entry can be any Git
// object type that is valid in a tree.
//
// Prefer using the constructor, NewGitTreeEntryResolver.
type GitTreeEntryResolver struct {
	commit *GitCommitResolver

	// stat is this tree entry's file info. Its Name method must return the full path relative to
	// the root, not the basename.
	stat fs.FileInfo
}

func NewGitTreeEntryResolver(commit *GitCommitResolver, stat fs.FileInfo) *GitTreeEntryResolver {
	return &GitTreeEntryResolver{commit: commit, stat: stat}
}

func (r *GitTreeEntryResolver) Path() string                    { return r.stat.Name() }
func (r *GitTreeEntryResolver) Commit() *GitCommitResolver      { return r.commit }
func (r *GitTreeEntryResolver) Repository() *RepositoryResolver { return r.commit.repoResolver }
