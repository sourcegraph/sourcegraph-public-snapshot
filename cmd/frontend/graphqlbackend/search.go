package graphqlbackend

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"sync"

	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"

	searchrepos "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/search/repos"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/endpoint"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/search"
	searchbackend "github.com/sourcegraph/sourcegraph/internal/search/backend"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	querytypes "github.com/sourcegraph/sourcegraph/internal/search/query/types"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/vcs"
	"github.com/sourcegraph/sourcegraph/schema"
)

// This file contains the root resolver for search. It currently has a lot of
// logic that spans out into all the other search_* files.
var mockResolveRepositories func(effectiveRepoFieldValues []string) (resolved searchrepos.Resolved, err error)

type SearchArgs struct {
	Version        string
	PatternType    *string
	Query          string
	After          *string
	First          *int32
	VersionContext *string

	// For tests
	Settings *schema.Settings
}

type SearchImplementer interface {
	Results(context.Context) (*SearchResultsResolver, error)
	Suggestions(context.Context, *searchSuggestionsArgs) ([]*searchSuggestionResolver, error)
	//lint:ignore U1000 is used by graphql via reflection
	Stats(context.Context) (*searchResultsStats, error)

	SetStream(c SearchStream)
	Inputs() *SearchInputs
}

// NewSearchImplementer returns a SearchImplementer that provides search results and suggestions.
func NewSearchImplementer(ctx context.Context, args *SearchArgs) (_ SearchImplementer, err error) {
	tr, ctx := trace.New(ctx, "NewSearchImplementer", args.Query)
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	settings := args.Settings
	if settings == nil {
		var err error
		settings, err = decodedViewerFinalSettings(ctx)
		if err != nil {
			return nil, err
		}
	}

	searchType, err := detectSearchType(args.Version, args.PatternType)
	if err != nil {
		return nil, err
	}
	searchType = overrideSearchType(args.Query, searchType)

	if searchType == query.SearchTypeStructural && !conf.StructuralSearchEnabled() {
		return nil, errors.New("Structural search is disabled in the site configuration.")
	}

	var queryInfo query.QueryInfo
	globbing := getBoolPtr(settings.SearchGlobbing, false)
	tr.LogFields(otlog.Bool("globbing", globbing))
	queryInfo, err = query.ProcessAndOr(args.Query, query.ParserOptions{SearchType: searchType, Globbing: globbing})
	if err != nil {
		return alertForQuery(args.Query, err), nil
	}
	if getBoolPtr(settings.SearchUppercase, false) {
		q := queryInfo.(*query.AndOrQuery)
		q.Query = query.SearchUppercase(q.Query)
	}
	tr.LazyPrintf("parsing done")

	// If stable:truthy is specified, make the query return a stable result ordering.
	if queryInfo.BoolValue(query.FieldStable) {
		args, queryInfo, err = queryForStableResults(args, queryInfo)
		if err != nil {
			return alertForQuery(args.Query, err), nil
		}
	}

	// If the request is a paginated one, decode those arguments now.
	var pagination *searchPaginationInfo
	if args.First != nil {
		pagination, err = processPaginationRequest(args, queryInfo)
		if err != nil {
			return nil, err
		}
	}

	return &searchResolver{
		SearchInputs: &SearchInputs{
			Query:          queryInfo,
			OriginalQuery:  args.Query,
			VersionContext: args.VersionContext,
			UserSettings:   settings,
			Pagination:     pagination,
			PatternType:    searchType,
		},
		zoekt:        search.Indexed(),
		searcherURLs: search.SearcherURLs(),
		reposMu:      &sync.Mutex{},
		resolved:     &searchrepos.Resolved{},
	}, nil
}

func (r *schemaResolver) Search(ctx context.Context, args *SearchArgs) (SearchImplementer, error) {
	return NewSearchImplementer(ctx, args)
}

