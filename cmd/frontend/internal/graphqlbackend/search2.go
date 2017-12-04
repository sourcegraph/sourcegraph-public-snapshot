package graphqlbackend

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/envvar"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/pathmatch"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/searchquery"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/searchquery/types"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

var (
	maxReposToSearch, _ = strconv.Atoi(env.Get("MAX_REPOS_TO_SEARCH", "30", `the maximum number of repos to search across (the user is prompted to narrow their query if exceeded)`))
)

const maxQueryLength = 400

type searchArgs2 struct {
	// Query is the search query.
	Query string

	// ScopeQuery is the query of the active search scope.
	ScopeQuery string
}

// Search2 provides search results and suggestions.
func (r *rootResolver) Search2(args *searchArgs2) (*searchResolver2, error) {
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
	return &searchResolver2{
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

type searchResolver2 struct {
	root *rootResolver
	args searchArgs2

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
func (r *searchResolver2) resolveRepositories(ctx context.Context, effectiveRepoFieldValues []string) (repoRevs, missingRepoRevs []*repositoryRevisions, repoResults []*searchResultResolver, overLimit bool, err error) {
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

func (r *searchResolver2) resolveFiles(ctx context.Context, limit int) ([]*searchResultResolver, error) {
	repoRevisions, _, _, overLimit, err := r.resolveRepositories(ctx, nil)
	if err != nil {
		return nil, err
	}

	if overLimit {
		// If we've exceeded the repo limit, then we may miss files from repos we care
		// about, so don't bother searching filenames at all.
		return nil, nil
	}

	repos := make([]string, len(repoRevisions))
	for i, repoRevision := range repoRevisions {
		repos[i] = repoRevision.repo
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

	return searchTree(ctx, matcher, repos, limit)
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
