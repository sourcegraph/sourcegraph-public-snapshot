package result

import "github.com/sourcegraph/sourcegraph/internal/search/filter"

// Match is *FileMatch | *RepoMatch | *CommitMatch. We have a private method
// to ensure only those types implement Match.
type Match interface {
	ResultCount() int
	Limit(int) int
	Select(filter.SelectPath) Match

	// ensure only types in this package can be a Match.
	searchResultMarker()
}

var _ Match = (*FileMatch)(nil)
var _ Match = (*RepoMatch)(nil)
var _ Match = (*CommitMatch)(nil)
