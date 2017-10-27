package graphqlbackend

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/envvar"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/pathmatch"

	"github.com/neelance/parallel"

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

func (r *searchResolver2) resolveRepoGroups(ctx context.Context) (map[string][]*searchResultResolver, error) {
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

	return map[string][]*searchResultResolver{
		"active":   toSearchResultResolvers(active, math.MaxInt32-1),
		"inactive": toSearchResultResolvers(inactive, math.MaxInt32-2),
		"sample":   toSearchResultResolvers(sample, math.MaxInt32-1),
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
			"github.com/Microsoft/vscode",
			"github.com/sourcegraph/go-langserver",
			"github.com/sourcegraph/jsonrpc2",
			"github.com/sourcegraph/javascript-typescript-langserver",
			"github.com/gorilla/mux",
			"github.com/gorilla/schema",
			"github.com/gorilla/securecookie",
			"github.com/gorilla/websocket",
			"github.com/golang/go",
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

func toSearchResultResolvers(repos []*sourcegraph.Repo, score int) []*searchResultResolver {
	resolvers := make([]*searchResultResolver, len(repos))
	for i, repo := range repos {
		resolvers[i] = &searchResultResolver{
			result: &repositoryResolver{repo: repo},
			score:  score,
		}
	}
	return resolvers
}

func (r *searchResolver2) resolveRepositories(ctx context.Context, effectiveRepoFieldValues []string) ([]*repositoryRevision, []*searchResultResolver, error) {
	var (
		mu            sync.Mutex
		repoRevisions []*repositoryRevision
		repoResolvers []*searchResultResolver

		run = parallel.NewRun(8)
	)

	var repoPatterns []string
	if len(effectiveRepoFieldValues) > 0 {
		repoPatterns = effectiveRepoFieldValues
	} else {
		repoPatterns = r.query.fieldValues[searchFieldRepo]
	}

	// Omit empty fields.
	x := make([]string, 0, len(repoPatterns))
	for _, p := range repoPatterns {
		if p != "" {
			x = append(x, p)
		}
	}
	repoPatterns = x

	if len(repoPatterns) == 0 {
		repoPatterns = []string{""} // search across all repos
	}

	for _, repoPattern := range repoPatterns {
		run.Acquire()
		go func(repoPattern string) {
			defer func() {
				if r := recover(); r != nil {
					run.Error(fmt.Errorf("recover: %v", r))
				}
				run.Release()
			}()

			repoRev := parseRepositoryRevision(repoPattern)

			repos, err := searchRepos(ctx, repoRev.Repo, nil, 15) // TODO(sqs): un-hardcode 15
			if err != nil {
				run.Error(err)
				return
			}

			for _, repo := range repos {
				repo.score = math.MaxInt32
			}

			mu.Lock()
			repoResolvers = append(repoResolvers, repos...)
			for _, repo := range repos {
				repoRevisions = append(repoRevisions, &repositoryRevision{
					Repo: repo.result.(*repositoryResolver).URI(),
					Rev:  repoRev.Rev,
				})
			}
			mu.Unlock()
		}(repoPattern)
	}
	if err := run.Wait(); err != nil {
		return nil, nil, err
	}

	// If any repo groups are specified, take the intersection of the repo
	// groups and the set of repos specified with repo:. (If none are specified
	// with repo:, then include all from the group.)
	if groupNames := r.query.fieldValues[searchFieldRepoGroup]; len(groupNames) > 0 {
		reposFromRepoField := make(map[string]struct{}, len(repoRevisions))
		for _, repo := range repoRevisions {
			reposFromRepoField[repo.Repo] = struct{}{}
		}

		groups, err := r.resolveRepoGroups(ctx)
		if err != nil {
			return nil, nil, err
		}
		var repoRevisionsFromGroups []*repositoryRevision
		var repoResolversFromGroups []*searchResultResolver
		reposFromGroups := map[string]struct{}{}
		for _, groupName := range groupNames {
			for _, repo := range groups[groupName] {
				repoRevisionsFromGroups = append(repoRevisionsFromGroups, &repositoryRevision{Repo: repo.result.(*repositoryResolver).URI()})
				repoResolversFromGroups = append(repoResolversFromGroups, repo)
				reposFromGroups[repo.result.(*repositoryResolver).URI()] = struct{}{}
			}
		}

		if len(repoRevisions) == 0 {
			repoRevisions = repoRevisionsFromGroups
			repoResolvers = repoResolversFromGroups
		} else {
			filter := func(repo string) bool {
				_, isInRepoGroup := reposFromGroups[repo]
				return isInRepoGroup
			}
			repoRevisions = append(repoRevisions, repoRevisionsFromGroups...)
			repoResolvers = append(repoResolvers, repoResolversFromGroups...)
			repoRevisions, repoResolvers = filterRepos(repoRevisions, repoResolvers, filter)
		}
	}

	// Eliminate duplicates.
	repoRevisions, repoResolvers = uniqueRepos(repoRevisions, repoResolvers)

	return repoRevisions, repoResolvers, nil
}

func uniqueRepos(repoRevisions []*repositoryRevision, repoResolvers []*searchResultResolver) ([]*repositoryRevision, []*searchResultResolver) {
	if repoRevisions == nil || repoResolvers == nil {
		return nil, nil
	}

	type key struct{ repo, rev string }
	seen := map[key]struct{}{}
	filteredRepoRevisions := repoRevisions[:0]
	filteredRepoResolvers := repoResolvers[:0]
	for i, repo := range repoRevisions {
		k := key{repo: repo.Repo}
		if repo.Rev != nil {
			k.rev = *repo.Rev
		}

		if _, dup := seen[k]; !dup {
			filteredRepoRevisions = append(filteredRepoRevisions, repo)
			filteredRepoResolvers = append(filteredRepoResolvers, repoResolvers[i])
			seen[k] = struct{}{}
		}
	}
	return filteredRepoRevisions, filteredRepoResolvers
}

func filterRepos(repoRevisions []*repositoryRevision, repoResolvers []*searchResultResolver, filter func(repo string) bool) ([]*repositoryRevision, []*searchResultResolver) {
	if repoRevisions == nil || repoResolvers == nil {
		return nil, nil
	}

	filteredRepoRevisions := repoRevisions[:0]
	filteredRepoResolvers := repoResolvers[:0]
	for i, repo := range repoRevisions {
		key := repo.Repo
		if filter(key) {
			filteredRepoRevisions = append(filteredRepoRevisions, repo)
			filteredRepoResolvers = append(filteredRepoResolvers, repoResolvers[i])
		}
	}
	return filteredRepoRevisions, filteredRepoResolvers
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
