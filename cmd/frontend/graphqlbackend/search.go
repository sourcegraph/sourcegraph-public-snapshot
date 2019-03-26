package graphqlbackend

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	zoektrpc "github.com/google/zoekt/rpc"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/inventory/filelang"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search/query"
	searchquerytypes "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search/query/types"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/endpoint"
	"github.com/sourcegraph/sourcegraph/pkg/env"
	"github.com/sourcegraph/sourcegraph/pkg/errcode"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/jsonc"
	searchbackend "github.com/sourcegraph/sourcegraph/pkg/search/backend"
	"github.com/sourcegraph/sourcegraph/pkg/trace"
	"github.com/sourcegraph/sourcegraph/pkg/vcs"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
	"github.com/sourcegraph/sourcegraph/schema"
	"gopkg.in/inconshreveable/log15.v2"
)

// This file contains the root resolver for search. It currently has a lot of
// logic that spans out into all the other search_* files.
//
// NOTE: This file and most supporting code will be deleted when search2.go is
// rolled out. However, right now to understand search and fix bugs in it you
// should start here.

func maxReposToSearch() int {
	switch max := conf.Get().MaxReposToSearch; max {
	case 0:
		// Not specified OR specified as literal zero. Use our default value.
		return 500
	case -1:
		// Default to a very large number that will not overflow if incremented.
		return math.MaxInt32 >> 1
	default:
		return max
	}
}

// Search provides search results and suggestions.
func (r *schemaResolver) Search(args *struct {
	Query string
}) (interface {
	Results(context.Context) (*searchResultsResolver, error)
	Suggestions(context.Context, *searchSuggestionsArgs) ([]*searchSuggestionResolver, error)
	//lint:ignore U1000 is used by graphql via reflection
	Stats(context.Context) (*searchResultsStats, error)
}, error) {

	go addQueryToSearchesTable(args.Query)
	query, err := query.ParseAndCheck(args.Query)
	if err != nil {
		log15.Debug("graphql search failed to parse", "query", args.Query, "error", err)
		return nil, err
	}
	return &searchResolver{
		query: query,
	}, nil
}

func addQueryToSearchesTable(q string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := db.Searches.Add(ctx, q, 1e5); err != nil {
		log15.Error(`adding query to searches table: %v`, err)
	}
}

func asString(v *searchquerytypes.Value) string {
	switch {
	case v.String != nil:
		return *v.String
	case v.Regexp != nil:
		return v.Regexp.String()
	default:
		panic("unable to get value as string")
	}
}

// searchResolver is a resolver for the GraphQL type `Search`
type searchResolver struct {
	query *query.Query // the parsed search query

	// Cached resolveRepositories results.
	reposMu                   sync.Mutex
	repoRevs, missingRepoRevs []*search.RepositoryRevisions
	repoResults               []*searchSuggestionResolver
	repoOverLimit             bool
	repoErr                   error
}

// rawQuery returns the original query string input.
func (r *searchResolver) rawQuery() string {
	return r.query.Syntax.Input
}

func (r *searchResolver) countIsSet() bool {
	count, _ := r.query.StringValues(query.FieldCount)
	max, _ := r.query.StringValues(query.FieldMax)
	return len(count) > 0 || len(max) > 0
}

const defaultMaxSearchResults = 30

func (r *searchResolver) maxResults() int32 {
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

	if envvar.SourcegraphDotComMode() {
		sampleRepos, err := getSampleRepos(ctx)
		if err != nil {
			return nil, err
		}
		groups["sample"] = sampleRepos
	}

	return groups, nil
}

var (
	sampleReposMu sync.Mutex
	sampleRepos   []*types.Repo
)

func getSampleRepos(ctx context.Context) ([]*types.Repo, error) {
	sampleReposMu.Lock()
	defer sampleReposMu.Unlock()
	if sampleRepos == nil {
		sampleRepoPaths := []api.RepoName{
			"github.com/sourcegraph/jsonrpc2",
			"github.com/sourcegraph/javascript-typescript-langserver",
			"github.com/gorilla/mux",
			"github.com/gorilla/schema",
			"github.com/golang/lint",
			"github.com/golang/oauth2",
			"github.com/pallets/flask",
		}
		repos := make([]*types.Repo, len(sampleRepoPaths))
		for i, path := range sampleRepoPaths {
			repo, err := backend.Repos.GetByName(ctx, path)
			if err != nil {
				return nil, fmt.Errorf("get %q: %s", path, err)
			}
			repos[i] = repo
		}
		sampleRepos = repos
	}
	return sampleRepos, nil
}

