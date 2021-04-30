package result

import "github.com/sourcegraph/sourcegraph/internal/search/filter"

// Match is *FileMatch | *RepoMatch | *CommitMatch. We have a private method
// to ensure only those types implement Match.
type Match interface {
	ResultCount() int
	Limit(int) int
	Select(filter.SelectPath) Match

	// Key returns a key which uniquely identifies this match.
	Key() Key

	// ensure only types in this package can be a Match.
	searchResultMarker()
}

var _ Match = (*FileMatch)(nil)
var _ Match = (*RepoMatch)(nil)
var _ Match = (*CommitMatch)(nil)

// Key is a sorting or deduplicating key for a Match.
// It contains all the identifying information for the Match.
type Key struct {
	// TypeRank is the sorting rank of the type this key belongs to.
	// FileMatch = 0
	// CommitMatch (Commit) = 1
	// CommitMatch (Diff) = 2
	// RepoMatch = 3
	TypeRank int

	// Repo is the name of the repo the match belongs to
	Repo string

	// Commit is the commit hash of the commit the match belongs to.
	// Empty if there is no commit associated with the match (e.g. RepoMatch)
	Commit string

	// Path is the path of the file the match belongs to.
	// Empty if there is no file associated with the match (e.g. RepoMatch or CommitMatch)
	Path string
}

// Less compares one key to another for sorting
func (k Key) Less(other Key) bool {
	if k.TypeRank != other.TypeRank {
		return k.TypeRank < other.TypeRank
	}

	if k.Repo != other.Repo {
		return k.Repo < other.Repo
	}

	if k.Commit != other.Commit {
		return k.Commit < other.Commit
	}

	return k.Path < other.Path
}
