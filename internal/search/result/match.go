package result

import (
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/search/filter"
)

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

// Guard to ensure all match types implement the interface
var (
	_ Match = (*FileMatch)(nil)
	_ Match = (*RepoMatch)(nil)
	_ Match = (*CommitMatch)(nil)
)

// Match ranks are used for sorting the different match types.
// Match types with lower ranks will be sorted before match types
// with higher ranks.
const (
	rankFileMatch   = 0
	rankCommitMatch = 1
	rankDiffMatch   = 2
	rankRepoMatch   = 3
)

// Key is a sorting or deduplicating key for a Match.
// It contains all the identifying information for the Match.
type Key struct {
	// Repo is the name of the repo the match belongs to
	Repo api.RepoName

	// Commit is the commit hash of the commit the match belongs to.
	// Empty if there is no commit associated with the match (e.g. RepoMatch)
	Commit api.CommitID

	// Path is the path of the file the match belongs to.
	// Empty if there is no file associated with the match (e.g. RepoMatch or CommitMatch)
	Path string

	// TypeRank is the sorting rank of the type this key belongs to.
	TypeRank int
}

// Less compares one key to another for sorting
func (k Key) Less(other Key) bool {
	if k.Repo != other.Repo {
		return k.Repo < other.Repo
	}

	if k.Commit != other.Commit {
		return k.Commit < other.Commit
	}

	if k.Path != other.Path {
		return k.Path < other.Path
	}

	return k.TypeRank < other.TypeRank
}
