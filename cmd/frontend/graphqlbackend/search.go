package graphqlbackend

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	regexpsyntax "regexp/syntax"
	"sort"
	"strconv"
	"strings"
	"sync"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/neelance/parallel"
	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/endpoint"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/search"
	searchbackend "github.com/sourcegraph/sourcegraph/internal/search/backend"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/query/syntax"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/vcs"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
	"github.com/sourcegraph/sourcegraph/schema"
)

// This file contains the root resolver for search. It currently has a lot of
// logic that spans out into all the other search_* files.

func maxReposToSearch() int {
	switch max := conf.Get().MaxReposToSearch; {
	case max <= 0:
		// Default to a very large number that will not overflow if incremented.
		return math.MaxInt32 >> 1
	default:
		return max
	}
}

type SearchArgs struct {
	Version     string
	PatternType *string
	Query       string
	After       *string
	First       *int32
}

type SearchImplementer interface {
	Results(context.Context) (*SearchResultsResolver, error)
	Suggestions(context.Context, *searchSuggestionsArgs) ([]*searchSuggestionResolver, error)
	//lint:ignore U1000 is used by graphql via reflection
	Stats(context.Context) (*searchResultsStats, error)
}

// NewSearchImplementer returns a SearchImplementer that provides search results and suggestions.
func NewSearchImplementer(args *SearchArgs) (SearchImplementer, error) {
	tr, _ := trace.New(context.Background(), "graphql.schemaResolver", "Search")
	defer tr.Finish()

	searchType, err := detectSearchType(args.Version, args.PatternType, args.Query)
	if err != nil {
		return nil, err
	}

	if searchType == query.SearchTypeStructural && !conf.StructuralSearchEnabled() {
		return nil, errors.New("Structural search is disabled in the site configuration.")
	}

	var queryString string
	if searchType == query.SearchTypeLiteral {
		queryString = query.ConvertToLiteral(args.Query)
	} else {
		queryString = args.Query
	}

	q, p, err := query.Process(queryString, searchType)
	if err != nil {
		return alertForQuery(queryString, err), nil
	}

	// If the request is a paginated one, decode those arguments now.
	var pagination *searchPaginationInfo
	if args.First != nil {
		cursor, err := unmarshalSearchCursor(args.After)
		if err != nil {
			return nil, err
		}
		if *args.First < 0 || *args.First > 5000 {
			return nil, errors.New("search: requested pagination 'first' value outside allowed range (0 - 5000)")
		}
		pagination = &searchPaginationInfo{
			cursor: cursor,
			limit:  *args.First,
		}
	} else if args.After != nil {
		return nil, errors.New("Search: paginated requests providing a 'after' but no 'first' is forbidden")
	}

	return &searchResolver{
		query:         q,
		parseTree:     p,
		originalQuery: args.Query,
		pagination:    pagination,
		patternType:   searchType,
		zoekt:         search.Indexed(),
		searcherURLs:  search.SearcherURLs(),
	}, nil
}

func (r *schemaResolver) Search(args *SearchArgs) (SearchImplementer, error) {
	return NewSearchImplementer(args)
}

// detectSearchType returns the search type to perfrom ("regexp", or
// "literal"). The search type derives from three sources: the version and
// patternType parameters passed to the search endpoint (literal search is the
// default in V2), and the `patternType:` filter in the input query string which
// overrides the searchType, if present.
func detectSearchType(version string, patternType *string, input string) (query.SearchType, error) {
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
			return -1, fmt.Errorf("unrecognized version: %v", version)
		}
	}

	// The patterntype field is Singular, but not enforced since we do not
	// properly parse the input. The regex extraction, takes the left-most
	// "patterntype:value" match.
	var patternTypeRegex = lazyregexp.New(`(?i)patterntype:([a-zA-Z"']+)`)
	patternFromField := patternTypeRegex.FindStringSubmatch(input)
	if len(patternFromField) > 1 {
		extracted := patternFromField[1]
		if match, _ := regexp.MatchString("regex", extracted); match {
			searchType = query.SearchTypeRegex
		} else if match, _ := regexp.MatchString("literal", extracted); match {
			searchType = query.SearchTypeLiteral

		} else if match, _ := regexp.MatchString("structural", extracted); match {
			searchType = query.SearchTypeStructural
		}
	}

	return searchType, nil
}

