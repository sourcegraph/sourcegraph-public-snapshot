package repos

import (
	"context"
	"regexp"
	regexpsyntax "regexp/syntax"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/inconshreveable/log15"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/searchcontexts"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/vcs"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
	"github.com/sourcegraph/sourcegraph/schema"
)

type Resolver struct {
	DB                  dbutil.DB
	SearchableReposFunc searchableReposFunc
}

func (r *Resolver) Resolve(ctx context.Context, op search.RepoOptions) (*search.Repos, error) {
	var err error
	tr, ctx := trace.New(ctx, "resolveRepositories", op.String())
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	includePatterns := op.RepoFilters
	if includePatterns != nil {
		// Copy to avoid race condition.
		includePatterns = append([]string{}, includePatterns...)
	}

	excludePatterns := op.MinusRepoFilters

	limit := op.Limit
	if limit == 0 {
		limit = search.SearchLimits(conf.Get()).MaxRepos
	}

	// If any repo groups are specified, take the intersection of the repo
	// groups and the set of repos specified with repo:. (If none are specified
	// with repo:, then include all from the group.)
	if groupNames := op.RepoGroupFilters; len(groupNames) > 0 {
		groups, err := ResolveRepoGroups(ctx, op.UserSettings)
		if err != nil {
			return nil, err
		}

		unionedPatterns, numPatterns := RepoGroupsToIncludePatterns(groupNames, groups)
		includePatterns = append(includePatterns, unionedPatterns)

		tr.LazyPrintf("repogroups: adding %d repos to include pattern", numPatterns)

		// Ensure we don't omit any repos explicitly included via a repo group. (Each explicitly
		// listed repo generates at least one pattern.)
		if numPatterns > limit {
			limit = numPatterns
		}
	}

	// note that this mutates the strings in includePatterns, stripping their
	// revision specs, if they had any.
	includePatternRevs, err := findPatternRevs(includePatterns)
	if err != nil {
		return nil, err
	}

	// If a version context is specified, gather the list of repository names
	// to limit the results to these repositories.
	var versionContextRepositories []string
	var versionContext *schema.VersionContext
	// If a ref is specified we skip using version contexts.
	if len(includePatternRevs) == 0 && op.VersionContextName != "" {
		versionContext, err = resolveVersionContext(op.VersionContextName)
		if err != nil {
			return nil, err
		}

		for _, revision := range versionContext.Revisions {
			versionContextRepositories = append(versionContextRepositories, revision.Repo)
		}
	}

	searchContext, err := searchcontexts.ResolveSearchContextSpec(ctx, r.DB, op.SearchContextSpec)
	if err != nil {
		return nil, err
	}

	res := search.NewRepos()

	if envvar.SourcegraphDotComMode() && len(includePatterns) == 0 && !query.HasTypeRepo(op.Query) && searchcontexts.IsGlobalSearchContext(searchContext) {
		start := time.Now()
		res.Public, res.Private, err = r.SearchableReposFunc(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "getting list of indexable repos")
		}

		// Remove excluded repos.
		if len(excludePatterns) > 0 {
			patterns, _ := regexp.Compile(`(?i)` + UnionRegExps(excludePatterns))
			exclude := func(r *types.RepoName) {
				if matched := patterns.MatchString(string(r.Name)); matched {
					res.Excluded.Add(r)
				}
			}

			for _, repo := range res.Private.Repos {
				exclude(repo)
			}

			for _, repo := range res.Public.Repos {
				exclude(repo)
			}

			tr.LazyPrintf("construct excluded set - done")
		}

		tr.LazyPrintf("searchableRepos: took %s to add %d repos", time.Since(start), res.Len())

		// Search all indexable repos since indexed search is fast.
		if res.Len() > limit {
			limit = res.Len()
		}
	} else {
		tr.LazyPrintf("Repos.List - start")

		options := database.ReposListOptions{
			IncludePatterns: includePatterns,
			Names:           versionContextRepositories,
			ExcludePattern:  UnionRegExps(excludePatterns),
			// List N+1 repos so we can see if there are repos omitted due to our repo limit.
			LimitOffset:  &database.LimitOffset{Limit: limit + 1},
			NoForks:      op.NoForks,
			OnlyForks:    op.OnlyForks,
			NoArchived:   op.NoArchived,
			OnlyArchived: op.OnlyArchived,
			NoPrivate:    op.Visibility == query.Public,
			OnlyPrivate:  op.Visibility == query.Private,
		}

		if searchContext.ID != 0 {
			options.SearchContextID = searchContext.ID
		} else if searchContext.NamespaceUserID != 0 {
			options.UserID = searchContext.NamespaceUserID
			options.IncludeUserPublicRepos = true
		}

		if op.Ranked {
			options.OrderBy = database.RepoListOrderBy{
				{
					Field:      database.RepoListStars,
					Descending: true,
					Nulls:      "LAST",
				},
			}
		}

		// PERF: We Query concurrently since Count and List call can be slow
		// on Sourcegraph.com (100ms+).
		excludedC := make(chan search.ExcludedRepos)
		go func() {
			excludedC <- computeExcludedRepositories(ctx, r.DB, op.Query, options)
		}()

		err = database.Repos(r.DB).StreamingListRepoNames(ctx, options, func(r *types.RepoName) {
			if r.Private {
				res.Private.Add(r)
			} else {
				res.Public.Add(r)
			}
		})
		tr.LazyPrintf("Repos.List - done")

		excluded := <-excludedC

		res.Excluded.Forks = excluded.Forks
		res.Excluded.Archived = excluded.Archived

		tr.LazyPrintf("excluded repos: %+v", res.Excluded)

		if err != nil {
			return nil, err
		}
	}

	res.OverLimit = res.Len() > limit

	tr.LazyPrintf("Associate/validate revs - start")

	repoRevs := make(map[api.RepoName]search.RevSpecs)

	// For auto-defined search contexts we only search the main branch
	if !searchcontexts.IsAutoDefinedSearchContext(searchContext) {
		searchContextRepositoryRevisions, err := searchcontexts.GetRepositoryRevisions(ctx, r.DB, searchContext.ID)
		if err != nil {
			return nil, err
		}
		for name, revs := range searchContextRepositoryRevisions {
			repoRevs[name] = append(repoRevs[name], revs...)
		}
	}

	if versionContext != nil {
		for _, vcRev := range versionContext.Revisions {
			rev := search.RevisionSpecifier{RevSpec: vcRev.Rev}
			name := api.RepoName(vcRev.Repo)
			repoRevs[name] = append(repoRevs[name], rev)
		}
	}

	for _, p := range includePatternRevs {
		res.ForEach(func(r *types.RepoName, _ search.RevSpecs) error {
			if p.includePattern.MatchString(string(r.Name)) {
				repoRevs[r.Name] = append(repoRevs[r.Name], p.revs...)
			}
			return nil
		})
	}

	// Filter out invalid revisions that the user specified and expand any globs.
	// TODO(tsenart): This could use some concurrency.
	for repo, revs := range repoRevs {
		valid := revs[:0]
		for _, rev := range revs {
			var globs []git.RefGlob
			switch {
			case rev.RefGlob != "":
				globs = append(globs, git.RefGlob{Include: rev.RefGlob})
			case rev.ExcludeRefGlob != "":
				globs = append(globs, git.RefGlob{Exclude: rev.ExcludeRefGlob})
			}

			if len(globs) > 0 {
				refs, err := git.ExpandRefGlobs(ctx, repo, globs)
				if err != nil {
					return nil, err
				}

				if len(refs) == 0 {
					res.MissingRepoRevs[repo] = append(res.MissingRepoRevs[repo], rev)
					continue
				}

				for _, ref := range refs {
					revSpec := search.RevisionSpecifier{RevSpec: strings.TrimPrefix(ref.Name, "refs/heads/")}
					valid = append(valid, revSpec)
				}

				continue
			}

			// not a glob, fast path
			if rev.RevSpec == "" || rev.RevSpec == "HEAD" {
				valid = append(valid, rev)
				// skip default branch resolution to save time
				continue
			}

			// Validate the revspec.
			trimmedRefSpec := strings.TrimPrefix(rev.RevSpec, "^") // handle negated revisions, such as ^<branch>, ^<tag>, or ^<commit>
			if _, err := git.ResolveRevision(ctx, repo, trimmedRefSpec, git.ResolveRevisionOptions{NoEnsureRevision: true}); err != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					return nil, context.DeadlineExceeded
				}
				if errors.HasType(err, git.BadCommitError{}) {
					return nil, err
				}
				if errors.HasType(err, &gitserver.RevisionNotFoundError{}) {
					// The revspec does not exist, so don't include it, and report that it's missing.
					if rev.RevSpec == "" {
						// Report as HEAD not "" (empty string) to avoid user confusion.
						rev.RevSpec = "HEAD"
					}
					res.MissingRepoRevs[repo] = append(res.MissingRepoRevs[repo], rev)
				}
				// If err != nil and is not one of the err values checked for above, cloning and other errors will be handled later, so just ignore an error
				// if there is one.
				continue
			}

			valid = append(valid, rev)
		}

		res.Add(res.GetByName(repo), valid...)
	}

	tr.LazyPrintf("Associate/validate revs - done")

	if op.CommitAfter != "" {
		start := time.Now()
		before := res.Excluded.Len()
		if err = filterRepoHasCommitAfter(ctx, res, op.CommitAfter); err != nil {
			return nil, err
		}
		tr.LazyPrintf("repohascommitafter removed %d repos in %s", res.Excluded.Len()-before, time.Since(start))
	}

	return res, nil
}

