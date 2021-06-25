package run

import (
	"math"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/schema"
)

// SearchInputs contains fields we set before kicking off search.
type SearchInputs struct {
	Plan           query.Plan            // the comprehensive query plan
	Query          query.Q               // the current basic query being evaluated, one part of query.Plan
	OriginalQuery  string                // the raw string of the original search query
	Pagination     *SearchPaginationInfo // pagination information, or nil if the request is not paginated.
	PatternType    query.SearchType
	VersionContext *string
	UserSettings   *schema.Settings

	// DefaultLimit is the default limit to use if not specified in query.
	DefaultLimit int
}

// MaxResults computes the limit for the query.
func (inputs SearchInputs) MaxResults() int {
	if inputs.Pagination != nil {
		// Paginated search requests always consume an entire result set for a
		// given repository, so we do not want any limit here. See
		// search_pagination.go for details on why this is necessary .
		return math.MaxInt32
	}

	if inputs.Query == nil {
		return 0
	}

	if count := inputs.Query.Count(); count != nil {
		return *count
	}

	if inputs.DefaultLimit != 0 {
		return inputs.DefaultLimit
	}

	return search.DefaultMaxSearchResults
}

// SearchPaginationInfo describes information around a paginated search
// request.
type SearchPaginationInfo struct {
	// cursor indicates where to resume searching from (see docstrings on
	// SearchCursor) or nil when requesting the first page of results.
	Cursor *SearchCursor

	// limit indicates at max how many search results to return.
	Limit int32
}

// SearchCursor represents a decoded search pagination cursor. From an API
// consumer standpoint, it is an encoded opaque string.
type SearchCursor struct {
	// RepositoryOffset indicates how many repositories (which are globally
	// sorted and ordered) to offset by.
	RepositoryOffset int32

	// ResultOffset indicates how many results within the first repository we
	// would search in to further offset by. This is so that we can paginate
	// results within e.g. a single large repository.
	ResultOffset int32

	// Finished tells if there are more results for the query or if we've
	// consumed them all.
	Finished bool
}

func statsDeref(s *streaming.Stats) streaming.Stats {
	if s == nil {
		return streaming.Stats{}
	}
	return *s
}