// searchResolver is a resolver for the GraphQL type `Search`
type searchResolver struct {
	query         *query.Query          // the validated search query
	parseTree     syntax.ParseTree      // the parsed search query
	originalQuery string                // the raw string of the original search query
	pagination    *searchPaginationInfo // pagination information, or nil if the request is not paginated.
	patternType   query.SearchType

	// Cached resolveRepositories results.
	reposMu                   sync.Mutex
	repoRevs, missingRepoRevs []*search.RepositoryRevisions
	repoOverLimit             bool
	repoErr                   error

	zoekt        *searchbackend.Zoekt
	searcherURLs *endpoint.Map
}

// rawQuery returns the original query string input.
func (r *searchResolver) rawQuery() string {
	return r.originalQuery
}

func (r *searchResolver) countIsSet() bool {
	count, _ := r.query.StringValues(query.FieldCount)
	max, _ := r.query.StringValues(query.FieldMax)
	return len(count) > 0 || len(max) > 0
}

const defaultMaxSearchResults = 30

func (r *searchResolver) maxResults() int32 {
	if r.pagination != nil {
		// Paginated search requests always consume an entire result set for a
		// given repository, so we do not want any limit here. See
		// search_pagination.go for details on why this is necessary .
		return math.MaxInt32
	}
	count, _ := r.query.StringValues(query.FieldCount)
	if len(count) > 0 {
		n, _ := strconv.Atoi(count[0])
		if n > 0 {
			return int32(n)
		}
	}
	max, _ := r.query.StringValues(query.FieldMax)
	if len(max) > 0 {
		n, _ := strconv.Atoi(max[0])
		if n > 0 {
			return int32(n)
		}
	}
	return defaultMaxSearchResults
}

var mockResolveRepoGroups func() (map[string][]*types.Repo, error)

func resolveRepoGroups(ctx context.Context) (map[string][]*types.Repo, error) {
	if mockResolveRepoGroups != nil {
		return mockResolveRepoGroups()
	}

	groups := map[string][]*types.Repo{}

	// Repo groups can be defined in the search.repoGroups settings field.
	merged, err := viewerFinalSettings(ctx)
	if err != nil {
		return nil, err
	}
	var settings schema.Settings
	if err := json.Unmarshal([]byte(merged.Contents()), &settings); err != nil {
		return nil, err
	}
	for name, repoPaths := range settings.SearchRepositoryGroups {
		repos := make([]*types.Repo, len(repoPaths))
		for i, repoPath := range repoPaths {
			repos[i] = &types.Repo{Name: api.RepoName(repoPath)}
		}
		groups[name] = repos
	}

	return groups, nil
}

// Cf. golang/go/src/regexp/syntax/parse.go.
const regexpFlags regexpsyntax.Flags = regexpsyntax.ClassNL | regexpsyntax.PerlX | regexpsyntax.UnicodeGroups

// exactlyOneRepo returns whether exactly one repo: literal field is specified and
// delineated by regex anchors ^ and $. This function helps determine whether we
// should return results for a single repo regardless of whether it is a fork or
// archive.
func exactlyOneRepo(repoFilters []string) bool {
	if len(repoFilters) == 1 {
		filter := repoFilters[0]
		if strings.HasPrefix(filter, "^") && strings.HasSuffix(filter, "$") {
			filter := strings.TrimSuffix(strings.TrimPrefix(filter, "^"), "$")
			r, err := regexpsyntax.Parse(filter, regexpFlags)
			if err != nil {
				return false
			}
			return r.Op == regexpsyntax.OpLiteral
		}
	}
	return false
}