// ExactlyOneRepo returns whether exactly one repo: literal field is specified and
// delineated by regex anchors ^ and $. This function helps determine whether we
// should return results for a single repo regardless of whether it is a fork or
// archive.
func ExactlyOneRepo(repoFilters []string) bool {
	if len(repoFilters) == 1 {
		filter, _ := search.ParseRepositoryRevisions(repoFilters[0])
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

func UnionRegExps(patterns []string) string {
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

// NOTE: This function is not called if the version context is not used
func resolveVersionContext(versionContext string) (*schema.VersionContext, error) {
	for _, vc := range conf.Get().ExperimentalFeatures.VersionContexts {
		if vc.Name == versionContext {
			return vc, nil
		}
	}

	return nil, errors.New("version context not found")
}

// Cf. golang/go/src/regexp/syntax/parse.go.
const regexpFlags = regexpsyntax.ClassNL | regexpsyntax.PerlX | regexpsyntax.UnicodeGroups

// computeExcludedRepositories returns a list of excluded repositories (Forks or
// archives) based on the search Query.
func computeExcludedRepositories(ctx context.Context, db dbutil.DB, q query.Q, op database.ReposListOptions) (excluded search.ExcludedRepos) {
	if q == nil {
		return search.ExcludedRepos{}
	}

	// PERF: We Query concurrently since each count call can be slow on
	// Sourcegraph.com (100ms+).
	var wg sync.WaitGroup
	var numExcludedForks, numExcludedArchived int

	if q.Fork() == nil && !ExactlyOneRepo(op.IncludePatterns) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// 'fork:...' was not specified and Forks are excluded, find out
			// which repos are excluded.
			selectForks := op
			selectForks.OnlyForks = true
			selectForks.NoForks = false
			var err error
			numExcludedForks, err = database.Repos(db).Count(ctx, selectForks)
			if err != nil {
				log15.Warn("repo count for excluded fork", "err", err)
			}
		}()
	}

	if q.Archived() == nil && !ExactlyOneRepo(op.IncludePatterns) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Archived...: was not specified and archives are excluded,
			// find out which repos are excluded.
			selectArchived := op
			selectArchived.OnlyArchived = true
			selectArchived.NoArchived = false
			var err error
			numExcludedArchived, err = database.Repos(db).Count(ctx, selectArchived)
			if err != nil {
				log15.Warn("repo count for excluded archive", "err", err)
			}
		}()
	}

	wg.Wait()

	return search.ExcludedRepos{Forks: numExcludedForks, Archived: numExcludedArchived}
}