// queryForStableResults transforms a query that returns a stable result
// ordering. The transformed query uses pagination underneath the hood.
func queryForStableResults(args *SearchArgs, queryInfo query.QueryInfo) (*SearchArgs, query.QueryInfo, error) {
	if queryInfo.BoolValue(query.FieldStable) {
		var stableResultCount int32
		if _, countPresent := queryInfo.Fields()["count"]; countPresent {
			count, _ := queryInfo.StringValue(query.FieldCount)
			count64, err := strconv.ParseInt(count, 10, 32)
			if err != nil {
				return nil, nil, err
			}
			stableResultCount = int32(count64)
			if stableResultCount > maxSearchResultsPerPaginatedRequest {
				return nil, nil, fmt.Errorf("Stable searches are limited to at max count:%d results. Consider removing 'stable:', narrowing the search with 'repo:', or using the paginated search API.", maxSearchResultsPerPaginatedRequest)
			}
		} else {
			stableResultCount = defaultMaxSearchResults
		}
		args.First = &stableResultCount
		fileValue := "file"
		// Pagination only works for file content searches, and will
		// raise an error otherwise. If stable is explicitly set, this
		// is implied. So, force this query to only return file content
		// results.
		queryInfo.Fields()["type"] = []*querytypes.Value{{String: &fileValue}}
	}
	return args, queryInfo, nil
}

func processPaginationRequest(args *SearchArgs, queryInfo query.QueryInfo) (*searchPaginationInfo, error) {
	var pagination *searchPaginationInfo
	if args.First != nil {
		cursor, err := unmarshalSearchCursor(args.After)
		if err != nil {
			return nil, err
		}
		if *args.First < 0 || *args.First > maxSearchResultsPerPaginatedRequest {
			return nil, fmt.Errorf("search: requested pagination 'first' value outside allowed range (0 - %d)", maxSearchResultsPerPaginatedRequest)
		}
		pagination = &searchPaginationInfo{
			cursor: cursor,
			limit:  *args.First,
		}
	} else if args.After != nil {
		return nil, errors.New("search: paginated requests providing an 'after' cursor but no 'first' value is forbidden")
	}
	return pagination, nil
}

// detectSearchType returns the search type to perfrom ("regexp", or
// "literal"). The search type derives from three sources: the version and
// patternType parameters passed to the search endpoint (literal search is the
// default in V2), and the `patternType:` filter in the input query string which
// overrides the searchType, if present.
func detectSearchType(version string, patternType *string) (query.SearchType, error) {
	var searchType query.SearchType
	if patternType != nil {
		switch *patternType {
		case "literal":
			searchType = query.SearchTypeLiteral
		case "regexp":
			searchType = query.SearchTypeRegex
		case "structural":
			searchType = query.SearchTypeStructural
		default:
			return -1, fmt.Errorf("unrecognized patternType: %v", patternType)
		}
	} else {
		switch version {
		case "V1":
			searchType = query.SearchTypeRegex
		case "V2":
			searchType = query.SearchTypeLiteral
		default:
			return -1, fmt.Errorf("unrecognized version want \"V1\" or \"V2\": %v", version)
		}
	}
	return searchType, nil
}

var patternTypeRegex = lazyregexp.New(`(?i)patterntype:([a-zA-Z"']+)`)

func overrideSearchType(input string, searchType query.SearchType) query.SearchType {
	q, err := query.ParseAndOr(input, query.SearchTypeLiteral)
	q = query.LowercaseFieldNames(q)
	if err != nil {
		// If parsing fails, return the default search type. Any actual
		// parse errors will be raised by subsequent parser calls.
		return searchType
	}
	query.VisitField(q, "patterntype", func(value string, _ bool, _ query.Annotation) {
		switch value {
		case "regex", "regexp":
			searchType = query.SearchTypeRegex
		case "literal":
			searchType = query.SearchTypeLiteral
		case "structural":
			searchType = query.SearchTypeStructural
		}
	})
	return searchType
}

func getBoolPtr(b *bool, def bool) bool {
	if b == nil {
		return def
	}
	return *b
}