// resolveRepositories calls doResolveRepositories, caching the result for the common
// case where effectiveRepoFieldValues == nil.
func (r *searchResolver) resolveRepositories(ctx context.Context, effectiveRepoFieldValues []string) (repoRevs, missingRepoRevs []*search.RepositoryRevisions, overLimit bool, err error) {
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
		if r.repoRevs != nil || r.missingRepoRevs != nil || r.repoErr != nil {
			tr.LazyPrintf("cached")
			return r.repoRevs, r.missingRepoRevs, r.repoOverLimit, r.repoErr
		}
	}

	repoFilters, minusRepoFilters := r.query.RegexpPatterns(query.FieldRepo)
	if effectiveRepoFieldValues != nil {
		repoFilters = effectiveRepoFieldValues
	}
	repoGroupFilters, _ := r.query.StringValues(query.FieldRepoGroup)

	forkStr, _ := r.query.StringValue(query.FieldFork)
	fork := parseYesNoOnly(forkStr)
	if fork == Invalid && !exactlyOneRepo(repoFilters) {
		fork = No // fork defaults to No unless exactly one repo is being searched.
	}

	archivedStr, _ := r.query.StringValue(query.FieldArchived)
	archived := parseYesNoOnly(archivedStr)
	if archived == Invalid && !exactlyOneRepo(repoFilters) {
		archived = No // archived defaults to No unless exactly one repo is being searched.
	}

	commitAfter, _ := r.query.StringValue(query.FieldRepoHasCommitAfter)

	tr.LazyPrintf("resolveRepositories - start")
	repoRevs, missingRepoRevs, overLimit, err = resolveRepositories(ctx, resolveRepoOp{
		repoFilters:      repoFilters,
		minusRepoFilters: minusRepoFilters,
		repoGroupFilters: repoGroupFilters,
		onlyForks:        fork == Only || fork == True,
		noForks:          fork == No || fork == False,
		onlyArchived:     archived == Only || archived == True,
		noArchived:       archived == No || archived == False,
		commitAfter:      commitAfter,
	})
	tr.LazyPrintf("resolveRepositories - done")
	if effectiveRepoFieldValues == nil {
		r.repoRevs = repoRevs
		r.missingRepoRevs = missingRepoRevs
		r.repoOverLimit = overLimit
		r.repoErr = err
	}
	return repoRevs, missingRepoRevs, overLimit, err
}

// a patternRevspec maps an include pattern to a list of revisions
// for repos matching that pattern. "map" in this case does not mean
// an actual map, because we want regexp matches, not identity matches.
type patternRevspec struct {
	includePattern *regexp.Regexp
	revs           []search.RevisionSpecifier
}

// given a repo name, determine whether it matched any patterns for which we have
// revspecs (or ref globs), and if so, return the matching/allowed ones.
func getRevsForMatchedRepo(repo api.RepoName, pats []patternRevspec) (matched []search.RevisionSpecifier, clashing []search.RevisionSpecifier) {
	revLists := make([][]search.RevisionSpecifier, 0, len(pats))
	for _, rev := range pats {
		if rev.includePattern.MatchString(string(repo)) {
			revLists = append(revLists, rev.revs)
		}
	}
	// exactly one match: we accept that list
	if len(revLists) == 1 {
		matched = revLists[0]
		return
	}
	// no matches: we generate a dummy list containing only master
	if len(revLists) == 0 {
		matched = []search.RevisionSpecifier{{RevSpec: ""}}
		return
	}
	// if two repo specs match, and both provided non-empty rev lists,
	// we want their intersection
	allowedRevs := make(map[search.RevisionSpecifier]struct{}, len(revLists[0]))
	allRevs := make(map[search.RevisionSpecifier]struct{}, len(revLists[0]))
	// starting point: everything is "true" if it is currently allowed
	for _, rev := range revLists[0] {
		allowedRevs[rev] = struct{}{}
		allRevs[rev] = struct{}{}
	}
	// in theory, "master-by-default" entries won't even be participating
	// in this.
	for _, revList := range revLists[1:] {
		restrictedRevs := make(map[search.RevisionSpecifier]struct{}, len(revList))
		for _, rev := range revList {
			allRevs[rev] = struct{}{}
			if _, ok := allowedRevs[rev]; ok {
				restrictedRevs[rev] = struct{}{}
			}
		}
		allowedRevs = restrictedRevs
	}
	if len(allowedRevs) > 0 {
		matched = make([]search.RevisionSpecifier, 0, len(allowedRevs))
		for rev := range allowedRevs {
			matched = append(matched, rev)
		}
		sort.Slice(matched, func(i, j int) bool { return matched[i].Less(matched[j]) })
		return
	}
	// build a list of the revspecs which broke this, return it
	// as the "clashing" list.
	clashing = make([]search.RevisionSpecifier, 0, len(allRevs))
	for rev := range allRevs {
		clashing = append(clashing, rev)
	}
	// ensure that lists are always returned in sorted order.
	sort.Slice(clashing, func(i, j int) bool { return clashing[i].Less(clashing[j]) })
	return
}

