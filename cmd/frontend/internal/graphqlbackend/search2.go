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
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/pathmatch"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/search2"
)

const (
	maxQueryLength = 200

	searchFieldRepo      search2.Field = "repo"
	searchFieldFile      search2.Field = "file"
	searchFieldRepoGroup search2.Field = "repogroup"
	searchFieldTerm      search2.Field = ""
	searchFieldRegExp    search2.Field = "regexp"
	searchFieldCase      search2.Field = "case"
)

var searchFieldAliases = map[search2.Field][]search2.Field{
	searchFieldRepo:             {"r"},
	searchFieldFile:             {"f"},
	minusField(searchFieldFile): {minusField("f")},
	searchFieldRepoGroup:        {"g"},
	searchFieldTerm:             {},
	searchFieldRegExp:           {"re"},
	searchFieldCase:             {},
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
	query := args.Query
	if args.ScopeQuery != "" {
		query += " " + args.ScopeQuery
	}
	resolvedQuery, err := resolveQuery(query)
	if err != nil {
		return nil, err
	}

	resolvedUserQuery, err := resolveQuery(args.Query)
	if err != nil {
		return nil, err
	}

	return &searchResolver2{
		root:      r,
		args:      *args,
		query:     *resolvedQuery,
		userQuery: *resolvedUserQuery,
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
	fieldValues, unknownFields := tokens.Extract(searchFieldAliases)

	return &resolvedQuery{
		tokens:        tokens,
		fieldValues:   fieldValues,
		unknownFields: unknownFields,
	}, nil
}

type resolvedQuery struct {
	tokens        search2.Tokens
	fieldValues   map[search2.Field][]string
	unknownFields []search2.Field
}

func (q resolvedQuery) isCaseSensitive() bool {
	for _, s := range q.fieldValues[searchFieldCase] {
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

	query     resolvedQuery // the scope and user query combined
	userQuery resolvedQuery // the user query only (ONLY USE for UX hints)
}

func (r *searchResolver2) resolveRepoGroups(ctx context.Context) (map[string][]*sourcegraph.Repo, error) {
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

func (r *searchResolver2) resolveRepositories(ctx context.Context, effectiveRepoFieldValues []string) ([]*repositoryRevision, []*searchResultResolver, error) {
	var includePatterns []string
	if len(effectiveRepoFieldValues) > 0 {
		includePatterns = effectiveRepoFieldValues
	} else {
		includePatterns = r.query.fieldValues[searchFieldRepo]
	}
	excludePatterns := r.query.fieldValues[minusField(searchFieldRepo)]

	maxRepoListSize := 15

	// If any repo groups are specified, take the intersection of the repo
	// groups and the set of repos specified with repo:. (If none are specified
	// with repo:, then include all from the group.)
	if groupNames := r.query.fieldValues[searchFieldRepoGroup]; len(groupNames) > 0 {
		groups, err := r.resolveRepoGroups(ctx)
		if err != nil {
			return nil, nil, err
		}
		var patterns []string
		for _, reposInGroup := range groups {
			for _, repo := range reposInGroup {
				patterns = append(patterns, "^"+regexp.QuoteMeta(repo.URI)+"$")
			}
		}
		includePatterns = append(includePatterns, unionRegExps(patterns))

		// Ensure we don't omit any repos explicitly included via a repo group.
		maxRepoListSize += len(patterns)
	}

	// Treat an include pattern with a suffix of "@rev" as meaning that all
	// matched repos should be resolved to "rev".
	includePatternRevs := make([]string, len(includePatterns))
	for i, includePattern := range includePatterns {
		repoRev := parseRepositoryRevision(includePattern)
		if repoRev.hasRev() {
			repoPattern := repoRev.Repo // trim "@rev" from pattern
			// Optimization: make the "." in "github.com" a literal dot
			// so that the regexp can be optimized more effectively.
			if strings.HasPrefix(repoPattern, "github.com") {
				repoPattern = "^" + repoPattern
			}
			repoPattern = strings.Replace(repoPattern, "github.com", `github\.com`, -1)
			includePatterns[i] = repoPattern
			includePatternRevs[i] = *repoRev.Rev
		}
	}

	// Support determining which include pattern with a rev (if any) matched
	// a repo in the result set.
	compiledIncludePatterns := make([]*regexp.Regexp, len(includePatterns))
	for i, includePattern := range includePatterns {
		p, err := regexp.Compile(includePattern)
		if err != nil {
			return nil, nil, err
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
		ListOptions:     sourcegraph.ListOptions{PerPage: int32(maxRepoListSize)},
	})
	if err != nil {
		return nil, nil, err
	}

	repoRevisions := make([]*repositoryRevision, 0, len(repos.Repos))
	repoResolvers := make([]*searchResultResolver, 0, len(repos.Repos))
	for _, repo := range repos.Repos {
		repoResolvers = append(repoResolvers, newSearchResultResolver(
			&repositoryResolver{repo: repo},
			math.MaxInt32,
		))
		repoRevisions = append(repoRevisions, &repositoryRevision{
			Repo: repo.URI,
			Rev:  getRevForMatchedRepo(repo.URI),
		})
	}

	return repoRevisions, repoResolvers, nil
}

func (r *searchResolver2) resolveFiles(ctx context.Context) ([]*searchResultResolver, error) {
	repoRevisions, _, err := r.resolveRepositories(ctx, nil)
	if err != nil {
		return nil, err
	}
	repos := make([]string, len(repoRevisions))
	for i, repoRevision := range repoRevisions {
		repos[i] = repoRevision.Repo
	}

	// TODO(sqs): make more efficient
	files, err := searchFiles(ctx, "", repos, math.MaxInt32)
	if err != nil {
		return nil, err
	}

	includePatterns := r.query.fieldValues[searchFieldFile]
	excludePattern := unionRegExps(r.query.fieldValues[minusField(searchFieldFile)])
	pathOptions := pathmatch.CompileOptions{
		RegExp:        true,
		CaseSensitive: r.query.isCaseSensitive(),
	}

	// If a single term is specified in the user query, and no other file patterns,
	// then treat it as an include pattern (which is a nice UX for users).
	if len(r.userQuery.fieldValues[searchFieldTerm]) == 1 {
		includePatterns = append(includePatterns, r.userQuery.fieldValues[searchFieldTerm][0])
	}

	matchPath, err := pathmatch.CompilePathPatterns(includePatterns, excludePattern, pathOptions)
	if err != nil {
		return nil, err
	}

	matchingPaths := files[:0]
	for _, file := range files {
		// TODO(sqs): make scorer support multiple queries, use scorer here
		path := file.result.(*fileResolver).path
		if matchPath.MatchPath(path) {
			matchingPaths = append(matchingPaths, file)
		}
	}
	files = matchingPaths

	return files, nil
}

func unionRegExps(patterns []string) string {
	if len(patterns) == 0 {
		return ""
	}
	if len(patterns) == 1 {
		return patterns[0]
	}
	return "(" + strings.Join(patterns, ")|(") + ")"
}

func withoutEmptyStrings(list []string) []string {
	emptyElements := 0
	for _, s := range list {
		if s == "" {
			emptyElements++
		}
	}

	// Only allocate if needed.
	if emptyElements == len(list) {
		return nil
	}
	if emptyElements == 0 {
		return list
	}

	list2 := make([]string, 0, len(list)-emptyElements)
	for _, s := range list {
		if s != "" {
			list2 = append(list2, s)
		}
	}
	return list2
}
