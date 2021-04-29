package result

import "github.com/sourcegraph/sourcegraph/internal/search/filter"

// Match is *FileMatch | *RepoMatch | *CommitMatch. We have a private method
// to ensure only those types implement Match.
type Match interface {
	ResultCount() int
	Limit(int) int
	Select(filter.SelectPath) Match
	Key() Key

	// ensure only types in this package can be a Match.
	searchResultMarker()
}

var _ Match = (*FileMatch)(nil)
var _ Match = (*RepoMatch)(nil)
var _ Match = (*CommitMatch)(nil)

type MatchSlice []Match

func (m MatchSlice) Len() int           { return len(m) }
func (m MatchSlice) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m MatchSlice) Less(i, j int) bool { return m[i].Key().Less(m[j].Key()) }

type Key struct {
	TypeRank int
	Repo     string
	Commit   string
	Path     string
}

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