// findPatternRevs mutates the given list of include patterns to
// be a raw list of the repository name patterns we want, separating
// out their revision specs, if any.
func findPatternRevs(includePatterns []string) (includePatternRevs []patternRevspec, err error) {
	includePatternRevs = make([]patternRevspec, 0, len(includePatterns))
	for i, includePattern := range includePatterns {
		repoPattern, revs := search.ParseRepositoryRevisions(includePattern)
		// Validate pattern now so the error message is more recognizable to the
		// user
		if _, err := regexp.Compile(string(repoPattern)); err != nil {
			return nil, &badRequestError{err}
		}
		repoPattern = api.RepoName(optimizeRepoPatternWithHeuristics(string(repoPattern)))
		includePatterns[i] = string(repoPattern)
		if len(revs) > 0 {
			p, err := regexp.Compile("(?i:" + includePatterns[i] + ")")
			if err != nil {
				return nil, &badRequestError{err}
			}
			patternRev := patternRevspec{includePattern: p, revs: revs}
			includePatternRevs = append(includePatternRevs, patternRev)
		}
	}
	return
}

type resolveRepoOp struct {
	repoFilters      []string
	minusRepoFilters []string
	repoGroupFilters []string
	noForks          bool
	onlyForks        bool
	noArchived       bool
	onlyArchived     bool
	commitAfter      string
}

