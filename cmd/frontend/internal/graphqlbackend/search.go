package graphqlbackend

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"log"

	"github.com/felixfbecker/stringscore"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/envvar"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/pathmatch"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/searchquery"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/searchquery/types"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

var (
	maxReposToSearch, _ = strconv.Atoi(conf.Get().MaxReposToSearch)
)

func init() {
	if maxReposToSearch == 0 {
		// Default to a very large number that will not overflow if incremented.
		maxReposToSearch = int(math.MaxInt32 >> 1)
	}
}

const maxQueryLength = 400

type searchArgs struct {
	// Query is the search query.
	Query string

	// ScopeQuery is the query of the active search scope.
	ScopeQuery string
}

// Search provides search results and suggestions.
func (r *schemaResolver) Search(args *searchArgs) (*searchResolver, error) {
	if len(args.Query)+len(args.ScopeQuery) > maxQueryLength {
		return nil, fmt.Errorf("query exceeds max length (%d)", maxQueryLength)
	}

	combinedQuery, err := searchquery.ParseAndCheck(args.Query + " " + args.ScopeQuery)
	if err != nil {
		return nil, err
	}
	query, err := searchquery.ParseAndCheck(args.Query)
	if err != nil {
		return nil, err
	}
	scopeQuery, err := searchquery.ParseAndCheck(args.ScopeQuery)
	if err != nil {
		return nil, err
	}
	return &searchResolver{
		root:          r,
		args:          *args,
		combinedQuery: *combinedQuery,
		query:         *query,
		scopeQuery:    *scopeQuery,
	}, nil
}

func asString(v *types.Value) string {
	switch {
	case v.String != nil:
		return *v.String
	case v.Regexp != nil:
		return v.Regexp.String()
	default:
		panic("unable to get value as string")
	}
}

type searchResolver struct {
	root *schemaResolver
	args searchArgs

	combinedQuery searchquery.Query // the scope and user query combined (most callers should use this)
	query         searchquery.Query // the user query only
	scopeQuery    searchquery.Query // the scope query only

	// Cached resolveRepositories results.
	reposMu                   sync.Mutex
	repoRevs, missingRepoRevs []*repositoryRevisions
	repoResults               []*searchResultResolver
	repoOverLimit             bool
	repoErr                   error
}

var mockResolveRepoGroups func() (map[string][]*sourcegraph.Repo, error)

func resolveRepoGroups(ctx context.Context) (map[string][]*sourcegraph.Repo, error) {
	if mockResolveRepoGroups != nil {
		return mockResolveRepoGroups()
	}

	var active, inactive []*sourcegraph.Repo
	if len(inactiveReposMap) != 0 {
		var err error
		active, inactive, err = listActiveAndInactive(ctx)
		if err != nil {
			return nil, err
		}
	}

	var sample []*sourcegraph.Repo
	if !envvar.DeploymentOnPrem() {
		var err error
		sample, err = getSampleRepos(ctx)
		if err != nil {
			return nil, err
		}
	}

	return map[string][]*sourcegraph.Repo{
		"active":   active,
		"inactive": inactive,
		"sample":   sample,
	}, nil
}

var (
	sampleReposMu sync.Mutex
	sampleRepos   []*sourcegraph.Repo
)