// SearchInputs contains fields we set before kicking off search.
type SearchInputs struct {
	Query          query.QueryInfo       // the query, either containing and/or expressions or otherwise ordinary
	OriginalQuery  string                // the raw string of the original search query
	Pagination     *searchPaginationInfo // pagination information, or nil if the request is not paginated.
	PatternType    query.SearchType
	VersionContext *string
	UserSettings   *schema.Settings
}

// searchResolver is a resolver for the GraphQL type `Search`
type searchResolver struct {
	*SearchInputs
	invalidateRepoCache bool // if true, invalidates the repo cache when evaluating search subexpressions.

	// resultChannel if non-nil will send all results we receive down it. See
	// searchResolver.SetResultChannel
	resultChannel SearchStream

	// Cached resolveRepositories results. We use a pointer to the mutex so that we
	// can copy the resolver, while sharing the mutex. If we didn't use a pointer,
	// the mutex would lead to unexpected behaviour.
	reposMu  *sync.Mutex
	resolved *searchrepos.Resolved
	repoErr  error

	zoekt        *searchbackend.Zoekt
	searcherURLs *endpoint.Map
}

// SearchEvent is an event on a search stream. It contains fields which can be
// aggregated up into a final result.
type SearchEvent struct {
	Results []SearchResultResolver
	Stats   streaming.Stats
	Error   error
}

// SearchStream is a send only channel of SearchEvent. All streaming search
// backends write to a SearchStream which is then streamed out by the HTTP
// layer.
type SearchStream chan<- SearchEvent

// collectStream is a helper for batch interfaces calling stream based
// functions. It returns a context, stream and cleanup/get function. The
// cleanup/get function will return the aggregated event and must be called
// once you have stopped sending to stream.
//
// For collecting errors we only collect the first error reported and
// afterwards cancel the context.
func collectStream(ctx context.Context) (context.Context, SearchStream, func() SearchEvent) {
	var agg SearchEvent

	ctx, cancel := context.WithCancel(ctx)

	done := make(chan struct{})
	stream := make(chan SearchEvent)
	go func() {
		defer close(done)
		for event := range stream {
			agg.Results = append(agg.Results, event.Results...)
			agg.Stats.Update(&event.Stats)
			// Only collect first error
			if event.Error != nil && agg.Error == nil {
				cancel()
				agg.Error = event.Error
			}
		}
	}()

	return ctx, stream, func() SearchEvent {
		cancel()
		close(stream)
		<-done
		return agg
	}
}

// SetStream will send all results down c.
//
// This is how our streaming and our batch interface co-exist. When this is
// set, it exposes a way to stream out results as we collect them.
//
// TODO(keegan) This is not our final design. For example this doesn't allow
// us to stream out things like dynamic filters or take into account
// AND/OR. However, streaming is behind a feature flag for now, so this is to
// make it visible in the browser.
func (r *searchResolver) SetStream(c SearchStream) {
	r.resultChannel = c
}

func (r *searchResolver) Inputs() *SearchInputs {
	return r.SearchInputs
}

// rawQuery returns the original query string input.
func (r *searchResolver) rawQuery() string {
	return r.OriginalQuery
}

func (r *searchResolver) countIsSet() bool {
	count, _ := r.Query.StringValues(query.FieldCount)
	max, _ := r.Query.StringValues(query.FieldMax)
	return len(count) > 0 || len(max) > 0
}

const defaultMaxSearchResults = 30
const maxSearchResultsPerPaginatedRequest = 5000

func (r *searchResolver) maxResults() int32 {
	if r.Pagination != nil {
		// Paginated search requests always consume an entire result set for a
		// given repository, so we do not want any limit here. See
		// search_pagination.go for details on why this is necessary .
		return math.MaxInt32
	}
	count, _ := r.Query.StringValues(query.FieldCount)
	if len(count) > 0 {
		n, _ := strconv.Atoi(count[0])
		if n > 0 {
			return int32(n)
		}
	}
	max, _ := r.Query.StringValues(query.FieldMax)
	if len(max) > 0 {
		n, _ := strconv.Atoi(max[0])
		if n > 0 {
			return int32(n)
		}
	}
	return defaultMaxSearchResults
}