// resolveRepositories calls doResolveRepositories, caching the result for the common
// case where effectiveRepoFieldValues == nil.
func (r *searchResolver) resolveRepositories(ctx context.Context, effectiveRepoFieldValues []string) (repoRevs, missingRepoRevs []*search.RepositoryRevisions, repoResults []*searchSuggestionResolver, overLimit bool, err error) {
	tr, ctx := trace.New(ctx, "graphql.resolveRepositories", fmt.Sprintf("effectiveRepoFieldValues: %v", effectiveRepoFieldValues))
	defer func() {
		if err != nil {
			tr.SetError(err)
		} else {
			tr.LazyPrintf("numRepoRevs: %d, numMissingRepoRevs: %d, numRepoResults: %d, overLimit: %v", len(repoRevs), len(missingRepoRevs), len(repoResults), overLimit)
		}
		tr.Finish()
	}()
	if effectiveRepoFieldValues == nil {
		r.reposMu.Lock()
		defer r.reposMu.Unlock()
		if r.repoRevs != nil || r.missingRepoRevs != nil || r.repoResults != nil || r.repoErr != nil {
			tr.LazyPrintf("cached")
			return r.repoRevs, r.missingRepoRevs, r.repoResults, r.repoOverLimit, r.repoErr
		}
	}

	repoFilters, minusRepoFilters := r.query.RegexpPatterns(query.FieldRepo)
	if effectiveRepoFieldValues != nil {
		repoFilters = effectiveRepoFieldValues
	}
	repoGroupFilters, _ := r.query.StringValues(query.FieldRepoGroup)

	// HACK: Demo mode for dotGo on 2019 Mar 25 to make it easier to use Sourcegraph.com as a demo
	// of a self-hosted instance with a limited set of repositories. This limits a user to the
	// repositories specified in their setting `search.defaultRepositories` if there are no
	// repository filters in the search.
	//
	// TODO(sqs): Remove this after 2019 Mar 25.
	if len(repoFilters) == 0 && len(repoGroupFilters) == 0 && envvar.SourcegraphDotComMode() {
		final, err := viewerFinalSettings(ctx)
		if err != nil {
			log.Println(err)
		} else {
			var settings struct {
				SearchDefaultRepositories []string `json:"search.defaultRepositories"`
			}
			if err := jsonc.Unmarshal(final.Contents(), &settings); err != nil {
				log.Println(err)
			}
			if len(settings.SearchDefaultRepositories) > 0 {
				repoFilters = settings.SearchDefaultRepositories
			}
		}
	}

	forkStr, _ := r.query.StringValue(query.FieldFork)
	fork := parseYesNoOnly(forkStr)

	archivedStr, _ := r.query.StringValue(query.FieldArchived)
	archived := parseYesNoOnly(archivedStr)

	tr.LazyPrintf("resolveRepositories - start")
	repoRevs, missingRepoRevs, repoResults, overLimit, err = resolveRepositories(ctx, resolveRepoOp{
		repoFilters:      repoFilters,
		minusRepoFilters: minusRepoFilters,
		repoGroupFilters: repoGroupFilters,
		onlyForks:        fork == Only || fork == True,
		noForks:          fork == No || fork == False,
		onlyArchived:     archived == Only || archived == True,
		noArchived:       archived == No || archived == False,
	})
	tr.LazyPrintf("resolveRepositories - done")
	if effectiveRepoFieldValues == nil {
		r.repoRevs = repoRevs
		r.missingRepoRevs = missingRepoRevs
		r.repoResults = repoResults
		r.repoOverLimit = overLimit
		r.repoErr = err
	}
	return repoRevs, missingRepoRevs, repoResults, overLimit, err
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
}