func getSampleRepos(ctx context.Context) ([]*sourcegraph.Repo, error) {
	sampleReposMu.Lock()
	defer sampleReposMu.Unlock()
	if sampleRepos == nil {
		sampleRepoPaths := []string{
			"github.com/sourcegraph/jsonrpc2",
			"github.com/sourcegraph/javascript-typescript-langserver",
			"github.com/gorilla/mux",
			"github.com/gorilla/schema",
			"github.com/golang/lint",
			"github.com/golang/oauth2",
			"github.com/pallets/flask",
		}
		repos := make([]*sourcegraph.Repo, len(sampleRepoPaths))
		for i, path := range sampleRepoPaths {
			repo, err := backend.Repos.GetByURI(ctx, path)
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
func (r *searchResolver) resolveRepositories(ctx context.Context, effectiveRepoFieldValues []string) (repoRevs, missingRepoRevs []*repositoryRevisions, repoResults []*searchResultResolver, overLimit bool, err error) {
	if effectiveRepoFieldValues == nil {
		r.reposMu.Lock()
		defer r.reposMu.Unlock()
		if r.repoRevs != nil || r.missingRepoRevs != nil || r.repoResults != nil || r.repoErr != nil {
			return r.repoRevs, r.missingRepoRevs, r.repoResults, r.repoOverLimit, r.repoErr
		}
	}

	repoFilters, minusRepoFilters := r.combinedQuery.RegexpPatterns(searchquery.FieldRepo)
	if effectiveRepoFieldValues != nil {
		repoFilters = effectiveRepoFieldValues
	}
	repoGroupFilters, _ := r.combinedQuery.StringValues(searchquery.FieldRepoGroup)

	repoRevs, missingRepoRevs, repoResults, overLimit, err = resolveRepositories(ctx, repoFilters, minusRepoFilters, repoGroupFilters)
	if effectiveRepoFieldValues == nil {
		r.repoRevs = repoRevs
		r.missingRepoRevs = missingRepoRevs
		r.repoResults = repoResults
		r.repoOverLimit = overLimit
		r.repoErr = err
	}
	return repoRevs, missingRepoRevs, repoResults, overLimit, err
}

func resolveRepositories(ctx context.Context, repoFilters []string, minusRepoFilters []string, repoGroupFilters []string) (repoRevisions, missingRepoRevisions []*repositoryRevisions, repoResolvers []*searchResultResolver, overLimit bool, err error) {
	includePatterns := repoFilters
	if includePatterns != nil {
		// Copy to avoid race condition.
		includePatterns = append([]string{}, includePatterns...)
	}
	excludePatterns := minusRepoFilters

	maxRepoListSize := maxReposToSearch

	// If any repo groups are specified, take the intersection of the repo
	// groups and the set of repos specified with repo:. (If none are specified
	// with repo:, then include all from the group.)
	if groupNames := repoGroupFilters; len(groupNames) > 0 {
		groups, err := resolveRepoGroups(ctx)
		if err != nil {
			return nil, nil, nil, false, err
		}
		var patterns []string
		for _, groupName := range groupNames {
			for _, repo := range groups[groupName] {
				patterns = append(patterns, "^"+regexp.QuoteMeta(repo.URI)+"$")
			}
		}
		includePatterns = append(includePatterns, unionRegExps(patterns))

		// Ensure we don't omit any repos explicitly included via a repo group.
		if len(patterns) > maxRepoListSize {
			maxRepoListSize = len(patterns)
		}
	}

	// Treat an include pattern with a suffix of "@rev" as meaning that all
	// matched repos should be resolved to "rev".
	includePatternRevs := make([][]revspecOrRefGlob, len(includePatterns))
	for i, includePattern := range includePatterns {
		repoRev := parseRepositoryRevisions(includePattern)
		repoPattern := repoRev.repo // trim "@rev" from pattern
		// Validate pattern now so the error message is more recognizable to the
		// user
		if _, err := regexp.Compile(repoPattern); err != nil {
			return nil, nil, nil, false, &badRequestError{err}
		}
		// Optimization: make the "." in "github.com" a literal dot
		// so that the regexp can be optimized more effectively.
		if strings.HasPrefix(repoPattern, "github.com") {
			repoPattern = "^" + repoPattern
		}
		repoPattern = strings.Replace(repoPattern, "github.com", `github\.com`, -1)
		includePatterns[i] = repoPattern
		includePatternRevs[i] = repoRev.revs
	}

	// Support determining which include pattern with a rev (if any) matched
	// a repo in the result set.
	compiledIncludePatterns := make([]*regexp.Regexp, len(includePatterns))
	for i, includePattern := range includePatterns {
		p, err := regexp.Compile("(?i:" + includePattern + ")")
		if err != nil {
			return nil, nil, nil, false, &badRequestError{err}
		}
		compiledIncludePatterns[i] = p
	}
	getRevsForMatchedRepo := func(repo string) []revspecOrRefGlob {
		for i, pat := range compiledIncludePatterns {
			if pat.MatchString(repo) {
				return includePatternRevs[i]
			}
		}
		return nil
	}

	repos, err := backend.Repos.List(ctx, &sourcegraph.RepoListOptions{
		IncludePatterns: includePatterns,
		ExcludePattern:  unionRegExps(excludePatterns),
		// List N+1 repos so we can see if there are repos omitted due to our repo limit.
		ListOptions: sourcegraph.ListOptions{PerPage: int32(maxRepoListSize + 1)},
	})
	if err != nil {
		return nil, nil, nil, false, err
	}
	overLimit = len(repos.Repos) >= maxRepoListSize

	repoRevisions = make([]*repositoryRevisions, 0, len(repos.Repos))
	repoResolvers = make([]*searchResultResolver, 0, len(repos.Repos))
	for _, repo := range repos.Repos {
		repoRev := &repositoryRevisions{
			repo: repo.URI,
			revs: getRevsForMatchedRepo(repo.URI),
		}
		repoResolver := &repositoryResolver{repo: repo}

		if len(repoRev.revspecs()) == 1 {
			// Check if the repository actually has the revision that the user
			// specified.
			//
			// TODO(sqs): make this support multiple revspecs and ref globs
			_, err := repoResolver.RevState(ctx, &struct {
				Rev string
			}{
				Rev: repoRev.revSpecsOrDefaultBranch()[0],
			})
			if err == vcs.ErrRevisionNotFound {
				// revision does not exist, so do not include this repository.
				missingRepoRevisions = append(missingRepoRevisions, repoRev)
				continue
			}
			// else, real errors will be handled later, so just ignore it.
		}

		repoResolvers = append(repoResolvers, newSearchResultResolver(
			repoResolver,
			math.MaxInt32,
		))
		repoRevisions = append(repoRevisions, repoRev)
	}

	return repoRevisions, missingRepoRevisions, repoResolvers, overLimit, nil
}

func (r *searchResolver) resolveFiles(ctx context.Context, limit int) ([]*searchResultResolver, error) {
	repoRevisions, _, _, overLimit, err := r.resolveRepositories(ctx, nil)
	if err != nil {
		return nil, err
	}

	if overLimit {
		// If we've exceeded the repo limit, then we may miss files from repos we care
		// about, so don't bother searching filenames at all.
		return nil, nil
	}

	includePatterns, excludePatterns := r.combinedQuery.RegexpPatterns(searchquery.FieldFile)
	excludePattern := unionRegExps(excludePatterns)
	pathOptions := pathmatch.CompileOptions{
		RegExp:        true,
		CaseSensitive: r.combinedQuery.IsCaseSensitive(),
	}

	// If a single term is specified in the user query, and no other file patterns,
	// then treat it as an include pattern (which is a nice UX for users).
	if vs := r.query.Values(searchquery.FieldDefault); len(vs) == 1 {
		includePatterns = append(includePatterns, asString(vs[0]))
	}

	matchPath, err := pathmatch.CompilePathPatterns(includePatterns, excludePattern, pathOptions)
	if err != nil {
		return nil, &badRequestError{err}
	}

	var scorerQuery string
	if len(includePatterns) > 0 {
		// Try to extract the text-only (non-regexp) part of the query to
		// pass to stringscore, which doesn't use regexps. This is best-effort.
		scorerQuery = strings.TrimSuffix(strings.TrimPrefix(includePatterns[0], "^"), "$")
	}
	matcher := matcher{
		match:       matchPath.MatchPath,
		scorerQuery: scorerQuery,
	}

	return searchTree(ctx, matcher, repoRevisions, limit)
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

type searchResultResolver struct {
	// result is either a repositoryResolver or a fileResolver
	result interface{}
	// score defines how well this item matches the query for sorting purposes
	score int
	// length holds the length of the item name as a second sorting criterium
	length int
	// label to sort alphabetically by when all else is equal.
	label string
}

func (r *searchResultResolver) ToRepository() (*repositoryResolver, bool) {
	res, ok := r.result.(*repositoryResolver)
	return res, ok
}

func (r *searchResultResolver) ToFile() (*fileResolver, bool) {
	res, ok := r.result.(*fileResolver)
	return res, ok
}

// A matcher describes how to filter and score results (for repos and files).
// Exactly one of (query) and (match, scoreQuery) must be set.
type matcher struct {
	query string // query to match using stringscore algorithm

	match       func(path string) bool // func that returns true if the item matches
	scorerQuery string                 // effective query to use in stringscore algorithm
}

// searchTree searches the specified repositories for files and dirs whose name matches the matcher.
func searchTree(ctx context.Context, matcher matcher, repos []*repositoryRevisions, limit int) ([]*searchResultResolver, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		resMu sync.Mutex
		res   []*searchResultResolver
	)
	done := make(chan error, len(repos))
	for _, repoRev := range repos {
		if len(repoRev.revs) >= 2 {
			return nil, errMultipleRevsNotSupported
		}

		go func(repoRev repositoryRevisions) {
			fileResults, err := searchTreeForRepo(ctx, matcher, repoRev.repo, repoRev.revSpecsOrDefaultBranch()[0], limit)
			if err != nil {
				done <- err
				return
			}
			resMu.Lock()
			res = append(res, fileResults...)
			resMu.Unlock()
			done <- nil
		}(*repoRev)
	}
	for range repos {
		if err := <-done; err != nil {
			// TODO collect error
			log.Println("searchFiles error: " + err.Error())
		}
	}
	return res, nil
}

var mockSearchFilesForRepo func(matcher matcher, repoURI string, limit int) ([]*searchResultResolver, error)

// searchTreeForRepo searches the specified repository for files whose name matches
// the matcher
func searchTreeForRepo(ctx context.Context, matcher matcher, repoPath, rev string, limit int) (res []*searchResultResolver, err error) {
	if mockSearchFilesForRepo != nil {
		return mockSearchFilesForRepo(matcher, repoPath, limit)
	}

	repo, err := backend.Repos.GetByURI(ctx, repoPath)
	if err != nil {
		return nil, err
	}
	repoResolver := &repositoryResolver{repo: repo}
	commitStateResolver, err := repoResolver.Commit(ctx, &struct {
		Rev string
	}{Rev: rev})
	if err != nil {
		return nil, err
	}
	if commitStateResolver.cloneInProgress {
		// TODO report a cloning repo
		return res, nil
	}
	commitResolver := commitStateResolver.Commit()
	if commitResolver == nil {
		return nil, fmt.Errorf("unable to resolve commit for repo %s", repoPath)
	}
	treeResolver, err := commitResolver.Tree(ctx, &struct {
		Path      string
		Recursive bool
	}{Path: "", Recursive: true})
	if err != nil {
		return nil, err
	}

	var scorerQuery string
	if matcher.query != "" {
		scorerQuery = matcher.query
	} else {
		scorerQuery = matcher.scorerQuery
	}

	scorer := newScorer(scorerQuery)
	for _, fileResolver := range treeResolver.Entries() {
		score := scorer.calcScore(fileResolver)
		if score <= 0 && matcher.scorerQuery != "" && matcher.match(fileResolver.path) {
			score = 1 // minimum to ensure everything included by match.match is included
		}
		if score > 0 {
			res = append(res, newSearchResultResolver(fileResolver, score))
		}
	}

	sort.Sort(searchResultSorter(res))
	if len(res) > limit {
		res = res[:limit]
	}

	return res, nil
}

// newSearchResultResolver returns a new searchResultResolver wrapping the
// given result.
//
// A panic occurs if the type of result is not a *repositoryResolver or
// *fileResolver.
func newSearchResultResolver(result interface{}, score int) *searchResultResolver {
	switch r := result.(type) {
	case *repositoryResolver:
		return &searchResultResolver{result: r, score: score, length: len(r.repo.URI), label: r.repo.URI}

	case *fileResolver:
		return &searchResultResolver{result: r, score: score, length: len(r.name), label: r.name}

	default:
		panic("never here")
	}
}

// scorer is a structure for holding some scorer state that can be shared
// across calcScore calls for the same query string.
type scorer struct {
	query      string
	queryEmpty bool
	queryParts []string
}

// newScorer returns a scorer to be used for calculating sort scores of results
// against the specified query.
func newScorer(query string) *scorer {
	return &scorer{
		query:      query,
		queryEmpty: strings.TrimSpace(query) == "",
		queryParts: splitNoEmpty(query, "/"),
	}
}

// score values to add to different types of results to e.g. get forks lower in
// search results, etc.
const (
	// Files > Repos > Forks
	scoreBumpFile = 1 * (math.MaxInt32 / 16)
	scoreBumpRepo = 0 * (math.MaxInt32 / 16)
	scoreBumpFork = -10
)

// calcScore calculates and assigns the sorting score to the given result.
//
// A panic occurs if the type of result is not a *repositoryResolver or
// *fileResolver.
func (s *scorer) calcScore(result interface{}) int {
	var score int
	if s.queryEmpty {
		// If no query, then it will show *all* results; score must be nonzero in order to
		// have scoreBump* constants applied.
		score = 1
	}

	switch r := result.(type) {
	case *repositoryResolver:
		if !s.queryEmpty {
			score = postfixFuzzyAlignScore(splitNoEmpty(r.repo.URI, "/"), s.queryParts)
		}
		// Push forks down
		if r.repo.Fork {
			score += scoreBumpFork
		}
		if score > 0 {
			score += scoreBumpRepo
		}
		return score

	case *fileResolver:
		if !s.queryEmpty {
			pathParts := splitNoEmpty(r.path, "/")
			score = postfixFuzzyAlignScore(pathParts, s.queryParts)
		}
		if score > 0 {
			score += scoreBumpFile
		}
		return score

	default:
		panic("never here")
	}
}

// postfixFuzzyAlignScore is used to calculate how well a targets component
// matches a query from the back. It rewards consecutive alignment as well as
// aligning to the right. For example for the query "a/b" we get the
// following ranking:
//
//   /a/b == /x/a/b
//   /a/b/x
//   /a/x/b
//
// The following will get zero score
//
//   /x/b
//   /ab/
func postfixFuzzyAlignScore(targetParts, queryParts []string) int {
	total := 0
	consecutive := true
	queryIdx := len(queryParts) - 1
	for targetIdx := len(targetParts) - 1; targetIdx >= 0 && queryIdx >= 0; targetIdx-- {
		score := stringscore.Score(targetParts[targetIdx], queryParts[queryIdx])
		if score <= 0 {
			consecutive = false
			continue
		}
		// Consecutive and align bonus
		if consecutive {
			score *= 2
		}
		consecutive = true
		total += score
		queryIdx--
	}
	// Did not match whole of queryIdx
	if queryIdx >= 0 {
		return 0
	}
	return total
}

// splitNoEmpty is like strings.Split except empty strings are removed.
func splitNoEmpty(s, sep string) []string {
	split := strings.Split(s, sep)
	res := make([]string, 0, len(split))
	for _, part := range split {
		if part != "" {
			res = append(res, part)
		}
	}
	return res
}

// searchResultSorter implements the sort.Interface interface to sort a list of
// searchResultResolvers.
type searchResultSorter []*searchResultResolver

func (s searchResultSorter) Len() int      { return len(s) }
func (s searchResultSorter) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s searchResultSorter) Less(i, j int) bool {
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
}
