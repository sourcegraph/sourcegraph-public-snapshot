package graphqlbackend

import (
	"context"
	"fmt"
	"time"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
)

// searchCursor represents a decoded search pagination cursor. From an API
// consumer standpoint, it is an encoded opaque string.
type searchCursor struct {
	// RepositoryOffset indicates how many repositories (which are globally
	// sorted and ordered) to offset by.
	RepositoryOffset int32

	// ResultOffset indicates how many results within the first repository we
	// would search in to further offset by. This is so that we can paginate
	// results within e.g. a single large repository.
	ResultOffset int32

	// UserID is the ID of the user that created this cursor. This is useful
	// for two reasons:
	//
	// 1. It tells us if the user making the request is different than the user
	//    that created the cursor, in which case their result set may differ
	//    because they have access to a different set of repositories (and we
	//    could e.g. warn them of this or handle it more fancily).
	//
	// 2. When we pre-emptively fetch more results for this search query so
	//    that our next response to user A is super fast, we _cannot_ give
	//    the pre-emptively fetched results for user A to a different user B
	//    making a request for the same cursor. This is because user A and
	//    user B may have access to a different set of repositories.
	//
	// Note that when a user is providing a cursor, they can forge any field
	// they like and as such this user ID cannot be trusted in that case.
	UserID int32
}

const searchCursorKind = "SearchCursor"

// marshalSearchCursor marshals a search pagination cursor.
func marshalSearchCursor(c *searchCursor) graphql.ID {
	return relay.MarshalID(searchCursorKind, c)
}

// unmarshalSearchCursor unmarshals a search pagination cursor.
func unmarshalSearchCursor(cursor *graphql.ID) (*searchCursor, error) {
	if cursor == nil {
		return nil, nil
	}
	if kind := relay.UnmarshalKind(*cursor); kind != searchCursorKind {
		return nil, fmt.Errorf("cannot unmarshal search cursor type: %q", kind)
	}
	var spec *searchCursor
	if err := relay.UnmarshalSpec(*cursor, &spec); err != nil {
		return nil, err
	}
	return spec, nil
}

// searchPaginationInfo describes information around a paginated search
// request.
type searchPaginationInfo struct {
	// cursor indicates where to resume searching from (see docstrings on
	// searchCursor) or nil when requesting the first page of results.
	cursor *searchCursor

	// limit indicates at max how many search results to return.
	limit int32
}

// Finished tells whether or not pagination has consumed all results that are
// available.
func (r *searchResolver) Finished(ctx context.Context) bool {
	if r.pagination == nil {
		return false // Will always be false for non-paginated requests.
	}
	panic("TODO: implement")
}

// Cursor returns the cursor that can be passed into a future search request in
// order to fetch more results starting where this search left off.
func (r *searchResolver) Cursor(ctx context.Context) graphql.ID {
	if r.pagination == nil {
		return "" // Only valid when the original request was a paginated one.
	}
	panic("TODO: implement")
}

// paginatedResults handles serving paginated search queries. It's logic does
// not live alongside the non-paginated doResults because:
//
// 1. It would introduce many `if r.pagination != nil` conditionals which would
//    make that code harder to reason about.
// 2. That method is already very large and brittle, common logic can be
//    refactored out instead.
// 3. The way that method operates (mixing in search result types depending on
//    a timeout, searcing result types in parallel) is fundamentally incompatible
//    with the absolute ordering we do here for pagination.
//
func (r *searchResolver) paginatedResults(ctx context.Context) (*searchResultsResolver, error) {
	if r.pagination == nil {
		panic("(bug) this method should never be called in this state")
	}

	// All paginated search requests should complete within this timeframe.
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	panic("TODO: implement")
}