// a patternRevspec maps an include pattern to a list of revisions
// for repos matching that pattern. "map" in this case does not mean
// an actual map, because we want regexp matches, not identity matches.
type patternRevspec struct {
	includePattern *regexp.Regexp
	revs           search.RevSpecs
}

func getRevsForMatchedRepo(repo api.RepoName, pats []patternRevspec) (matched search.RevSpecs, clashing search.RevSpecs) {
	revLists := make([]search.RevSpecs, 0, len(pats))
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
		matched = search.RevSpecs{{RevSpec: ""}}
		return
	}
	// if two repo specs match, and both provided non-empty rev lists,
	// we want their intersection, so we count the number of times we
	// see a revision in the rev lists, and make sure it matches the number
	// of rev lists
	revCounts := make(map[search.RevisionSpecifier]int, len(revLists[0]))

	var aliveCount int
	for i, revList := range revLists {
		aliveCount = 0
		for _, rev := range revList {
			if revCounts[rev] == i {
				aliveCount += 1
			}
			revCounts[rev] += 1
		}
	}

	if aliveCount > 0 {
		matched = make(search.RevSpecs, 0, len(revCounts))
		for rev, seenCount := range revCounts {
			if seenCount == len(revLists) {
				matched = append(matched, rev)
			}
		}
		sort.Slice(matched, func(i, j int) bool { return matched[i].Less(matched[j]) })
		return
	}

	clashing = make(search.RevSpecs, 0, len(revCounts))
	for rev := range revCounts {
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
		if _, err := regexp.Compile(repoPattern); err != nil {
			return nil, &badRequestError{err}
		}
		repoPattern = optimizeRepoPatternWithHeuristics(repoPattern)
		includePatterns[i] = repoPattern
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

type searchableReposFunc func(ctx context.Context) (public, private *types.RepoSet, err error)

func filterRepoHasCommitAfter(ctx context.Context, res *search.Repos, after string) (err error) {
	var mu sync.Mutex
	sem := semaphore.NewWeighted(128)
	g, ctx := errgroup.WithContext(ctx)

	handle := func(repo *types.RepoName, rev *search.RevisionSpecifier) error {
		ok, err := git.HasCommitAfter(ctx, repo.Name, after, rev.RevSpec)
		if err != nil {
			if errors.HasType(err, &gitserver.RevisionNotFoundError{}) || vcs.IsRepoNotExist(err) {
				return nil
			}
			return err
		}

		if ok {
			mu.Lock()
			res.Excluded.Add(repo)
			mu.Unlock()
		}

		return nil
	}

	res.ForEach(func(r *types.RepoName, revs search.RevSpecs) error {
		if err = sem.Acquire(ctx, 1); err != nil {
			return nil
		}

		g.Go(func() error {
			defer sem.Release(1)

			if len(revs) == 0 {
				return handle(r, &search.RevisionSpecifier{RevSpec: "HEAD"})
			}

			for _, rev := range revs {
				if err := handle(r, &rev); err != nil {
					return err
				}
			}

			return nil
		})

		return nil
	})

	return g.Wait()
}

func optimizeRepoPatternWithHeuristics(repoPattern string) string {
	if envvar.SourcegraphDotComMode() && (strings.HasPrefix(repoPattern, "github.com") || strings.HasPrefix(repoPattern, `github\.com`)) {
		repoPattern = "^" + repoPattern
	}
	// Optimization: make the "." in "github.com" a literal dot
	// so that the regexp can be optimized more effectively.
	repoPattern = strings.ReplaceAll(repoPattern, "github.com", `github\.com`)
	return repoPattern
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

// HandleRepoSearchResult handles the limitHit and searchErr returned by a search function,
// returning common as to reflect that new information. If searchErr is a fatal error,
// it returns a non-nil error; otherwise, if searchErr == nil or a non-fatal error, it returns a
// nil error.
func HandleRepoSearchResult(repoID api.RepoID, revs search.RevSpecs, limitHit, timedOut bool, searchErr error) (_ streaming.Stats, fatalErr error) {
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
	} else if errors.HasType(searchErr, &gitserver.RevisionNotFoundError{}) {
		if len(revs) == 0 || len(revs) == 1 && (revs[0].RevSpec == "" || revs[0].RevSpec == "HEAD") {
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
	}
	return streaming.Stats{
		Status:     search.RepoStatusSingleton(repoID, status),
		IsLimitHit: limitHit,
	}, fatalErr
}
