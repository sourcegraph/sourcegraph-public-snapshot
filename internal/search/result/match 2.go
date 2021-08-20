package result

import (
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/search/filter"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// Match is *FileMatch | *RepoMatch | *CommitMatch. We have a private method
// to ensure only those types implement Match.
type Match interface {
	ResultCount() int
	Limit(int) int
	Select(filter.SelectPath) Match
	RepoName() types.RepoName

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

	// Rev is the revision associated with the repo if it exists
	Rev string

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

	if k.Rev != other.Rev {
		return k.Rev < other.Rev
	}

	if k.Commit != other.Commit {
		return k.Commit < other.Commit
	}

	if k.Path != other.Path {
		return k.Path < other.Path
	}

	return k.TypeRank < other.TypeRank
}

// Matches implements sort.Interface
type Matches []Match

func (m Matches) Len() int           { return len(m) }
func (m Matches) Less(i, j int) bool { return m[i].Key().Less(m[j].Key()) }
func (m Matches) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
