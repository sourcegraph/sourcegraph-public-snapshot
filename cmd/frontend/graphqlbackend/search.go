package graphqlbackend

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/felixfbecker/stringscore"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search/query"
	searchquerytypes "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search/query/types"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/errcode"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/inventory/filelang"
	"github.com/sourcegraph/sourcegraph/pkg/pathmatch"
	"github.com/sourcegraph/sourcegraph/pkg/trace"
	"github.com/sourcegraph/sourcegraph/pkg/vcs"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
	"github.com/sourcegraph/sourcegraph/schema"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

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
}) (*searchResolver, error) {
	query, err := query.ParseAndCheck(args.Query)
	if err != nil {
		return nil, err
	}
	return &searchResolver{
		root:  r,
		query: query,
	}, nil
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
	root *schemaResolver

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
	merged, err := viewerMergedConfiguration(ctx)
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
			repos[i] = &types.Repo{URI: api.RepoURI(repoPath)}
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
		sampleRepoPaths := []api.RepoURI{
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

// given a URI, determine whether it matched any patterns for which we have
// revspecs (or ref globs), and if so, return the matching/allowed ones.
func getRevsForMatchedRepo(repo api.RepoURI, pats []patternRevspec) (matched []search.RevisionSpecifier, clashing []search.RevisionSpecifier) {
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
		repoPattern = api.RepoURI(optimizeRepoPatternWithHeuristics(string(repoPattern)))
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
				patterns = append(patterns, "^"+regexp.QuoteMeta(string(repo.URI))+"$")
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
		repoRev := &search.RepositoryRevisions{
			Repo:          repo,
			GitserverRepo: gitserver.Repo{Name: repo.URI},
		}

		revs, clashingRevs := getRevsForMatchedRepo(repo.URI, includePatternRevs)

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
				if _, err := git.ResolveRevision(ctx, repoRev.GitserverRepo, nil, rev.RevSpec, &git.ResolveRevisionOptions{NoEnsureRevision: true}); git.IsRevisionNotFound(err) || err == context.DeadlineExceeded {
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
	repoRevisions, _, _, overLimit, err := r.resolveRepositories(ctx, nil)
	if err != nil {
		return nil, err
	}

	if overLimit {
		// If we've exceeded the repo limit, then we may miss files from repos we care
		// about, so don't bother searching filenames at all.
		return nil, nil
	}

	includePatterns, excludePatterns := r.query.RegexpPatterns(query.FieldFile)
	excludePattern := unionRegExps(excludePatterns)
	pathOptions := pathmatch.CompileOptions{
		RegExp:        true,
		CaseSensitive: r.query.IsCaseSensitive(),
	}

	// Treat all default terms as though they had `file:` before them (to make it easy for users to
	// jump to files by just typing their name).
	for _, v := range r.query.Values(query.FieldDefault) {
		includePatterns = append(includePatterns, asString(v))
	}

	matchPath, err := pathmatch.CompilePathPatterns(includePatterns, excludePattern, pathOptions)
	if err != nil {
		return nil, &badRequestError{err}
	}

	matcher := matcher{match: matchPath.MatchPath}

	// Rank matches if include patterns are specified.
	if len(includePatterns) > 0 {
		scorerQueryParts := make([]string, len(includePatterns))
		for i, includePattern := range includePatterns {
			// Try to extract the text-only (non-regexp) part of the query to
			// pass to stringscore, which doesn't use regexps. This is best-effort.
			scorerQueryParts[i] = strings.TrimSuffix(strings.TrimPrefix(strings.Replace(includePattern, `\`, "", -1), "^"), "$")
		}
		matcher.scorerQuery = strings.Join(scorerQueryParts, " ")
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

// A matcher describes how to filter and score results (for repos and files).
// Exactly one of (query) and (match, scoreQuery) must be set.
type matcher struct {
	query string // query to match using stringscore algorithm

	match       func(path string) bool // func that returns true if the item matches
	scorerQuery string                 // effective query to use in stringscore algorithm
}

// searchTree searches the specified repositories for files and dirs whose name matches the matcher.
func searchTree(ctx context.Context, matcher matcher, repos []*search.RepositoryRevisions, limit int) ([]*searchSuggestionResolver, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		resMu sync.Mutex
		res   []*searchSuggestionResolver
	)
	done := make(chan error, len(repos))
	for _, repoRev := range repos {
		if len(repoRev.Revs) >= 2 {
			return nil, errMultipleRevsNotSupported
		}

		go func(repoRev search.RepositoryRevisions) {
			fileResults, err := searchTreeForRepo(ctx, matcher, repoRev, limit, true)
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
			if errors.Cause(err) != context.Canceled {
				log15.Warn("searchFiles error", "err", err)
			}
		}
	}
	return res, nil
}

var mockSearchFilesForRepo func(matcher matcher, repoRevs search.RepositoryRevisions, limit int, includeDirs bool) ([]*searchSuggestionResolver, error)

// searchTreeForRepo searches the specified repository for files whose name matches
// the matcher
func searchTreeForRepo(ctx context.Context, matcher matcher, repoRevs search.RepositoryRevisions, limit int, includeDirs bool) (res []*searchSuggestionResolver, err error) {
	if mockSearchFilesForRepo != nil {
		return mockSearchFilesForRepo(matcher, repoRevs, limit, includeDirs)
	}

	if len(repoRevs.Revs) == 0 {
		return nil, nil // no revs to search
	}

	repoResolver := &repositoryResolver{repo: repoRevs.Repo}
	commitResolver, err := repoResolver.Commit(ctx, &repositoryCommitArgs{Rev: repoRevs.RevSpecs()[0]}) // TODO(sqs): search all revspecs
	if err != nil {
		return nil, err
	}
	if vcs.IsCloneInProgress(err) {
		// TODO report a cloning repo
		return res, nil
	}
	if commitResolver == nil {
		// TODO(sqs): this means the repository is empty or the revision did not resolve - in either case,
		// there no tree entries here, but maybe we should handle this better
		return nil, nil
	}
	treeResolver, err := commitResolver.Tree(ctx, &struct {
		Path      string
		Recursive bool
	}{Path: ""})
	if err != nil {
		return nil, err
	}
	entries, err := treeResolver.Entries(ctx, &gitTreeEntryConnectionArgs{
		ConnectionArgs: graphqlutil.ConnectionArgs{First: nil},
		Recursive:      true,
	})
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
	for _, entryResolver := range entries {
		if !includeDirs {
			if entryResolver.IsDirectory() {
				continue
			}
		}

		score := scorer.calcScore(entryResolver)
		if score <= 0 && matcher.scorerQuery != "" && matcher.match(entryResolver.path) {
			score = 1 // minimum to ensure everything included by match.match is included
		}
		if score > 0 {
			res = append(res, newSearchResultResolver(entryResolver, score))
		}
	}

	sortSearchSuggestions(res)
	if len(res) > limit {
		res = res[:limit]
	}

	return res, nil
}

// newSearchResultResolver returns a new searchResultResolver wrapping the
// given result.
//
// A panic occurs if the type of result is not a *repositoryResolver or
// *gitTreeEntryResolver.
func newSearchResultResolver(result interface{}, score int) *searchSuggestionResolver {
	switch r := result.(type) {
	case *repositoryResolver:
		return &searchSuggestionResolver{result: r, score: score, length: len(r.repo.URI), label: string(r.repo.URI)}

	case *gitTreeEntryResolver:
		return &searchSuggestionResolver{result: r, score: score, length: len(r.path), label: r.path}

	case *symbolResolver:
		return &searchSuggestionResolver{result: r, score: score, length: len(r.symbol.Name + " " + r.symbol.ContainerName), label: r.symbol.Name + " " + r.symbol.ContainerName}

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
// A panic occurs if the type of result is not a valid search result resolver type.
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
			score = postfixFuzzyAlignScore(splitNoEmpty(string(r.repo.URI), "/"), s.queryParts)
		}
		// Push forks down
		if r.repo.Fork {
			score += scoreBumpFork
		}
		if score > 0 {
			score += scoreBumpRepo
		}
		return score

	case *gitTreeEntryResolver:
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