func resolveRepositories(ctx context.Context, op resolveRepoOp) (repoRevisions, missingRepoRevisions []*search.RepositoryRevisions, overLimit bool, err error) {
	tr, ctx := trace.New(ctx, "resolveRepositories", fmt.Sprintf("%+v", op))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	includePatterns := op.repoFilters
	if includePatterns != nil {
		// Copy to avoid race condition.
		includePatterns = append([]string{}, includePatterns...)
	}

	excludePatterns := op.minusRepoFilters

	maxRepoListSize := maxReposToSearch()

	// If any repo groups are specified, take the intersection of the repo
	// groups and the set of repos specified with repo:. (If none are specified
	// with repo:, then include all from the group.)
	if groupNames := op.repoGroupFilters; len(groupNames) > 0 {
		groups, err := resolveRepoGroups(ctx)
		if err != nil {
			return nil, nil, false, err
		}
		var patterns []string
		for _, groupName := range groupNames {
			for _, repo := range groups[groupName] {
				patterns = append(patterns, "^"+regexp.QuoteMeta(string(repo.Name))+"$")
			}
		}
		includePatterns = append(includePatterns, unionRegExps(patterns))

		// Ensure we don't omit any repos explicitly included via a repo group.
		if len(patterns) > maxRepoListSize {
			maxRepoListSize = len(patterns)
		}
	}

	// note that this mutates the strings in includePatterns, stripping their
	// revision specs, if they had any.
	includePatternRevs, err := findPatternRevs(includePatterns)
	if err != nil {
		return nil, nil, false, err
	}

	var defaultRepos []*types.Repo
	if envvar.SourcegraphDotComMode() && len(includePatterns) == 0 {
		getIndexedRepos := func(ctx context.Context, revs []*search.RepositoryRevisions) (indexed, unindexed []*search.RepositoryRevisions, err error) {
			return zoektIndexedRepos(ctx, search.Indexed(), revs, nil)
		}
		defaultRepos, err = defaultRepositories(ctx, db.DefaultRepos.List, getIndexedRepos)
		if err != nil {
			return nil, nil, false, errors.Wrap(err, "getting list of default repos")
		}
	}

	var repos []*types.Repo
	if len(defaultRepos) > 0 {
		repos = defaultRepos
		if len(repos) > maxRepoListSize {
			repos = repos[:maxRepoListSize]
		}
	} else {
		tr.LazyPrintf("Repos.List - start")
		repos, err = db.Repos.List(ctx, db.ReposListOptions{
			OnlyRepoIDs:     true,
			IncludePatterns: includePatterns,
			ExcludePattern:  unionRegExps(excludePatterns),
			// List N+1 repos so we can see if there are repos omitted due to our repo limit.
			LimitOffset:  &db.LimitOffset{Limit: maxRepoListSize + 1},
			NoForks:      op.noForks,
			OnlyForks:    op.onlyForks,
			NoArchived:   op.noArchived,
			OnlyArchived: op.onlyArchived,
		})
		tr.LazyPrintf("Repos.List - done")
		if err != nil {
			return nil, nil, false, err
		}
	}
	overLimit = len(repos) >= maxRepoListSize

	repoRevisions = make([]*search.RepositoryRevisions, 0, len(repos))
	tr.LazyPrintf("Associate/validate revs - start")
	for _, repo := range repos {
		revs, clashingRevs := getRevsForMatchedRepo(repo.Name, includePatternRevs)
		repoRev := &search.RepositoryRevisions{Repo: repo}

		// We do in place filtering to reduce allocations. Common path is no
		// filtering of revs.
		if len(revs) > 0 {
			repoRev.Revs = revs[:0]
		}

		// if multiple specified revisions clash, report this usefully:
		if len(revs) == 0 && clashingRevs != nil {
			missingRepoRevisions = append(missingRepoRevisions, &search.RepositoryRevisions{
				Repo: repo,
				Revs: clashingRevs,
			})
		}
		// Check if the repository actually has the revisions that the user specified.
		for _, rev := range revs {
			if rev.RefGlob != "" || rev.ExcludeRefGlob != "" {
				// Do not validate ref patterns. A ref pattern matching 0 refs is not necessarily
				// invalid, so it's not clear what validation would even mean.
			} else if isDefaultBranch := rev.RevSpec == ""; !isDefaultBranch { // skip default branch resolution to save time
				// Validate the revspec.

				// Do not trigger a repo-updater lookup (e.g.,
				// backend.{GitRepo,Repos.ResolveRev}) because that would slow this operation
				// down by a lot (if we're looping over many repos). This means that it'll fail if a
				// repo is not on gitserver.
				//
				// TODO(sqs): make this NOT send gitserver this revspec in EnsureRevision, to avoid
				// searches like "repo:@foobar" (where foobar is an invalid revspec on most repos)
				// taking a long time because they all ask gitserver to try to fetch from the remote
				// repo.
				if _, err := git.ResolveRevision(ctx, repoRev.GitserverRepo(), nil, rev.RevSpec, &git.ResolveRevisionOptions{NoEnsureRevision: true}); gitserver.IsRevisionNotFound(err) || err == context.DeadlineExceeded {
					// The revspec does not exist, so don't include it, and report that it's missing.
					if rev.RevSpec == "" {
						// Report as HEAD not "" (empty string) to avoid user confusion.
						rev.RevSpec = "HEAD"
					}
					missingRepoRevisions = append(missingRepoRevisions, &search.RepositoryRevisions{
						Repo: repo,
						Revs: []search.RevisionSpecifier{{RevSpec: rev.RevSpec}},
					})
					continue
				}
				// If err != nil and is not one of the err values checked for above, cloning and other errors will be handled later, so just ignore an error
				// if there is one.
			}

			repoRev.Revs = append(repoRev.Revs, rev)
		}

		repoRevisions = append(repoRevisions, repoRev)
	}

	tr.LazyPrintf("Associate/validate revs - done")

	if op.commitAfter != "" {
		repoRevisions, err = filterRepoHasCommitAfter(ctx, repoRevisions, op.commitAfter)
	}

	return repoRevisions, missingRepoRevisions, overLimit, err
}