func resolveRepositories(ctx context.Context, op resolveRepoOp) (repoRevisions, missingRepoRevisions []*search.RepositoryRevisions, repoResolvers []*searchSuggestionResolver, overLimit bool, err error) {
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
			return nil, nil, nil, false, err
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
		return nil, nil, nil, false, err
	}

	tr.LazyPrintf("Repos.List - start")
	repos, err := backend.Repos.List(ctx, db.ReposListOptions{
		IncludePatterns: includePatterns,
		ExcludePattern:  unionRegExps(excludePatterns),
		Enabled:         true,
		// List N+1 repos so we can see if there are repos omitted due to our repo limit.
		LimitOffset:  &db.LimitOffset{Limit: maxRepoListSize + 1},
		NoForks:      op.noForks,
		OnlyForks:    op.onlyForks,
		NoArchived:   op.noArchived,
		OnlyArchived: op.onlyArchived,
	})
	tr.LazyPrintf("Repos.List - done")
	if err != nil {
		return nil, nil, nil, false, err
	}
	overLimit = len(repos) >= maxRepoListSize

	repoRevisions = make([]*search.RepositoryRevisions, 0, len(repos))
	repoResolvers = make([]*searchSuggestionResolver, 0, len(repos))
	tr.LazyPrintf("Associate/validate revs - start")
	for _, repo := range repos {
		repoRev := &search.RepositoryRevisions{Repo: repo}

		revs, clashingRevs := getRevsForMatchedRepo(repo.Name, includePatternRevs)

		repoResolver := &repositoryResolver{repo: repo}

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
				if _, err := git.ResolveRevision(ctx, repoRev.GitserverRepo(), nil, rev.RevSpec, &git.ResolveRevisionOptions{NoEnsureRevision: true}); git.IsRevisionNotFound(err) || err == context.DeadlineExceeded {
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

		repoResolvers = append(repoResolvers, newSearchResultResolver(
			repoResolver,
			math.MaxInt32,
		))
		repoRevisions = append(repoRevisions, repoRev)
	}
	tr.LazyPrintf("Associate/validate revs - done")

	return repoRevisions, missingRepoRevisions, repoResolvers, overLimit, nil
}

func optimizeRepoPatternWithHeuristics(repoPattern string) string {
	// Optimization: make the "." in "github.com" a literal dot
	// so that the regexp can be optimized more effectively.
	if strings.HasPrefix(string(repoPattern), "github.com") {
		repoPattern = "^" + repoPattern
	}
	repoPattern = strings.Replace(string(repoPattern), "github.com", `github\.com`, -1)
	return repoPattern
}

func (r *searchResolver) suggestFilePaths(ctx context.Context, limit int) ([]*searchSuggestionResolver, error) {
	repos, _, _, overLimit, err := r.resolveRepositories(ctx, nil)
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
	args := search.Args{
		Pattern:         p,
		Repos:           repos,
		Query:           r.query,
		UseFullDeadline: r.searchTimeoutFieldSet(),
	}
	if err := args.Pattern.Validate(); err != nil {
		return nil, err
	}

	fileResults, _, err := searchFilesInRepos(ctx, &args)
	if err != nil {
		return nil, err
	}

	var suggestions []*searchSuggestionResolver
	for i, result := range fileResults {
		assumedScore := len(fileResults) - i // Greater score is first, so we inverse the index.
		suggestions = append(suggestions, newSearchResultResolver(result.File(), assumedScore))
	}
	return suggestions, nil
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
	// result is either a repositoryResolver or a gitTreeEntryResolver
	result interface{}
	// score defines how well this item matches the query for sorting purposes
	score int
	// length holds the length of the item name as a second sorting criterium
	length int
	// label to sort alphabetically by when all else is equal.
	label string
}

func (r *searchSuggestionResolver) ToRepository() (*repositoryResolver, bool) {
	res, ok := r.result.(*repositoryResolver)
	return res, ok
}

func (r *searchSuggestionResolver) ToFile() (*gitTreeEntryResolver, bool) {
	res, ok := r.result.(*gitTreeEntryResolver)
	return res, ok
}

func (r *searchSuggestionResolver) ToGitBlob() (*gitTreeEntryResolver, bool) {
	res, ok := r.result.(*gitTreeEntryResolver)
	return res, ok && res.stat.Mode().IsRegular()
}

func (r *searchSuggestionResolver) ToGitTree() (*gitTreeEntryResolver, bool) {
	res, ok := r.result.(*gitTreeEntryResolver)
	return res, ok && res.stat.Mode().IsDir()
}

func (r *searchSuggestionResolver) ToSymbol() (*symbolResolver, bool) {
	res, ok := r.result.(*symbolResolver)
	return res, ok
}

