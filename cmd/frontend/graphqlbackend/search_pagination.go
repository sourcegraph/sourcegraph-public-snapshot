package graphqlbackend

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search/query"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/trace"
	log15 "gopkg.in/inconshreveable/log15.v2"
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
	panic("TODO(slimsag): before merge: implement")
}

// Cursor returns the cursor that can be passed into a future search request in
// order to fetch more results starting where this search left off.
func (r *searchResolver) Cursor(ctx context.Context) graphql.ID {
	if r.pagination == nil {
		return "" // Only valid when the original request was a paginated one.
	}
	panic("TODO(slimsag): before merge: implement")
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

	// TODO(slimsag): future: support non-file result types in the paginated
	// search API.
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
	results, fileCommon, err := paginatedSearchFilesInRepos(ctx, &args, r.pagination)
	// Timeouts are reported through searchResultsCommon so don't report an error for them
	if err != nil && !(err == context.DeadlineExceeded || err == context.Canceled) {
		return nil, err
	}
	common.update(*fileCommon)

	tr.LazyPrintf("results=%d limitHit=%v cloning=%d missing=%d timedout=%d", len(results), common.limitHit, len(common.cloning), len(common.missing), len(common.timedout))

	// Alert is a potential alert shown to the user.
	var alert *searchAlert

	if len(missingRepoRevs) > 0 {
		alert = r.alertForMissingRepoRevs(missingRepoRevs)
	}

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

// paginatedSearchFilesInRepos implements result-level pagination by calling
// searchFilesInRepos to search over subsets (batches) of the total list of
// repositories that may have results for this request (args.Repos). It does
// this by picking some tradeoffs to balance some conflicting facts:
//
// 1. Paginated text searches must currently ask Zoekt AND non-indexed search
//    to produce the entire result set for a repository. This is like querying
//    for `repo:^exact-repo$ count:1000000` in a non-paginated query, and is
//    more costly and slower than the default `count:30` used in non-paginated
//    requests (search for FileMatchLimit) which allows Zoekt/non-indexed
//    search to stop searching after finding enough results. Another reason for
//    needing to produce the entire result set for a repository is because
//    Zoekt does not today produce a stable order of results.
//
// 2. With NITH (needle-in-the-haystack) queries, if we don't search enough
//    repositories in parallel we would substantially harm the performance of
//    these queries. For example, if we were to search 100 repositories at a
//    time and there were 1000 repositories to search and only the last 100
//    repositories had results for you, you need to wait for the first 9
//    batched searches to complete making your results 10x slower to fetch on
//    top of the penalty we incur from the larger `count:` mentioned in point
//    2 above (in the worst case scenario).
//
func paginatedSearchFilesInRepos(ctx context.Context, args *search.Args, pagination *searchPaginationInfo) ([]searchResultResolver, *searchResultsCommon, error) {
	plan := &repoPaginationPlan{
		pagination:   pagination,
		repositories: args.Repos,
		// TODO(slimsag): future: Potentially update and reason about these choices
		// more. They are mostly arbitrary currently and should instead be
		// based on measurements from benchmarking + measuring NITH query
		// performance, etc.
		searchBucketDivisor: 8,
		searchBucketMin:     10,
		searchBucketMax:     1000,
	}
	return plan.execute(ctx, func(batch []*search.RepositoryRevisions) ([]searchResultResolver, *searchResultsCommon, error) {
		batchArgs := *args
		batchArgs.Repos = batch
		fileResults, fileCommon, err := searchFilesInRepos(ctx, &batchArgs)
		if err != nil {
			return nil, nil, err
		}
		results := make([]searchResultResolver, 0, len(fileResults))
		for _, r := range fileResults {
			results = append(results, r)
		}
		// TODO(slimsag): future: searchFilesInRepos _does_ appear to guarantee
		// stable result ordering in our case (with large count:) but we
		// definitely need a test ensuring this.
		return results, fileCommon, nil
	})
}

// repoPaginationPlan describes a plan for executing a search function that
// searches only over a set of repositories (i.e. the search function offers no
// pagination or result-level pagination capabilities) to provide result-level
// pagination. That is, if you have a function which can provide a complete
// list of results for a given repository, this planner can be used to
// implement result-level pagination on top of that function.
//
// It does this by searching over a globally-sorted list of repositories in
// batches.
type repoPaginationPlan struct {
	// pagination is the pagination request we're trying to fulfill.
	pagination *searchPaginationInfo

	// repositories is the exhaustive and complete list of sorted repositories
	// to be searched over multiple requests.
	repositories []*search.RepositoryRevisions

	// parameters for controlling the size of batches that the executor is
	// called to search. The final batch size is calculated as:
	//
	// 	batchSize = numTotalReposOnSourcegraph() / searchBucketDivisor
	//
	// With the additional constraint that it must be at least min and no
	// larger than max.
	searchBucketDivisor              int
	searchBucketMin, searchBucketMax int
}

// executor is a function which searches a batch of repositories.
type executor func(batch []*search.RepositoryRevisions) ([]searchResultResolver, *searchResultsCommon, error)

// execute executes the repository pagination plan by invoking the executor to
// search batches of repositories.
//
// If the executor returns any error, the search will be cancelled and the error
// returned.
func (p *repoPaginationPlan) execute(ctx context.Context, exec executor) (results []searchResultResolver, common *searchResultsCommon, err error) {
	// Determine how large the batches of repositories we will search over will be.
	batchSize := clamp(numTotalRepos.get(ctx)/p.searchBucketDivisor, p.searchBucketMin, p.searchBucketMax)

	// Determine where in the repositories list we will begin searching.
	repos := p.repositories
	if p.pagination.cursor != nil {
		// Clamping is required here because the repositories the user has
		// access to could have changed if e.g. permissions for that user
		// were updated OR if this cursor was generated by a user with
		// different permissions.
		repoOffset := clamp(int(p.pagination.cursor.RepositoryOffset), 0, len(repos)-1)
		repos = repos[repoOffset:]
	}

	// Search over the repos list in batches.
	//
	// TODO(slimsag): future: scrutinize this code for off-by-one errors (I wrote this while sleepy.)
	common = &searchResultsCommon{}
	for start := 0; start <= len(repos); start += batchSize {
		if start > len(repos) {
			break
		}

		batch := repos[start:clamp(start+batchSize, 0, len(repos)-1)]
		batchResults, batchCommon, err := exec(batch)
		if err != nil {
			return nil, nil, err
		}
		// TODO(slimsag): future: Unlike in non-paginated search, if we see:
		//
		// 	len(batchCommon.cloning) > 0 || len(batchCommon.missing) > 0 || len(batchCommon.timedout) > 0 || len(batchCommon.partial) > 0
		//
		// We most likely want to cancel the request and return an error?
		// Otherwise, this breaks ordering guarantees perhaps? Needs more
		// thought.

		// Accumulate the results and stop if we have enough for the user.
		results = append(results, batchResults...)
		common.update(*batchCommon)
		if len(results) >= int(p.pagination.limit) {
			break
		}
	}
	// If we found more results than the user wanted, discard the remaining
	// ones.
	//
	// TODO(slimsag): future: cache these results somewhere because the user is
	// likely to come back soon for more and we don't need to search a batch of
	// repositories every time. This would give substantial performance
	// benefits to subsequent requests against this API.
	offset := 0
	if cursor := p.pagination.cursor; cursor != nil {
		offset = int(cursor.ResultOffset)
	}
	results, common = sliceSearchResults(results, common, offset, int(p.pagination.limit))
	return results, common, nil
}

// sliceSearchResults returns results[offset:offset+limit] and a searchResultsCommon
// structure reflecting that slice of results.
func sliceSearchResults(results []searchResultResolver, common *searchResultsCommon, offset, limit int) ([]searchResultResolver, *searchResultsCommon) {
	if offset == 0 && len(results) < offset+limit {
		return results, common
	}

	// Break results into repositories because for each result we need to add
	// the respective repository to the new common structure.
	reposByName := map[string]*types.Repo{}
	for _, r := range common.repos {
		reposByName[string(r.Name)] = r
	}
	resultsByRepo := map[*types.Repo][]searchResultResolver{}
	for _, r := range results[offset : offset+limit] {
		repoName, _ := r.searchResultURIs()
		repo := reposByName[repoName]
		resultsByRepo[repo] = append(resultsByRepo[repo], r)
	}

	// Construct the new searchResultsCommon structure for just the results
	// we're returning.
	finalResults := make([]searchResultResolver, 0, limit)
	finalCommon := &searchResultsCommon{
		// TODO(slimsag): before merge: document limitHit in schema.graphql
		limitHit:         false, // irrelevant in paginated search
		indexUnavailable: common.indexUnavailable,
		partial:          make(map[api.RepoName]struct{}),
	}
	copy := func(repo *types.Repo, targetList *[]*types.Repo, ifInsideList []*types.Repo) {
		for _, r := range ifInsideList {
			if repo == r {
				*targetList = append(*targetList, repo)
				return
			}
		}
	}
	for repo, results := range resultsByRepo {
		// TODO(slimsag): future: approximateResultCount in paginated requests
		// is certainly wrong because it doesn't account for prior repos in
		// prior paginated requests with the cursor.

		// Include the results and copy over metadata from the common structure.
		finalResults = append(finalResults, results...)
		finalCommon.resultCount += int32(len(results))
		copy(repo, &finalCommon.repos, common.repos)
		copy(repo, &finalCommon.searched, common.searched)
		copy(repo, &finalCommon.indexed, common.indexed)
		copy(repo, &finalCommon.cloning, common.cloning)
		copy(repo, &finalCommon.missing, common.missing)
		copy(repo, &finalCommon.timedout, common.timedout)
		if _, ok := common.partial[repo.Name]; ok {
			finalCommon.partial[repo.Name] = struct{}{}
		}
	}
	return finalResults, finalCommon
}

// clamp clamps x into the range of [min, max].
func clamp(x, min, max int) int {
	if x < min {
		return min
	}
	if x > max {
		return max
	}
	return x
}

// Since we will need to know the number of total repos on Sourcegraph for
// every paginated search request, but the exact number doesn't matter, we
// cache the result for a minute to avoid executing many DB count operations.
type numTotalReposCache struct {
	sync.RWMutex
	lastUpdate time.Time
	count      int
}

func (n *numTotalReposCache) get(ctx context.Context) int {
	n.RLock()
	if !n.lastUpdate.IsZero() && time.Since(n.lastUpdate) < 1*time.Minute {
		defer n.RUnlock()
		return n.count
	}
	n.RUnlock()

	n.Lock()
	newCount, err := db.Repos.Count(ctx, db.ReposListOptions{Enabled: true})
	if err != nil {
		defer n.Unlock()
		log15.Error("failed to determine numTotalRepos", "error", err)
		return n.count
	}
	n.count = newCount
	n.Unlock()
	return newCount
}

var numTotalRepos = &numTotalReposCache{}