type indexedReposFunc func(ctx context.Context, revs []*search.RepositoryRevisions) (indexed, unindexed []*search.RepositoryRevisions, err error)
type defaultReposFunc func(ctx context.Context) ([]*types.Repo, error)

func defaultRepositories(ctx context.Context, getRawDefaultRepos defaultReposFunc, getIndexedRepos indexedReposFunc) ([]*types.Repo, error) {
	// Get the list of default repos from the db.
	defaultRepos, err := getRawDefaultRepos(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "querying db for default repos")
	}
	// Find out which of the default repos have been indexed.
	defaultRepoRevs := make([]*search.RepositoryRevisions, 0, len(defaultRepos))
	for _, r := range defaultRepos {
		rr := &search.RepositoryRevisions{
			Repo: r,
			Revs: []search.RevisionSpecifier{{RevSpec: ""}},
		}
		defaultRepoRevs = append(defaultRepoRevs, rr)
	}

	indexed, unindexed, err := getIndexedRepos(ctx, defaultRepoRevs)
	if err != nil {
		return nil, errors.Wrap(err, "finding subset of default repos that are indexed")
	}
	// If any are unindexed, log the first few so we can find out if something is going wrong with those.
	if len(unindexed) > 0 {
		N := len(unindexed)
		if N > 10 {
			N = 10
		}
		var names []string
		for i := 0; i < N; i++ {
			names = append(names, string(unindexed[i].Repo.Name))
		}
		log15.Info("some unindexed repos found; listing up to 10 of them", "unindexed", names)
	}
	// Exclude any that aren't indexed.
	indexedMap := make(map[api.RepoID]bool, len(indexed))
	for _, r := range indexed {
		indexedMap[r.Repo.ID] = true
	}
	defaultRepos2 := make([]*types.Repo, 0, len(indexedMap))
	for _, r := range defaultRepos {
		if indexedMap[r.ID] {
			defaultRepos2 = append(defaultRepos2, r)
		}
	}
	return defaultRepos2, nil
}

func filterRepoHasCommitAfter(ctx context.Context, revisions []*search.RepositoryRevisions, after string) ([]*search.RepositoryRevisions, error) {
	var (
		mut  sync.Mutex
		pass = []*search.RepositoryRevisions{}
		res  = make(chan *search.RepositoryRevisions, 100)
		run  = parallel.NewRun(128)
	)

	goroutine.Go(func() {
		for rev := range res {
			if len(rev.Revs) != 0 {
				mut.Lock()
				pass = append(pass, rev)
				mut.Unlock()
			}
			run.Release()
		}
	})

	for _, revs := range revisions {
		run.Acquire()

		revs := revs
		goroutine.Go(func() {
			var specifiers []search.RevisionSpecifier
			for _, rev := range revs.Revs {
				ok, err := git.HasCommitAfter(ctx, revs.GitserverRepo(), after, rev.RevSpec)
				if err != nil {
					if gitserver.IsRevisionNotFound(err) || vcs.IsRepoNotExist(err) {
						continue
					}

					run.Error(err)
					continue
				}
				if ok {
					specifiers = append(specifiers, rev)
				}
			}
			res <- &search.RepositoryRevisions{Repo: revs.Repo, Revs: specifiers}
		})
	}

	err := run.Wait()
	close(res)

	return pass, err
}

