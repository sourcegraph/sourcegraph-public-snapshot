package vcs

// A Merger is a repository that can perform actions related to
// merging.
type Merger interface {
	// MergeBase returns the merge base commit for the specified
	// commits (aka greatest common ancestor commit for hg).
	MergeBase(CommitID, CommitID) (CommitID, error)
}

// A CrossRepoMerger is a repository that can perform merge-related
// actions across separate repositories.
type CrossRepoMerger interface {
	// CrossRepoMergeBase returns the merge base commit for the
	// specified commits (aka greatest common ancestor commit for hg).
	//
	// The commit specified by `b` must exist in repoB but does not
	// need to exist in the repository that CrossRepoMergeBase is
	// called on. Likewise, the commit specified by `a` need not exist
	// in repoB.
	CrossRepoMergeBase(a CommitID, repoB Repository, b CommitID) (CommitID, error)
}