// newSearchResultResolver returns a new searchResultResolver wrapping the
// given result.
//
// A panic occurs if the type of result is not a *repositoryResolver or
// *gitTreeEntryResolver.
func newSearchResultResolver(result interface{}, score int) *searchSuggestionResolver {
	switch r := result.(type) {
	case *repositoryResolver:
		return &searchSuggestionResolver{result: r, score: score, length: len(r.repo.Name), label: string(r.repo.Name)}

	case *gitTreeEntryResolver:
		return &searchSuggestionResolver{result: r, score: score, length: len(r.path), label: r.path}

	case *symbolResolver:
		return &searchSuggestionResolver{result: r, score: score, length: len(r.symbol.Name + " " + r.symbol.Parent), label: r.symbol.Name + " " + r.symbol.Parent}

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

// langIncludeExcludePatterns returns regexps for the include/exclude path patterns given the lang:
// and -lang: filter values in a search query. For example, a query containing "lang:go" should
// include files whose paths match /\.go$/.
func langIncludeExcludePatterns(values, negatedValues []string) (includePatterns, excludePatterns []string, err error) {
	lookup := func(value string) *filelang.Language {
		value = strings.ToLower(value)
		for _, lang := range filelang.Langs {
			if strings.ToLower(lang.Name) == value {
				return lang
			}
			for _, alias := range lang.Aliases {
				if alias == value {
					return lang
				}
			}
		}
		return nil
	}

	do := func(values []string, patterns *[]string) error {
		for _, value := range values {
			lang := lookup(value)
			if lang == nil {
				return fmt.Errorf("unknown language: %q", value)
			}
			extPatterns := make([]string, len(lang.Extensions))
			for i, ext := range lang.Extensions {
				// Add `\.ext$` pattern to match files with the given extension.
				extPatterns[i] = regexp.QuoteMeta(ext) + "$"
			}
			*patterns = append(*patterns, unionRegExps(extPatterns))
		}
		return nil
	}

	if err := do(values, &includePatterns); err != nil {
		return nil, nil, err
	}
	if err := do(negatedValues, &excludePatterns); err != nil {
		return nil, nil, err
	}
	return includePatterns, excludePatterns, nil
}

// handleRepoSearchResult handles the limitHit and searchErr returned by a search function,
// updating common as to reflect that new information. If searchErr is a fatal error,
// it returns a non-nil error; otherwise, if searchErr == nil or a non-fatal error, it returns a
// nil error.
func handleRepoSearchResult(common *searchResultsCommon, repoRev search.RepositoryRevisions, limitHit, timedOut bool, searchErr error) (fatalErr error) {
	common.limitHit = common.limitHit || limitHit
	if vcs.IsRepoNotExist(searchErr) {
		if vcs.IsCloneInProgress(searchErr) {
			common.cloning = append(common.cloning, repoRev.Repo)
		} else {
			common.missing = append(common.missing, repoRev.Repo)
		}
	} else if git.IsRevisionNotFound(searchErr) {
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

// SearchProviders contains instances of our search providers.
type SearchProviders struct {
	// Text is our root text searcher.
	Text *searchbackend.Text

	// SearcherURLs is an endpoint map to our searcher service replicas.
	SearcherURLs *endpoint.Map

	// Index is a search.Searcher for Zoekt.
	Index *searchbackend.Zoekt
}

var (
	zoektAddr   = env.Get("ZOEKT_HOST", "indexed-search:80", "host:port of the zoekt instance")
	searcherURL = env.Get("SEARCHER_URL", "k8s+http://searcher:3181", "searcher server URL")

	searchOnce sync.Once
	searchP    *SearchProviders
)

// Search returns instances of our search providers.
func Search() *SearchProviders {
	searchOnce.Do(func() {
		// Zoekt
		index := &searchbackend.Zoekt{}
		if zoektAddr != "" {
			index.Client = zoektrpc.Client(zoektAddr)
		}
		go func() {
			conf.Watch(func() {
				index.SetEnabled(conf.SearchIndexEnabled())
			})
		}()

		// Searcher
		var searcherURLs *endpoint.Map
		if len(strings.Fields(searcherURL)) == 0 {
			searcherURLs = endpoint.Empty(errors.New("a searcher service has not been configured"))
		} else {
			searcherURLs = endpoint.New(searcherURL)
		}

		text := &searchbackend.Text{
			Index: index,
			Fallback: &searchbackend.TextJIT{
				Endpoints: searcherURLs,
				Resolve: func(ctx context.Context, name api.RepoName, spec string) (api.CommitID, error) {
					// Do not trigger a repo-updater lookup (e.g.,
					// backend.{GitRepo,Repos.ResolveRev}) because that would
					// slow this operation down by a lot (if we're looping
					// over many repos). This means that it'll fail if a repo
					// is not on gitserver.
					return git.ResolveRevision(ctx, gitserver.Repo{Name: name}, nil, spec, &git.ResolveRevisionOptions{NoEnsureRevision: true})
				},
			},
		}

		searchP = &SearchProviders{
			Text:         text,
			SearcherURLs: searcherURLs,
			Index:        index,
		}
	})
	return searchP
}
