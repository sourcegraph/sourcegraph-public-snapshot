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
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/search2"
)

var (
	maxReposToSearch, _ = strconv.Atoi(env.Get("MAX_REPOS_TO_SEARCH", "30", `the maximum number of repos to search across (the user is prompted to narrow their query if exceeded)`))
)

const (
	maxQueryLength = 200

	searchFieldRepo      search2.Field = "repo"
	searchFieldFile      search2.Field = "file"
	searchFieldRepoGroup search2.Field = "repogroup"
	searchFieldTerm      search2.Field = ""
	searchFieldCase      search2.Field = "case"
	searchFieldType      search2.Field = "type"

	// TODO(sqs): these only apply to type:diff searches
	searchFieldBefore    search2.Field = "before"
	searchFieldAfter     search2.Field = "after"
	searchFieldAuthor    search2.Field = "author"
	searchFieldCommitter search2.Field = "committer"
		searchFieldMessage   search2.Field = "message"
)

var searchFieldAliases = map[search2.Field][]search2.Field{
	searchFieldRepo:                {"r"},
	minusField(searchFieldRepo):    {minusField("r")},
	searchFieldFile:                {"f"},
	minusField(searchFieldFile):    {minusField("f")},
	searchFieldRepoGroup:           {"g"},
	searchFieldTerm:                {},
	searchFieldCase:                {},
	searchFieldMessage:             {"m", "msg"},
	minusField(searchFieldMessage): {minusField("m"), minusField("msg")},
}

func minusField(field search2.Field) search2.Field {
	return search2.Field("-" + field)
}

type searchArgs2 struct {
	// Query is the search query.
	Query string

	// ScopeQuery is the query of the active search scope.
	ScopeQuery string
}

// Search2 provides search results and suggestions.
func (r *rootResolver) Search2(args *searchArgs2) (*searchResolver2, error) {
	combinedQuery, err := resolveQuery(args.Query + " " + args.ScopeQuery)
	if err != nil {
		return nil, err
	}
	query, err := resolveQuery(args.Query)
	if err != nil {
		return nil, err
	}
	scopeQuery, err := resolveQuery(args.ScopeQuery)
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

func resolveQuery(query string) (*resolvedQuery, error) {
	if len(query) > maxQueryLength {
		return nil, fmt.Errorf("query exceeds max length (%d)", maxQueryLength)
	}

	tokens, err := search2.Parse(query)
	if err != nil {
		return nil, err
	}
	tokens.Normalize(searchFieldAliases)
	fieldValues := tokens.Extract()

	return &resolvedQuery{
		tokens:      tokens,
		fieldValues: fieldValues,
	}, nil
}

type resolvedQuery struct {
	tokens      search2.Tokens
	fieldValues map[search2.Field]search2.Values
}

func (q resolvedQuery) isCaseSensitive() bool {
	for _, s := range q.fieldValues[searchFieldCase].Values() {
		v, _ := strconv.ParseBool(s)
		v = v || (s == "yes" || s == "y")
		if v {
			return true
		}
	}
	return false
}

type searchResolver2 struct {
	root *rootResolver
	args searchArgs2

	combinedQuery resolvedQuery // the scope and user query combined (most callers should use this)
	query         resolvedQuery // the user query only
	scopeQuery    resolvedQuery // the scope query only

	// Cached resolveRepositories results.
	reposMu                   sync.Mutex
	repoRevs, missingRepoRevs []*repositoryRevision
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
func (r *searchResolver2) resolveRepositories(ctx context.Context, effectiveRepoFieldValues []string) (repoRevs, missingRepoRevs []*repositoryRevision, repoResults []*searchResultResolver, overLimit bool, err error) {
	if effectiveRepoFieldValues == nil {
		r.reposMu.Lock()
		defer r.reposMu.Unlock()
		if r.repoRevs != nil || r.missingRepoRevs != nil || r.repoResults != nil || r.repoErr != nil {
			return r.repoRevs, r.missingRepoRevs, r.repoResults, r.repoOverLimit, r.repoErr
		}
	}

	repoFilters := effectiveRepoFieldValues
	if repoFilters == nil {
		repoFilters = r.combinedQuery.fieldValues[searchFieldRepo].Values()
	}
	minusRepoFilters := r.combinedQuery.fieldValues[minusField(searchFieldRepo)].Values()
	repoGroupFilters := r.combinedQuery.fieldValues[searchFieldRepoGroup].Values()

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

func resolveRepositories(ctx context.Context, repoFilters []string, minusRepoFilters []string, repoGroupFilters []string) (repoRevisions, missingRepoRevisions []*repositoryRevision, repoResolvers []*searchResultResolver, overLimit bool, err error) {
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
	includePatternRevs := make([]string, len(includePatterns))
	for i, includePattern := range includePatterns {
		repoRev := parseRepositoryRevision(includePattern)
		repoPattern := repoRev.Repo // trim "@rev" from pattern
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
		if repoRev.hasRev() {
			includePatternRevs[i] = *repoRev.Rev
		}
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
	getRevForMatchedRepo := func(repo string) *string {
		for i, pat := range compiledIncludePatterns {
			if pat.MatchString(repo) && includePatternRevs[i] != "" {
				return &includePatternRevs[i]
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

	repoRevisions = make([]*repositoryRevision, 0, len(repos.Repos))
	repoResolvers = make([]*searchResultResolver, 0, len(repos.Repos))
	for _, repo := range repos.Repos {
		repoRev := &repositoryRevision{
			Repo: repo.URI,
			Rev:  getRevForMatchedRepo(repo.URI),
		}
		repoResolver := &repositoryResolver{repo: repo}

		if repoRev.hasRev() {
			// Check if the repository actually has the revision that the user
			// specified.
			_, err := repoResolver.RevState(ctx, &struct {
				Rev string
			}{
				Rev: *repoRev.Rev,
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
		repos[i] = repoRevision.Repo
	}

	includePatterns := r.combinedQuery.fieldValues[searchFieldFile].Values()
	excludePattern := unionRegExps(r.combinedQuery.fieldValues[minusField(searchFieldFile)].Values())
	pathOptions := pathmatch.CompileOptions{
		RegExp:        true,
		CaseSensitive: r.combinedQuery.isCaseSensitive(),
	}

	// If a single term is specified in the user query, and no other file patterns,
	// then treat it as an include pattern (which is a nice UX for users).
	if len(r.query.fieldValues[searchFieldTerm]) == 1 {
		includePatterns = append(includePatterns, r.query.fieldValues[searchFieldTerm][0].Value)
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
