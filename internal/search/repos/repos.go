package repos

import (
	"context"
	"fmt"
	"regexp"
	regexpsyntax "regexp/syntax"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/cockroachdb/errors"
	"github.com/hashicorp/go-multierror"
	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/searchcontexts"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

type Resolved struct {
	RepoRevs []*search.RepositoryRevisions

	MissingRepoRevs []*search.RepositoryRevisions
	OverLimit       bool

	// Next points to the next page of resolved repository revisions. It will
	// be nil if there are no more pages left.
	Next types.MultiCursor
}

func (r *Resolved) String() string {
	return fmt.Sprintf("Resolved{RepoRevs=%d, MissingRepoRevs=%d, OverLimit=%v}", len(r.RepoRevs), len(r.MissingRepoRevs), r.OverLimit)
}

// A Pager implements paginated repository resolution.
type Pager interface {
	// Paginate calls the given callback with each page of resolved repositories. If the callback
	// returns an error, Paginate will abort and return that error.
	Paginate(context.Context, *search.RepoOptions, func(*Resolved) error) error
}

type Resolver struct {
	DB   database.DB
	Opts search.RepoOptions
}

func (r *Resolver) Paginate(ctx context.Context, op *search.RepoOptions, handle func(*Resolved) error) (err error) {
	tr, ctx := trace.New(ctx, "searchrepos.Resolver.Paginate", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	var opts search.RepoOptions
	if op != nil {
		opts = *op
	} else {
		opts = r.Opts
	}

	if opts.Limit == 0 {
		opts.Limit = 500
	}

	errs := new(multierror.Error)

	for {
		page, err := r.Resolve(ctx, opts)
		if err != nil {
			errs = multierror.Append(errs, err)
			if !errors.Is(err, &MissingRepoRevsError{}) { // Non-fatal errors
				break
			}
		}
		tr.LazyPrintf("resolved %d repos, %d missing", len(page.RepoRevs), len(page.MissingRepoRevs))

		if err = handle(&page); err != nil {
			errs = multierror.Append(errs, err)
			break
		}

		if page.Next == nil {
			break
		}

		opts.Cursors = page.Next
	}

	return errs.ErrorOrNil()
}

func (r *Resolver) Resolve(ctx context.Context, op search.RepoOptions) (Resolved, error) {
	var err error
	tr, ctx := trace.New(ctx, "searchrepos.Resolver.Resolve", op.String())
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

	// note that this mutates the strings in includePatterns, stripping their
	// revision specs, if they had any.
	includePatternRevs, err := findPatternRevs(includePatterns)
	if err != nil {
		return Resolved{}, err
	}

	searchContext, err := searchcontexts.ResolveSearchContextSpec(ctx, r.DB, op.SearchContextSpec)
	if err != nil {
		return Resolved{}, err
	}

	options := database.ReposListOptions{
		IncludePatterns:       includePatterns,
		ExcludePattern:        search.UnionRegExps(excludePatterns),
		CaseSensitivePatterns: op.CaseSensitiveRepoFilters,
		Cursors:               op.Cursors,
		// List N+1 repos so we can see if there are repos omitted due to our repo limit.
		LimitOffset:  &database.LimitOffset{Limit: limit + 1},
		NoForks:      op.NoForks,
		OnlyForks:    op.OnlyForks,
		NoArchived:   op.NoArchived,
		OnlyArchived: op.OnlyArchived,
		NoPrivate:    op.Visibility == query.Public,
		OnlyPrivate:  op.Visibility == query.Private,
		OrderBy: database.RepoListOrderBy{
			{
				Field:      database.RepoListStars,
				Descending: true,
				Nulls:      "LAST",
			},
			{
				Field:      database.RepoListID,
				Descending: true,
			},
		},
	}

	// Filter by search context repository revisions only if this search context doesn't have
	// a query, which replaces the context:foo term at query parsing time.
	if searchContext.Query == "" {
		options.SearchContextID = searchContext.ID
		options.UserID = searchContext.NamespaceUserID
		options.OrgID = searchContext.NamespaceOrgID
		options.IncludeUserPublicRepos = searchContext.ID == 0 && searchContext.NamespaceUserID != 0
	}

	tr.LazyPrintf("Repos.ListMinimalRepos - start")
	repos, err := r.DB.Repos().ListMinimalRepos(ctx, options)
	tr.LazyPrintf("Repos.ListMinimalRepos - done (%d repos, err %v)", len(repos), err)

	if err != nil {
		return Resolved{}, err
	}

	if len(repos) == 0 && len(op.Cursors) == 0 { // Is the first page empty?
		return Resolved{}, ErrNoResolvedRepos
	}

	var next types.MultiCursor
	if len(repos) == limit+1 { // Do we have a next page?
		last := repos[len(repos)-1]
		for _, o := range options.OrderBy {
			c := types.Cursor{Column: string(o.Field)}

			switch c.Column {
			case "stars":
				c.Value = strconv.FormatInt(int64(last.Stars), 10)
			case "id":
				c.Value = strconv.FormatInt(int64(last.ID), 10)
			}

			if o.Descending {
				c.Direction = "prev"
			} else {
				c.Direction = "next"
			}

			next = append(next, &c)
		}
		repos = repos[:len(repos)-1]
	}

	tr.LazyPrintf("Associate/validate revs - start")

	var searchContextRepositoryRevisions map[api.RepoID]*search.RepositoryRevisions
	if !searchcontexts.IsAutoDefinedSearchContext(searchContext) && searchContext.Query == "" {
		scRepoRevs, err := searchcontexts.GetRepositoryRevisions(ctx, r.DB, searchContext.ID)
		if err != nil {
			return Resolved{}, err
		}

		searchContextRepositoryRevisions = make(map[api.RepoID]*search.RepositoryRevisions, len(scRepoRevs))
		for _, repoRev := range scRepoRevs {
			searchContextRepositoryRevisions[repoRev.Repo.ID] = repoRev
		}
	}

	var res struct {
		sync.Mutex
		Resolved
		multierror.Error
	}

	res.Resolved = Resolved{
		RepoRevs: make([]*search.RepositoryRevisions, len(repos)),
		Next:     next,
	}

	sem := semaphore.NewWeighted(128)
	g, ctx := errgroup.WithContext(ctx)

	for i, repo := range repos {
		if err = sem.Acquire(ctx, 1); err != nil {
			return Resolved{}, err
		}

		repo, i := repo, i // avoid race

		g.Go(func() error {
			defer sem.Release(1)

			var (
				repoRev = search.RepositoryRevisions{Repo: repo}
				revs    []search.RevisionSpecifier
			)

			if len(searchContextRepositoryRevisions) > 0 {
				if scRepoRev := searchContextRepositoryRevisions[repo.ID]; scRepoRev != nil {
					revs = scRepoRev.Revs
				}
			} else {
				var clashingRevs []search.RevisionSpecifier
				revs, clashingRevs = getRevsForMatchedRepo(repo.Name, includePatternRevs)

				// if multiple specified revisions clash, report this usefully:
				if len(revs) == 0 && clashingRevs != nil {
					res.Lock()
					res.MissingRepoRevs = append(res.MissingRepoRevs, &search.RepositoryRevisions{
						Repo: repo,
						Revs: clashingRevs,
					})
					res.Unlock()
				}
			}

			// We do in place filtering to reduce allocations. Common path is no
			// filtering of revs.
			if len(revs) > 0 {
				repoRev.Revs = revs[:0]
			}

			// Check if the repository actually has the revisions that the user specified.
			for _, rev := range revs {
				if rev.RefGlob != "" || rev.ExcludeRefGlob != "" {
					// Do not validate ref patterns. A ref pattern matching 0 refs is not necessarily
					// invalid, so it's not clear what validation would even mean.
					repoRev.Revs = append(repoRev.Revs, rev)
					continue
				}

				if rev.RevSpec == "" && op.CommitAfter == "" { // skip default branch resolution to save time
					repoRev.Revs = append(repoRev.Revs, rev)
					continue
				}

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
				if rev.RevSpec == "" {
					rev.RevSpec = "HEAD"
				}

				trimmedRefSpec := strings.TrimPrefix(rev.RevSpec, "^") // handle negated revisions, such as ^<branch>, ^<tag>, or ^<commit>
				commitID, err := git.ResolveRevision(ctx, repoRev.Repo.Name, trimmedRefSpec, git.ResolveRevisionOptions{NoEnsureRevision: true})
				if err != nil {
					if errors.Is(err, context.DeadlineExceeded) || errors.HasType(err, gitdomain.BadCommitError{}) {
						return err
					}

					if errors.HasType(err, &gitdomain.RevisionNotFoundError{}) {
						// The revspec does not exist, so don't include it, and report that it's missing.
						if rev.RevSpec == "" {
							// Report as HEAD not "" (empty string) to avoid user confusion.
							rev.RevSpec = "HEAD"
						}
						res.Lock()
						res.MissingRepoRevs = append(res.MissingRepoRevs, &search.RepositoryRevisions{
							Repo: repo,
							Revs: []search.RevisionSpecifier{{RevSpec: rev.RevSpec}},
						})
						res.Unlock()
					}
					// If err != nil and is not one of the err values checked for above, cloning and other errors will be handled later, so just ignore an error
					// if there is one.
					continue
				}

				if op.CommitAfter != "" {
					n, err := git.CommitCount(ctx, repoRev.Repo.Name, git.CommitsOptions{
						N:     1,
						After: op.CommitAfter,
						Range: string(commitID),
					})

					if err != nil {
						if !errors.HasType(err, &gitdomain.RevisionNotFoundError{}) && !gitdomain.IsRepoNotExist(err) {
							res.Lock()
							multierror.Append(&res.Error, err)
							res.Unlock()
						}
						continue
					}

					if n == 0 {
						continue
					}
				}

				repoRev.Revs = append(repoRev.Revs, rev)
			}

			if len(repoRev.Revs) > 0 {
				res.Lock()
				res.RepoRevs[i] = &repoRev
				res.Unlock()
			}

			return nil
		})
	}

	if err = g.Wait(); err != nil {
		return Resolved{}, err
	}

	// Remove any repos that failed to have their revs validated. We do this to preserve the original order.
	valid := res.RepoRevs[:0]
	for _, r := range res.RepoRevs {
		if r != nil {
			valid = append(valid, r)
		}
	}
	res.RepoRevs = valid

	tr.LazyPrintf("Associate/validate revs - done")

	err = res.ErrorOrNil()
	if len(res.MissingRepoRevs) > 0 {
		err = multierror.Append(err, &MissingRepoRevsError{Missing: res.MissingRepoRevs})
	}

	return res.Resolved, err
}

// Excluded computes the ExcludedRepos that the given RepoOptions would not match. This is
// used to show in the search UI what repos are excluded precisely.
func (r *Resolver) Excluded(ctx context.Context, op search.RepoOptions) (ex ExcludedRepos, err error) {
	tr, ctx := trace.New(ctx, "searchrepos.Resolver.Excluded", op.String())
	defer func() {
		tr.LazyPrintf("excluded repos: %+v", ex)
		tr.SetError(err)
		tr.Finish()
	}()

	if op.Query == nil {
		return ExcludedRepos{}, nil
	}

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

	// note that this mutates the strings in includePatterns, stripping their
	// revision specs, if they had any.
	_, err = findPatternRevs(includePatterns)
	if err != nil {
		return ExcludedRepos{}, err
	}

	searchContext, err := searchcontexts.ResolveSearchContextSpec(ctx, r.DB, op.SearchContextSpec)
	if err != nil {
		return ExcludedRepos{}, err
	}

	options := database.ReposListOptions{
		IncludePatterns: includePatterns,
		ExcludePattern:  search.UnionRegExps(excludePatterns),
		// List N+1 repos so we can see if there are repos omitted due to our repo limit.
		LimitOffset:            &database.LimitOffset{Limit: limit + 1},
		NoForks:                op.NoForks,
		OnlyForks:              op.OnlyForks,
		NoArchived:             op.NoArchived,
		OnlyArchived:           op.OnlyArchived,
		NoPrivate:              op.Visibility == query.Public,
		OnlyPrivate:            op.Visibility == query.Private,
		SearchContextID:        searchContext.ID,
		UserID:                 searchContext.NamespaceUserID,
		OrgID:                  searchContext.NamespaceOrgID,
		IncludeUserPublicRepos: searchContext.ID == 0 && searchContext.NamespaceUserID != 0,
	}

	g, ctx := errgroup.WithContext(ctx)

	var excluded struct {
		sync.Mutex
		ExcludedRepos
	}

	if op.Query.Fork() == nil && !ExactlyOneRepo(includePatterns) {
		g.Go(func() error {
			// 'fork:...' was not specified and Forks are excluded, find out
			// which repos are excluded.
			selectForks := options
			selectForks.OnlyForks = true
			selectForks.NoForks = false
			numExcludedForks, err := r.DB.Repos().Count(ctx, selectForks)
			if err != nil {
				return err
			}

			excluded.Lock()
			excluded.Forks = numExcludedForks
			excluded.Unlock()

			return nil
		})
	}

	if op.Query.Archived() == nil && !ExactlyOneRepo(includePatterns) {
		g.Go(func() error {
			// Archived...: was not specified and archives are excluded,
			// find out which repos are excluded.
			selectArchived := options
			selectArchived.OnlyArchived = true
			selectArchived.NoArchived = false
			numExcludedArchived, err := r.DB.Repos().Count(ctx, selectArchived)
			if err != nil {
				return err
			}

			excluded.Lock()
			excluded.Archived = numExcludedArchived
			excluded.Unlock()

			return nil
		})
	}

	return excluded.ExcludedRepos, g.Wait()
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

// Cf. golang/go/src/regexp/syntax/parse.go.
const regexpFlags = regexpsyntax.ClassNL | regexpsyntax.PerlX | regexpsyntax.UnicodeGroups

// ExcludedRepos is a type that counts how many repos with a certain label were
// excluded from search results.
type ExcludedRepos struct {
	Forks    int
	Archived int
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
		matched = make([]search.RevisionSpecifier, 0, len(revCounts))
		for rev, seenCount := range revCounts {
			if seenCount == len(revLists) {
				matched = append(matched, rev)
			}
		}
		sort.Slice(matched, func(i, j int) bool { return matched[i].Less(matched[j]) })
		return
	}

	clashing = make([]search.RevisionSpecifier, 0, len(revCounts))
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

var ErrNoResolvedRepos = errors.New("no resolved repositories")

type MissingRepoRevsError struct {
	Missing []*search.RepositoryRevisions
}

func (MissingRepoRevsError) Error() string { return "missing repo revs" }

// HandleRepoSearchResult handles the limitHit and searchErr returned by a search function,
// returning common as to reflect that new information. If searchErr is a fatal error,
// it returns a non-nil error; otherwise, if searchErr == nil or a non-fatal error, it returns a
// nil error.
func HandleRepoSearchResult(repoRev *search.RepositoryRevisions, limitHit, timedOut bool, searchErr error) (_ streaming.Stats, fatalErr error) {
	var status search.RepoStatus
	if limitHit {
		status |= search.RepoStatusLimitHit
	}

	if gitdomain.IsRepoNotExist(searchErr) {
		if gitdomain.IsCloneInProgress(searchErr) {
			status |= search.RepoStatusCloning
		} else {
			status |= search.RepoStatusMissing
		}
	} else if errors.HasType(searchErr, &gitdomain.RevisionNotFoundError{}) {
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
	}
	return streaming.Stats{
		Status:     search.RepoStatusSingleton(repoRev.Repo.ID, status),
		IsLimitHit: limitHit,
	}, fatalErr
}

func PrivateReposForUser(ctx context.Context, db database.DB, userID int32, repoOptions search.RepoOptions) []types.MinimalRepo {
	tr, ctx := trace.New(ctx, "privateReposForUser", strconv.Itoa(int(userID)))
	defer tr.Finish()

	// Get all private repos for the the current actor. On sourcegraph.com, those are
	// only the repos directly added by the user. Otherwise it's all repos the user has
	// access to on all connected code hosts / external services.
	//
	// TODO: We should use repos.Resolve here. However, the logic for
	// UserID is different to repos.Resolve, so we need to work out how
	// best to address that first.
	userPrivateRepos, err := db.Repos().ListMinimalRepos(ctx, database.ReposListOptions{
		UserID:         userID, // Zero valued when not in sourcegraph.com mode
		OnlyPrivate:    true,
		LimitOffset:    &database.LimitOffset{Limit: search.SearchLimits(conf.Get()).MaxRepos + 1},
		OnlyForks:      repoOptions.OnlyForks,
		NoForks:        repoOptions.NoForks,
		OnlyArchived:   repoOptions.OnlyArchived,
		NoArchived:     repoOptions.NoArchived,
		ExcludePattern: search.UnionRegExps(repoOptions.MinusRepoFilters),
	})

	if err != nil {
		log15.Error("doResults: failed to list user private repos", "error", err, "user-id", userID)
		tr.LazyPrintf("error resolving user private repos: %v", err)
	}
	return userPrivateRepos
}