var mockDecodedViewerFinalSettings *schema.Settings

func decodedViewerFinalSettings(ctx context.Context) (_ *schema.Settings, err error) {
	tr, ctx := trace.New(ctx, "decodedViewerFinalSettings", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()
	if mockDecodedViewerFinalSettings != nil {
		return mockDecodedViewerFinalSettings, nil
	}
	merged, err := viewerFinalSettings(ctx)
	if err != nil {
		return nil, err
	}
	var settings schema.Settings
	if err := json.Unmarshal([]byte(merged.Contents()), &settings); err != nil {
		return nil, err
	}
	return &settings, nil
}

// resolveRepositories calls ResolveRepositories, caching the result for the common case
// where effectiveRepoFieldValues == nil.
func (r *searchResolver) resolveRepositories(ctx context.Context, effectiveRepoFieldValues []string) (searchrepos.Resolved, error) {
	var err error
	var repoRevs, missingRepoRevs []*search.RepositoryRevisions
	var overLimit bool
	if mockResolveRepositories != nil {
		return mockResolveRepositories(effectiveRepoFieldValues)
	}

	tr, ctx := trace.New(ctx, "graphql.resolveRepositories", fmt.Sprintf("effectiveRepoFieldValues: %v", effectiveRepoFieldValues))
	defer func() {
		if err != nil {
			tr.SetError(err)
		} else {
			tr.LazyPrintf("numRepoRevs: %d, numMissingRepoRevs: %d, overLimit: %v", len(repoRevs), len(missingRepoRevs), overLimit)
		}
		tr.Finish()
	}()
	if effectiveRepoFieldValues == nil {
		r.reposMu.Lock()
		defer r.reposMu.Unlock()
		if r.resolved.RepoRevs != nil || r.resolved.MissingRepoRevs != nil || r.repoErr != nil {
			tr.LazyPrintf("cached")
			return *r.resolved, r.repoErr
		}
	}

	repoFilters, minusRepoFilters := r.Query.RegexpPatterns(query.FieldRepo)
	if effectiveRepoFieldValues != nil {
		repoFilters = effectiveRepoFieldValues
	}
	repoGroupFilters, _ := r.Query.StringValues(query.FieldRepoGroup)

	var settingForks, settingArchived bool
	if v := r.UserSettings.SearchIncludeForks; v != nil {
		settingForks = *v
	}
	if v := r.UserSettings.SearchIncludeArchived; v != nil {
		settingArchived = *v
	}

	forkStr, _ := r.Query.StringValue(query.FieldFork)
	fork := searchrepos.ParseYesNoOnly(forkStr)
	if fork == searchrepos.Invalid && !searchrepos.ExactlyOneRepo(repoFilters) && !settingForks {
		// fork defaults to No unless either of:
		// (1) exactly one repo is being searched, or
		// (2) user/org/global setting includes forks
		fork = searchrepos.No
	}

	archivedStr, _ := r.Query.StringValue(query.FieldArchived)
	archived := searchrepos.ParseYesNoOnly(archivedStr)
	if archived == searchrepos.Invalid && !searchrepos.ExactlyOneRepo(repoFilters) && !settingArchived {
		// archived defaults to No unless either of:
		// (1) exactly one repo is being searched, or
		// (2) user/org/global setting includes archives in all searches
		archived = searchrepos.No
	}

	visibilityStr, _ := r.Query.StringValue(query.FieldVisibility)
	visibility := query.ParseVisibility(visibilityStr)

	commitAfter, _ := r.Query.StringValue(query.FieldRepoHasCommitAfter)

	var versionContextName string
	if r.VersionContext != nil {
		versionContextName = *r.VersionContext
	}

	tr.LazyPrintf("resolveRepositories - start")
	options := searchrepos.Options{
		RepoFilters:        repoFilters,
		MinusRepoFilters:   minusRepoFilters,
		RepoGroupFilters:   repoGroupFilters,
		VersionContextName: versionContextName,
		UserSettings:       r.UserSettings,
		OnlyForks:          fork == searchrepos.Only,
		NoForks:            fork == searchrepos.No,
		OnlyArchived:       archived == searchrepos.Only,
		NoArchived:         archived == searchrepos.No,
		OnlyPrivate:        visibility == query.Private,
		OnlyPublic:         visibility == query.Public,
		CommitAfter:        commitAfter,
		Query:              r.Query,
	}
	resolved, err := searchrepos.ResolveRepositories(ctx, options)
	tr.LazyPrintf("resolveRepositories - done")
	if effectiveRepoFieldValues == nil {
		r.resolved = &resolved
		r.repoErr = err
	}
	return resolved, err
}

func (r *searchResolver) suggestFilePaths(ctx context.Context, limit int) ([]*searchSuggestionResolver, error) {
	resolved, err := r.resolveRepositories(ctx, nil)
	if err != nil {
		return nil, err
	}

	if resolved.OverLimit {
		// If we've exceeded the repo limit, then we may miss files from repos we care
		// about, so don't bother searching filenames at all.
		return nil, nil
	}

	p, err := r.getPatternInfo(&getPatternInfoOptions{forceFileSearch: true})
	if err != nil {
		return nil, err
	}

	args := search.TextParameters{
		PatternInfo:     p,
		RepoPromise:     (&search.Promise{}).Resolve(resolved.RepoRevs),
		Query:           r.Query,
		UseFullDeadline: r.searchTimeoutFieldSet(),
		Zoekt:           r.zoekt,
		SearcherURLs:    r.searcherURLs,
	}
	if err := args.PatternInfo.Validate(); err != nil {
		return nil, err
	}

	fileResults, _, err := searchFilesInReposBatch(ctx, &args)
	if err != nil {
		return nil, err
	}

	var suggestions []*searchSuggestionResolver
	for i, result := range fileResults {
		assumedScore := len(fileResults) - i // Greater score is first, so we inverse the index.
		suggestions = append(suggestions, newSearchSuggestionResolver(result.File(), assumedScore))
	}
	return suggestions, nil
}

type badRequestError struct {
	err error
}

func (e *badRequestError) BadRequest() bool {
	return true
}

func (e *badRequestError) Error() string {
	return "bad request: " + e.err.Error()
}

func (e *badRequestError) Cause() error {
	return e.err
}

// searchSuggestionResolver is a resolver for the GraphQL union type `SearchSuggestion`
type searchSuggestionResolver struct {
	// result is either a RepositoryResolver or a GitTreeEntryResolver
	result interface{}
	// score defines how well this item matches the query for sorting purposes
	score int
	// length holds the length of the item name as a second sorting criterium
	length int
	// label to sort alphabetically by when all else is equal.
	label string
}

func (r *searchSuggestionResolver) ToRepository() (*RepositoryResolver, bool) {
	res, ok := r.result.(*RepositoryResolver)
	return res, ok
}

func (r *searchSuggestionResolver) ToFile() (*GitTreeEntryResolver, bool) {
	res, ok := r.result.(*GitTreeEntryResolver)
	return res, ok
}

func (r *searchSuggestionResolver) ToGitBlob() (*GitTreeEntryResolver, bool) {
	res, ok := r.result.(*GitTreeEntryResolver)
	return res, ok && res.stat.Mode().IsRegular()
}

func (r *searchSuggestionResolver) ToGitTree() (*GitTreeEntryResolver, bool) {
	res, ok := r.result.(*GitTreeEntryResolver)
	return res, ok && res.stat.Mode().IsDir()
}

func (r *searchSuggestionResolver) ToSymbol() (*symbolResolver, bool) {
	s, ok := r.result.(*searchSymbolResult)
	if !ok {
		return nil, false
	}
	return toSymbolResolver(s.symbol, s.baseURI, s.lang, s.commit), true
}

func (r *searchSuggestionResolver) ToLanguage() (*languageResolver, bool) {
	res, ok := r.result.(*languageResolver)
	return res, ok
}

// newSearchSuggestionResolver returns a new searchSuggestionResolver wrapping the
// given result.
//
// A panic occurs if the type of result is not a *RepositoryResolver, *GitTreeEntryResolver,
// *searchSymbolResult or *languageResolver.
func newSearchSuggestionResolver(result interface{}, score int) *searchSuggestionResolver {
	switch r := result.(type) {
	case *RepositoryResolver:
		return &searchSuggestionResolver{result: r, score: score, length: len(r.innerRepo.Name), label: r.Name()}

	case *GitTreeEntryResolver:
		return &searchSuggestionResolver{result: r, score: score, length: len(r.Path()), label: r.Path()}

	case *searchSymbolResult:
		return &searchSuggestionResolver{result: r, score: score, length: len(r.symbol.Name + " " + r.symbol.Parent), label: r.symbol.Name + " " + r.symbol.Parent}

	case *languageResolver:
		return &searchSuggestionResolver{result: r, score: score, length: len(r.Name()), label: r.Name()}

	default:
		panic("never here")
	}
}

func sortSearchSuggestions(s []*searchSuggestionResolver) {
	sort.Slice(s, func(i, j int) bool {
		// Sort by score
		a, b := s[i], s[j]
		if a.score != b.score {
			return a.score > b.score
		}
		// Prefer shorter strings for the same match score
		// E.g. prefer gorilla/mux over gorilla/muxy, Microsoft/vscode over g3ortega/vscode-crystal
		if a.length != b.length {
			return a.length < b.length
		}

		// All else equal, sort alphabetically.
		return a.label < b.label
	})
}

// handleRepoSearchResult handles the limitHit and searchErr returned by a search function,
// returning common as to reflect that new information. If searchErr is a fatal error,
// it returns a non-nil error; otherwise, if searchErr == nil or a non-fatal error, it returns a
// nil error.
func handleRepoSearchResult(repoRev *search.RepositoryRevisions, limitHit, timedOut bool, searchErr error) (_ streaming.Stats, fatalErr error) {
	var status search.RepoStatus
	if limitHit {
		status |= search.RepoStatusLimitHit
	}

	if vcs.IsRepoNotExist(searchErr) {
		if vcs.IsCloneInProgress(searchErr) {
			status |= search.RepoStatusCloning
		} else {
			status |= search.RepoStatusMissing
		}
	} else if gitserver.IsRevisionNotFound(searchErr) {
		if len(repoRev.Revs) == 0 || len(repoRev.Revs) == 1 && repoRev.Revs[0].RevSpec == "" {
			// If we didn't specify an input revision, then the repo is empty and can be ignored.
		} else {
			fatalErr = searchErr
		}
	} else if errcode.IsNotFound(searchErr) {
		status |= search.RepoStatusMissing
	} else if errcode.IsTimeout(searchErr) || errcode.IsTemporary(searchErr) || timedOut {
		status |= search.RepoStatusTimedout
	} else if searchErr != nil {
		fatalErr = searchErr
	} else {
		status |= search.RepoStatusSearched
	}
	return streaming.Stats{
		Status:     search.RepoStatusSingleton(repoRev.Repo.ID, status),
		IsLimitHit: limitHit,
	}, fatalErr
}

// getRepos is a wrapper around p.Get. It returns an error if the promise
// contains an underlying type other than []*search.RepositoryRevisions.
func getRepos(ctx context.Context, p *search.Promise) ([]*search.RepositoryRevisions, error) {
	v, err := p.Get(ctx)
	if err != nil {
		return nil, err
	}
	repoRevs, ok := v.([]*search.RepositoryRevisions)
	if !ok {
		return nil, fmt.Errorf("unexpected underlying type (%T) of promise", v)
	}
	return repoRevs, nil
}
