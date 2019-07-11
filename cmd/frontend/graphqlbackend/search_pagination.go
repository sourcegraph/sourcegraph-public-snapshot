package graphqlbackend

import (
	"context"
	"fmt"
	"sort"
	"time"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search/query"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/trace"
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
func (r *searchResolver) paginatedResults(ctx context.Context) (result *searchResultsResolver, err error) {
	start := time.Now()
	if r.pagination == nil {
		panic("(bug) this method should never be called in this state")
	}

	tr, ctx := trace.New(ctx, "graphql.SearchResults.paginatedResults", r.rawQuery())
	if r.pagination.cursor != nil {
		tr.LogFields(
			otlog.Int("Cursor.RepositoryOffset", int(r.pagination.cursor.RepositoryOffset)),
			otlog.Int("Cursor.ResultOffset", int(r.pagination.cursor.ResultOffset)),
			otlog.Int("Cursor.UserID", int(r.pagination.cursor.UserID)),
		)
	} else {
		tr.LogFields(otlog.String("Cursor", "nil"))
	}
	tr.LogFields(otlog.Int("Limit", int(r.pagination.limit)))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	// All paginated search requests should complete within this timeframe.
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	////////////////////////////////////////////////////////////////////////////
	// TODO(slimsag): before merge: duplicate code from doResults begins here //
	////////////////////////////////////////////////////////////////////////////
	forceOnlyResultType := ""
	repos, missingRepoRevs, overLimit, err := r.resolveRepositories(ctx, nil)
	if err != nil {
		return nil, err
	}
	tr.LazyPrintf("searching %d repos, %d missing", len(repos), len(missingRepoRevs))
	if len(repos) == 0 {
		alert, err := r.alertForNoResolvedRepos(ctx)
		if err != nil {
			return nil, err
		}
		return &searchResultsResolver{alert: alert, start: start}, nil
	}
	if overLimit {
		alert, err := r.alertForOverRepoLimit(ctx)
		if err != nil {
			return nil, err
		}
		return &searchResultsResolver{alert: alert, start: start}, nil
	}

	p, err := r.getPatternInfo(nil)
	if err != nil {
		return nil, err
	}
	args := search.Args{
		Pattern: p,
		Repos:   repos,
		Query:   r.query,
		//UseFullDeadline: r.searchTimeoutFieldSet(),
		UseFullDeadline: true,
		Zoekt:           r.zoekt,
	}
	if err := args.Pattern.Validate(); err != nil {
		return nil, &badRequestError{err}
	}

	err = validateRepoHasFileUsage(r.query)
	if err != nil {
		return nil, err
	}

	// Determine which types of results to return.
	var resultTypes []string
	if forceOnlyResultType != "" {
		resultTypes = []string{forceOnlyResultType}
	} else {
		resultTypes, _ = r.query.StringValues(query.FieldType)
		if len(resultTypes) == 0 {
			//resultTypes = []string{"file", "path", "repo", "ref"}
			resultTypes = []string{"file"}
		}
	}
	for _, resultType := range resultTypes {
		if resultType == "file" {
			args.Pattern.PatternMatchesContent = true
		} else if resultType == "path" {
			args.Pattern.PatternMatchesPath = true
		}
	}
	tr.LazyPrintf("resultTypes: %v", resultTypes)
	//////////////////////////////////////////////////////////////////////////
	// TODO(slimsag): before merge: duplicate code from doResults ends here //
	//////////////////////////////////////////////////////////////////////////

	if len(resultTypes) != 1 || resultTypes[0] != "file" {
		return nil, fmt.Errorf("experimental paginated search currently only supports 'file' (text match) result types. Found %q", resultTypes)
	}

	// Since we're searching a subset of the repositories this query would
	// search overall, we must sort the repositories deterministically.
	for _, repoRev := range repos {
		sort.Slice(repoRev.Revs, func(i, j int) bool {
			return repoRev.Revs[i].Less(repoRev.Revs[j])
		})
	}
	sort.Slice(repos, func(i, j int) bool {
		return repoIsLess(repos[i].Repo, repos[j].Repo)
	})

	common := searchResultsCommon{maxResultsCount: r.maxResults()}
	fileResults, fileCommon, err := searchFilesInRepos(ctx, &args)
	// Timeouts are reported through searchResultsCommon so don't report an error for them
	if err != nil && !(err == context.DeadlineExceeded || err == context.Canceled) {
		return nil, err
	}
	common.update(*fileCommon)
	results := make([]searchResultResolver, 0, len(fileResults))
	for _, fr := range fileResults {
		results = append(results, fr)
	}

	tr.LazyPrintf("results=%d limitHit=%v cloning=%d missing=%d timedout=%d", len(results), common.limitHit, len(common.cloning), len(common.missing), len(common.timedout))

	// Alert is a potential alert shown to the user.
	var alert *searchAlert

	if len(missingRepoRevs) > 0 {
		alert = r.alertForMissingRepoRevs(missingRepoRevs)
	}

	sortResults(results)

	return &searchResultsResolver{
		start:               start,
		searchResultsCommon: common,
		results:             results,
		alert:               alert,
	}, nil
}

// repoIsLess sorts repositories first by name then by ID, suitable for use
// with sort.Slice.
func repoIsLess(i, j *types.Repo) bool {
	if i.Name != j.Name {
		return i.Name < j.Name
	}
	return i.ID < j.ID
}