func optimizeRepoPatternWithHeuristics(repoPattern string) string {
	if envvar.SourcegraphDotComMode() && strings.HasPrefix(string(repoPattern), "github.com") {
		repoPattern = "^" + repoPattern
	}
	// Optimization: make the "." in "github.com" a literal dot
	// so that the regexp can be optimized more effectively.
	repoPattern = strings.Replace(string(repoPattern), "github.com", `github\.com`, -1)
	return repoPattern
}

func (r *searchResolver) suggestFilePaths(ctx context.Context, limit int) ([]*searchSuggestionResolver, error) {
	repos, _, overLimit, err := r.resolveRepositories(ctx, nil)
	if err != nil {
		return nil, err
	}

	if overLimit {
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
		Repos:           repos,
		Query:           r.query,
		UseFullDeadline: r.searchTimeoutFieldSet(),
		Zoekt:           r.zoekt,
		SearcherURLs:    r.searcherURLs,
	}
	if err := args.PatternInfo.Validate(); err != nil {
		return nil, err
	}

	fileResults, _, err := searchFilesInRepos(ctx, &args)
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

// SearchRepos searches for the provided query but only the the unique list of
// repositories belonging to the search results.
// It's used by campaigns to search.
func SearchRepos(ctx context.Context, plainQuery string) ([]*RepositoryResolver, error) {
	queryString := query.ConvertToLiteral(plainQuery)

	q, err := query.ParseAndCheck(queryString)
	if err != nil {
		return nil, err
	}

	sr := &searchResolver{
		query:         q,
		originalQuery: plainQuery,
		pagination:    nil,
		patternType:   query.SearchTypeLiteral,
		zoekt:         search.Indexed(),
		searcherURLs:  search.SearcherURLs(),
	}

	results, err := sr.Results(ctx)
	if err != nil {
		return nil, err
	}
	return results.Repositories(), nil
}

func unionRegExps(patterns []string) string {
	if len(patterns) == 0 {
		return ""
	}
	if len(patterns) == 1 {
		return patterns[0]
	}

	// We only need to wrap the pattern in parentheses if it contains a "|" because
	// "|" has the lowest precedence of any operator.
	patterns2 := make([]string, len(patterns))
	for i, p := range patterns {
		if strings.Contains(p, "|") {
			p = "(" + p + ")"
		}
		patterns2[i] = p
	}
	return strings.Join(patterns2, "|")
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
		return &searchSuggestionResolver{result: r, score: score, length: len(r.repo.Name), label: string(r.repo.Name)}

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
// updating common as to reflect that new information. If searchErr is a fatal error,
// it returns a non-nil error; otherwise, if searchErr == nil or a non-fatal error, it returns a
// nil error.
func handleRepoSearchResult(common *searchResultsCommon, repoRev *search.RepositoryRevisions, limitHit, timedOut bool, searchErr error) (fatalErr error) {
	common.limitHit = common.limitHit || limitHit
	if vcs.IsRepoNotExist(searchErr) {
		if vcs.IsCloneInProgress(searchErr) {
			common.cloning = append(common.cloning, repoRev.Repo)
		} else {
			common.missing = append(common.missing, repoRev.Repo)
		}
	} else if gitserver.IsRevisionNotFound(searchErr) {
		if len(repoRev.Revs) == 0 || len(repoRev.Revs) == 1 && repoRev.Revs[0].RevSpec == "" {
			// If we didn't specify an input revision, then the repo is empty and can be ignored.
		} else {
			return searchErr
		}
	} else if errcode.IsNotFound(searchErr) {
		common.missing = append(common.missing, repoRev.Repo)
	} else if errcode.IsTimeout(searchErr) || errcode.IsTemporary(searchErr) || timedOut {
		common.timedout = append(common.timedout, repoRev.Repo)
	} else if searchErr != nil {
		return searchErr
	}
	return nil
}

var errMultipleRevsNotSupported = errors.New("not yet supported: searching multiple revs in the same repo")
