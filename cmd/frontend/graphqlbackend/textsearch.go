package graphqlbackend

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/metrics"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/endpoint"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/mutablelimiter"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
	"gopkg.in/inconshreveable/log15.v2"
)

var (
	// A global limiter on number of concurrent searcher searches.
	textSearchLimiter = mutablelimiter.New(32)

	requestCounter = metrics.NewRequestMeter("textsearch", "Total number of requests sent to the textsearch API.")

	searchHTTPClient = &http.Client{
		// nethttp.Transport will propagate opentracing spans
		Transport: &nethttp.Transport{
			RoundTripper: requestCounter.Transport(&http.Transport{
				// Default is 2, but we can send many concurrent requests
				MaxIdleConnsPerHost: 500,
			}, func(u *url.URL) string {
				// TODO(uwedeportivo): remove once codemod has its own client
				if strings.Contains(u.String(), "replacer") {
					return "replace"
				}
				return "search"
			}),
		},
	}
)

// A light wrapper around the search service. We implement the service here so
// that we can unmarshal the result directly into graphql resolvers.

// FileMatchResolver is a resolver for the GraphQL type `FileMatch`
type FileMatchResolver struct {
	JPath        string       `json:"Path"`
	JLineMatches []*lineMatch `json:"LineMatches"`
	JLimitHit    bool         `json:"LimitHit"`
	symbols      []*searchSymbolResult
	uri          string
	Repo         *types.Repo
	CommitID     api.CommitID
	// InputRev is the Git revspec that the user originally requested to search. It is used to
	// preserve the original revision specifier from the user instead of navigating them to the
	// absolute commit ID when they select a result.
	InputRev *string
}

func (fm *FileMatchResolver) Equal(other *FileMatchResolver) bool {
	return reflect.DeepEqual(fm, other)
}

func (fm *FileMatchResolver) Key() string {
	return fm.uri
}

func (fm *FileMatchResolver) File() *GitTreeEntryResolver {
	// NOTE(sqs): Omits other commit fields to avoid needing to fetch them
	// (which would make it slow). This GitCommitResolver will return empty
	// values for all other fields.
	return &GitTreeEntryResolver{
		commit: &GitCommitResolver{
			repo:     &RepositoryResolver{repo: fm.Repo},
			oid:      GitObjectID(fm.CommitID),
			inputRev: fm.InputRev,
		},
		stat: CreateFileInfo(fm.JPath, false),
	}
}

func (fm *FileMatchResolver) Repository() *RepositoryResolver {
	return &RepositoryResolver{repo: fm.Repo}
}

func (fm *FileMatchResolver) RevSpec() *gitRevSpec {
	if fm.InputRev == nil || *fm.InputRev == "" {
		return nil // default branch
	}
	return &gitRevSpec{
		expr: &gitRevSpecExpr{expr: *fm.InputRev, repo: fm.Repository()},
	}
}

func (fm *FileMatchResolver) Resource() string {
	return fm.uri
}

func (fm *FileMatchResolver) Symbols() []*symbolResolver {
	symbols := make([]*symbolResolver, len(fm.symbols))
	for i, s := range fm.symbols {
		symbols[i] = toSymbolResolver(s.symbol, s.baseURI, s.lang, s.commit)
	}
	return symbols
}

func (fm *FileMatchResolver) LineMatches() []*lineMatch {
	return fm.JLineMatches
}

func (fm *FileMatchResolver) LimitHit() bool {
	return fm.JLimitHit
}

func (fm *FileMatchResolver) ToRepository() (*RepositoryResolver, bool) { return nil, false }
func (fm *FileMatchResolver) ToFileMatch() (*FileMatchResolver, bool)   { return fm, true }
func (fm *FileMatchResolver) ToCommitSearchResult() (*commitSearchResultResolver, bool) {
	return nil, false
}

func (r *FileMatchResolver) ToCodemodResult() (*codemodResultResolver, bool) {
	return nil, false
}

func (fm *FileMatchResolver) searchResultURIs() (string, string) {
	return string(fm.Repo.Name), fm.JPath
}

func (fm *FileMatchResolver) resultCount() int32 {
	rc := len(fm.symbols) + len(fm.LineMatches())
	if rc > 0 {
		return int32(rc)
	}
	return 1 // 1 to count "empty" results like type:path results
}

// lineMatch is the struct used by vscode to receive search results for a line
type lineMatch struct {
	JPreview          string     `json:"Preview"`
	JOffsetAndLengths [][2]int32 `json:"OffsetAndLengths"`
	JLineNumber       int32      `json:"LineNumber"`
	JLimitHit         bool       `json:"LimitHit"`
}

func (lm *lineMatch) Preview() string {
	return lm.JPreview
}

func (lm *lineMatch) LineNumber() int32 {
	return lm.JLineNumber
}

func (lm *lineMatch) OffsetAndLengths() [][]int32 {
	r := make([][]int32, len(lm.JOffsetAndLengths))
	for i := range lm.JOffsetAndLengths {
		r[i] = lm.JOffsetAndLengths[i][:]
	}
	return r
}

func (lm *lineMatch) LimitHit() bool {
	return lm.JLimitHit
}

var mockTextSearch func(ctx context.Context, repo gitserver.Repo, commit api.CommitID, p *search.TextPatternInfo, fetchTimeout time.Duration) (matches []*FileMatchResolver, limitHit bool, err error)

// textSearch searches repo@commit with p.
// Note: the returned matches do not set fileMatch.uri
func textSearch(ctx context.Context, searcherURLs *endpoint.Map, repo gitserver.Repo, commit api.CommitID, p *search.TextPatternInfo, fetchTimeout time.Duration) (matches []*FileMatchResolver, limitHit bool, err error) {
	if mockTextSearch != nil {
		return mockTextSearch(ctx, repo, commit, p, fetchTimeout)
	}

	tr, ctx := trace.New(ctx, "searcher.client", fmt.Sprintf("%s@%s", repo.Name, commit))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	q := url.Values{
		"Repo":            []string{string(repo.Name)},
		"URL":             []string{repo.URL},
		"Commit":          []string{string(commit)},
		"Pattern":         []string{p.Pattern},
		"ExcludePattern":  []string{p.ExcludePattern},
		"IncludePatterns": p.IncludePatterns,
		"FetchTimeout":    []string{fetchTimeout.String()},
		"Languages":       p.Languages,
		"CombyRule":       []string{p.CombyRule},
	}
	if deadline, ok := ctx.Deadline(); ok {
		t, err := deadline.MarshalText()
		if err != nil {
			return nil, false, err
		}
		q.Set("Deadline", string(t))
	}
	q.Set("FileMatchLimit", strconv.FormatInt(int64(p.FileMatchLimit), 10))
	if p.IsRegExp {
		q.Set("IsRegExp", "true")
	}
	if p.IsStructuralPat {
		q.Set("IsStructuralPat", "true")
	}
	if p.IsWordMatch {
		q.Set("IsWordMatch", "true")
	}
	if p.IsCaseSensitive {
		q.Set("IsCaseSensitive", "true")
	}
	if p.PathPatternsAreRegExps {
		q.Set("PathPatternsAreRegExps", "true")
	}
	if p.PathPatternsAreCaseSensitive {
		q.Set("PathPatternsAreCaseSensitive", "true")
	}
	// TEMP BACKCOMPAT: always set even if false so that searcher can distinguish new frontends that send
	// these fields from old frontends that do not (and provide a default in the latter case).
	q.Set("PatternMatchesContent", strconv.FormatBool(p.PatternMatchesContent))
	q.Set("PatternMatchesPath", strconv.FormatBool(p.PatternMatchesPath))
	rawQuery := q.Encode()

	// Searcher caches the file contents for repo@commit since it is
	// relatively expensive to fetch from gitserver. So we use consistent
	// hashing to increase cache hits.
	consistentHashKey := string(repo.Name) + "@" + string(commit)
	tr.LazyPrintf("%s", consistentHashKey)

	var (
		// When we retry do not use a host we already tried.
		excludedSearchURLs = map[string]bool{}
		attempt            = 0
		maxAttempts        = 2
	)
	for {
		attempt++

		searcherURL, err := searcherURLs.Get(consistentHashKey, excludedSearchURLs)
		if err != nil {
			return nil, false, err
		}

		// Fallback to a bad host if nothing is left
		if searcherURL == "" {
			tr.LazyPrintf("failed to find endpoint, trying again without excludes")
			searcherURL, err = searcherURLs.Get(consistentHashKey, nil)
			if err != nil {
				return nil, false, err
			}
		}

		url := searcherURL + "?" + rawQuery
		tr.LazyPrintf("attempt %d: %s", attempt, url)
		matches, limitHit, err = textSearchURL(ctx, url)
		// Useful trace for debugging:
		//
		// tr.LazyPrintf("%d matches, limitHit=%v, err=%v, ctx.Err()=%v", len(matches), limitHit, err, ctx.Err())
		if err == nil || errcode.IsTimeout(err) {
			return matches, limitHit, err
		}

		// If we are canceled, return that error.
		if err := ctx.Err(); err != nil {
			return nil, false, err
		}

		// If not temporary or our last attempt then don't try again.
		if !errcode.IsTemporary(err) || attempt == maxAttempts {
			return nil, false, err
		}

		tr.LazyPrintf("transient error %s", err.Error())
		// Retry search on another searcher instance (if possible)
		excludedSearchURLs[searcherURL] = true
	}
}

func textSearchURL(ctx context.Context, url string) ([]*FileMatchResolver, bool, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, false, err
	}
	req = req.WithContext(ctx)

	req, ht := nethttp.TraceRequest(opentracing.GlobalTracer(), req,
		nethttp.OperationName("Searcher Client"),
		nethttp.ClientTrace(false))
	defer ht.Finish()

	// Do not lose the context returned by TraceRequest
	ctx = req.Context()

	resp, err := searchHTTPClient.Do(req)
	if err != nil {
		// If we failed due to cancellation or timeout (with no partial results in the response
		// body), return just that.
		if ctx.Err() != nil {
			err = ctx.Err()
		}
		return nil, false, errors.Wrap(err, "searcher request failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, false, err
		}
		return nil, false, errors.WithStack(&searcherError{StatusCode: resp.StatusCode, Message: string(body)})
	}

	r := struct {
		Matches     []*FileMatchResolver
		LimitHit    bool
		DeadlineHit bool
	}{}
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return nil, false, errors.Wrap(err, "searcher response invalid")
	}
	if r.DeadlineHit {
		err = context.DeadlineExceeded
	}
	return r.Matches, r.LimitHit, err
}

type searcherError struct {
	StatusCode int
	Message    string
}

func (e *searcherError) BadRequest() bool {
	return e.StatusCode == http.StatusBadRequest
}

func (e *searcherError) Temporary() bool {
	return e.StatusCode == http.StatusServiceUnavailable
}

func (e *searcherError) Error() string {
	return e.Message
}

var mockSearchFilesInRepo func(ctx context.Context, repo *types.Repo, gitserverRepo gitserver.Repo, rev string, info *search.TextPatternInfo, fetchTimeout time.Duration) (matches []*FileMatchResolver, limitHit bool, err error)

func searchFilesInRepo(ctx context.Context, searcherURLs *endpoint.Map, repo *types.Repo, gitserverRepo gitserver.Repo, rev string, info *search.TextPatternInfo, fetchTimeout time.Duration) (matches []*FileMatchResolver, limitHit bool, err error) {
	if mockSearchFilesInRepo != nil {
		return mockSearchFilesInRepo(ctx, repo, gitserverRepo, rev, info, fetchTimeout)
	}

	// Do not trigger a repo-updater lookup (e.g.,
	// backend.{GitRepo,Repos.ResolveRev}) because that would slow this operation
	// down by a lot (if we're looping over many repos). This means that it'll fail if a
	// repo is not on gitserver.
	commit, err := git.ResolveRevision(ctx, gitserverRepo, nil, rev, &git.ResolveRevisionOptions{NoEnsureRevision: true})
	if err != nil {
		return nil, false, err
	}

	shouldBeSearched, err := repoShouldBeSearched(ctx, searcherURLs, info, gitserverRepo, commit, fetchTimeout)
	if err != nil {
		return nil, false, err
	}
	if !shouldBeSearched {
		return matches, false, err
	}

	matches, limitHit, err = textSearch(ctx, searcherURLs, gitserverRepo, commit, info, fetchTimeout)

	workspace := fileMatchURI(repo.Name, rev, "")
	for _, fm := range matches {
		fm.uri = workspace + fm.JPath
		fm.Repo = repo
		fm.CommitID = commit
		fm.InputRev = &rev
	}

	return matches, limitHit, err
}

// repoShouldBeSearched determines whether a repository should be searched in, based on whether the repository
// fits in the subset of repositories specified in the query's `repohasfile` and `-repohasfile` flags if they exist.
func repoShouldBeSearched(ctx context.Context, searcherURLs *endpoint.Map, searchPattern *search.TextPatternInfo, gitserverRepo gitserver.Repo, commit api.CommitID, fetchTimeout time.Duration) (shouldBeSearched bool, err error) {
	shouldBeSearched = true
	flagInQuery := len(searchPattern.FilePatternsReposMustInclude) > 0
	if flagInQuery {
		shouldBeSearched, err = repoHasFilesWithNamesMatching(ctx, searcherURLs, true, searchPattern.FilePatternsReposMustInclude, gitserverRepo, commit, fetchTimeout)
		if err != nil {
			return shouldBeSearched, err
		}
	}
	negFlagInQuery := len(searchPattern.FilePatternsReposMustExclude) > 0
	if negFlagInQuery {
		shouldBeSearched, err = repoHasFilesWithNamesMatching(ctx, searcherURLs, false, searchPattern.FilePatternsReposMustExclude, gitserverRepo, commit, fetchTimeout)
		if err != nil {
			return shouldBeSearched, err
		}
	}
	return shouldBeSearched, nil
}

// repoHasFilesWithNamesMatching searches in a repository for matches for the patterns in the `repohasfile` or `-repohasfile` flags, and returns
// whether or not the repoShouldBeSearched in or not, based on whether matches were returned.
func repoHasFilesWithNamesMatching(ctx context.Context, searcherURLs *endpoint.Map, include bool, repoHasFileFlag []string, gitserverRepo gitserver.Repo, commit api.CommitID, fetchTimeout time.Duration) (bool, error) {
	for _, pattern := range repoHasFileFlag {
		p := search.TextPatternInfo{IsRegExp: true, FileMatchLimit: 1, IncludePatterns: []string{pattern}, PathPatternsAreRegExps: true, PathPatternsAreCaseSensitive: false, PatternMatchesContent: true, PatternMatchesPath: true}
		matches, _, err := textSearch(ctx, searcherURLs, gitserverRepo, commit, &p, fetchTimeout)
		if err != nil {
			return false, err
		}
		if include && len(matches) == 0 || !include && len(matches) > 0 {
			// repo shouldn't be searched if it does not have matches for the patterns in `repohasfile`
			// or if it has file matches for the patterns in `-repohasfile`.
			return false, nil
		}
	}

	return true, nil
}

func fileMatchURI(name api.RepoName, ref, path string) string {
	var b strings.Builder
	ref = url.QueryEscape(ref)
	b.Grow(len(name) + len(ref) + len(path) + len("git://?#"))
	b.WriteString("git://")
	b.WriteString(string(name))
	if ref != "" {
		b.WriteByte('?')
		b.WriteString(ref)
	}
	b.WriteByte('#')
	b.WriteString(path)
	return b.String()
}

var mockSearchFilesInRepos func(args *search.TextParameters) ([]*FileMatchResolver, *searchResultsCommon, error)

// searchFilesInRepos searches a set of repos for a pattern.
func searchFilesInRepos(ctx context.Context, args *search.TextParameters) (res []*FileMatchResolver, common *searchResultsCommon, err error) {
	if mockSearchFilesInRepos != nil {
		return mockSearchFilesInRepos(args)
	}

	tr, ctx := trace.New(ctx, "searchFilesInRepos", fmt.Sprintf("query: %s, numRepoRevs: %d", args.PatternInfo.Pattern, len(args.Repos)))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()
	tr.LogFields(
		trace.Stringer("query", &args.Query.Fields),
		trace.Stringer("info", args.PatternInfo),
	)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	common = &searchResultsCommon{partial: make(map[api.RepoName]struct{})}

	var (
		searcherRepos = args.Repos
		zoektRepos    []*search.RepositoryRevisions
	)

	if args.Zoekt.Enabled() {
		zoektRepos, searcherRepos, err = zoektIndexedRepos(ctx, args.Zoekt, args.Repos, nil)
		if err != nil {
			// Don't hard fail if index is not available yet.
			tr.LogFields(otlog.String("indexErr", err.Error()))
			if ctx.Err() == nil {
				log15.Warn("zoektIndexedRepos failed", "error", err)
			}
			common.indexUnavailable = true
			err = nil
		}
	}

	// if there are no indexed repos and this is a structural search
	// query, there will be no results. Raise a friendly alert.
	if len(zoektRepos) == 0 && args.PatternInfo.IsStructuralPat {
		return nil, nil, errors.New("no indexed repositories for structural search")
	}

	common.repos = make([]*types.Repo, len(args.Repos))
	for i, repo := range args.Repos {
		common.repos[i] = repo.Repo
	}

	if args.PatternInfo.IsEmpty() {
		// Empty query isn't an error, but it has no results.
		return nil, common, nil
	}

	// Support index:yes (default), index:only, and index:no in search query.
	index, _ := args.Query.StringValues(query.FieldIndex)
	if len(index) > 0 {
		index := index[len(index)-1]
		switch parseYesNoOnly(index) {
		case Yes, True:
			// default
			if args.Zoekt.Enabled() {
				tr.LazyPrintf("%d indexed repos, %d unindexed repos", len(zoektRepos), len(searcherRepos))
			}
		case Only:
			if !args.Zoekt.Enabled() {
				return nil, common, fmt.Errorf("invalid index:%q (indexed search is not enabled)", index)
			}
			common.missing = make([]*types.Repo, len(searcherRepos))
			for i, r := range searcherRepos {
				common.missing[i] = r.Repo
			}
			tr.LazyPrintf("index:only, ignoring %d unindexed repos", len(searcherRepos))
			searcherRepos = nil
		case No, False:
			tr.LazyPrintf("index:no, bypassing zoekt (using searcher) for %d indexed repos", len(zoektRepos))
			searcherRepos = append(searcherRepos, zoektRepos...)
			zoektRepos = nil
		default:
			return nil, common, fmt.Errorf("invalid index:%q (valid values are: yes, only, no)", index)
		}
	}

	var (
		// TODO: convert wg to an errgroup
		wg                sync.WaitGroup
		mu                sync.Mutex
		searchErr         error
		unflattened       [][]*FileMatchResolver
		flattenedSize     int
		overLimitCanceled bool // canceled because we were over the limit
	)

	// addMatches assumes the caller holds mu.
	addMatches := func(matches []*FileMatchResolver) {
		if len(matches) > 0 {
			common.resultCount += int32(len(matches))
			sort.Slice(matches, func(i, j int) bool {
				a, b := matches[i].uri, matches[j].uri
				return a > b
			})
			unflattened = append(unflattened, matches)
			flattenedSize += len(matches)

			// Stop searching once we have found enough matches. This does
			// lead to potentially unstable result ordering, but is worth
			// it for the performance benefit.
			if flattenedSize > int(args.PatternInfo.FileMatchLimit) {
				tr.LazyPrintf("cancel due to result size: %d > %d", flattenedSize, args.PatternInfo.FileMatchLimit)
				overLimitCanceled = true
				common.limitHit = true
				cancel()
			}
		}
	}

	// callSearcherOverRepos calls searcher on a set of repos.
	// searcherReposFilteredFiles is an optional map of {repo name => file list}
	// that forces the searcher to only include the file list in the
	// search. It is currently only set when Zoekt restricts the file list for structural search.
	callSearcherOverRepos := func(
		searcherRepos []*search.RepositoryRevisions,
		searcherReposFilteredFiles map[string][]string,
	) error {
		var fetchTimeout time.Duration
		if len(searcherRepos) == 1 || args.UseFullDeadline {
			// When searching a single repo or when an explicit timeout was specified, give it the remaining deadline to fetch the archive.
			deadline, ok := ctx.Deadline()
			if ok {
				fetchTimeout = time.Until(deadline)
			} else {
				// In practice, this case should not happen because a deadline should always be set
				// but if it does happen just set a long but finite timeout.
				fetchTimeout = time.Minute
			}
		} else {
			// When searching many repos, don't wait long for any single repo to fetch.
			fetchTimeout = 500 * time.Millisecond
		}

		if len(searcherRepos) > 0 {
			// The number of searcher endpoints can change over time. Inform our
			// limiter of the new limit, which is a multiple of the number of
			// searchers.
			eps, err := args.SearcherURLs.Endpoints()
			if err != nil {
				return err
			}
			textSearchLimiter.SetLimit(len(eps) * 32)
		}

	outer:
		for _, repoAllRevs := range searcherRepos {
			if len(repoAllRevs.Revs) == 0 {
				continue
			}

			revSpecs, err := repoAllRevs.ExpandedRevSpecs(ctx)
			if err != nil {
				return err
			}

			if len(revSpecs) >= 2 && !conf.SearchMultipleRevisionsPerRepository() {
				return errMultipleRevsNotSupported
			}

			for _, rev := range revSpecs {
				// Only reason acquire can fail is if ctx is cancelled. So we can stop
				// looping through searcherRepos.
				limitCtx, limitDone, acquireErr := textSearchLimiter.Acquire(ctx)
				if acquireErr != nil {
					break outer
				}

				// Make a new repoRev for just the operation of searching this revspec.
				repoRev := &search.RepositoryRevisions{Repo: repoAllRevs.Repo, Revs: []search.RevisionSpecifier{{RevSpec: rev}}}

				args := *args
				if args.PatternInfo.IsStructuralPat && searcherReposFilteredFiles != nil {
					// Modify the search query to only run for the filtered files
					if v, ok := searcherReposFilteredFiles[string(repoRev.Repo.Name)]; ok {
						patternCopy := *args.PatternInfo
						args.PatternInfo = &patternCopy
						includePatternsCopy := []string{}
						args.PatternInfo.IncludePatterns = append(includePatternsCopy, v...)
					}
				}

				wg.Add(1)
				go func(ctx context.Context, done context.CancelFunc) {
					defer wg.Done()
					defer done()

					matches, repoLimitHit, err := searchFilesInRepo(ctx, args.SearcherURLs, repoRev.Repo, repoRev.GitserverRepo(), repoRev.RevSpecs()[0], args.PatternInfo, fetchTimeout)
					if err != nil {
						tr.LogFields(otlog.String("repo", string(repoRev.Repo.Name)), otlog.Error(err), otlog.Bool("timeout", errcode.IsTimeout(err)), otlog.Bool("temporary", errcode.IsTemporary(err)))
						log15.Warn("searchFilesInRepo failed", "error", err, "repo", repoRev.Repo.Name)
					}
					mu.Lock()
					defer mu.Unlock()
					if ctx.Err() == nil {
						common.searched = append(common.searched, repoRev.Repo)
					}
					if repoLimitHit {
						// We did not return all results in this repository.
						common.partial[repoRev.Repo.Name] = struct{}{}
					}
					// non-diff search reports timeout through err, so pass false for timedOut
					if fatalErr := handleRepoSearchResult(common, repoRev, repoLimitHit, false, err); fatalErr != nil {
						if ctx.Err() == context.Canceled {
							// Our request has been canceled (either because another one of searcherRepos
							// had a fatal error, or otherwise), so we can just ignore these results. We
							// handle this here, not in handleRepoSearchResult, because different callers of
							// handleRepoSearchResult (for different result types) currently all need to
							// handle cancellations differently.
							return
						}
						if searchErr == nil {
							searchErr = errors.Wrapf(err, "failed to search %s", repoRev.String())
							tr.LazyPrintf("cancel due to error: %v", searchErr)
							cancel()
						}
					}
					addMatches(matches)
				}(limitCtx, limitDone) // ends the Go routine for a call to searcher for a repo
			} // ends the for loop iterating over repo's revs
		} // ends the for loop iterating over repos
		return nil
	} // ends callSearcherOverRepos

	wg.Add(1)
	go func() {
		// TODO limitHit, handleRepoSearchResult
		defer wg.Done()
		var matches []*FileMatchResolver
		var reposLimitHit map[string]struct{}
		var limitHit bool
		var err error
		if !args.PatternInfo.IsStructuralPat {
			matches, limitHit, reposLimitHit, err = zoektSearchHEAD(ctx, args, zoektRepos, false, time.Since)
		} else {
			matches, limitHit, reposLimitHit, err = zoektSearchHEADOnlyFiles(ctx, args, zoektRepos, false, time.Since)
		}
		mu.Lock()
		defer mu.Unlock()
		if ctx.Err() == nil {
			for _, repo := range zoektRepos {
				common.searched = append(common.searched, repo.Repo)
				common.indexed = append(common.indexed, repo.Repo)
			}
			for repo := range reposLimitHit {
				// Repos that aren't included in the result set due to exceeded limits are partially searched
				// for dynamic filter purposes. Note, reposLimitHit may include repos that did not have any results
				// returned in the original result set, because indexed search has `limitHit` for the
				// entire search rather than per repo as in non-indexed search.
				common.partial[api.RepoName(repo)] = struct{}{}
			}
		}
		if limitHit {
			common.limitHit = true
		}
		if err == errNoResultsInTimeout {
			// Effectively, all repositories have timed out.
			for _, repo := range zoektRepos {
				common.timedout = append(common.timedout, repo.Repo)
			}
		}
		tr.LogFields(otlog.Error(err), otlog.Bool("overLimitCanceled", overLimitCanceled))
		if err != nil && err != errNoResultsInTimeout && searchErr == nil && !overLimitCanceled {
			searchErr = err
			tr.LazyPrintf("cancel indexed search due to error: %v", err)
			cancel()
		}

		if args.PatternInfo.IsStructuralPat {
			// A partition of {repo name => file list} that we will build from Zoekt matches
			partition := make(map[string][]string)
			var repos []*search.RepositoryRevisions

			for _, m := range matches {
				name := string(m.Repo.Name)
				partition[name] = append(partition[name], m.JPath)
			}

			// Filter Zoekt repos that didn't contain matches
			for _, repo := range zoektRepos {
				for key := range partition {
					if string(repo.Repo.Name) == key {
						repos = append(repos, repo)
					}
				}
			}

			// For structural search, we run callSearcherOverRepos
			// over the set of repos and files known to contain
			// parts of the pattern as determined by Zoekt.
			// callSearcherOverRepos must acquire the lock, so we
			// must release the lock held by Zoekt at this point.
			// The Zoekt part of the search is done here as far as
			// structural search is concerned, so the lock can be
			// freely released.
			mu.Unlock()
			err := callSearcherOverRepos(repos, partition)
			mu.Lock()
			if err != nil {
				searchErr = err
			}
		} else {
			addMatches(matches)
		}
	}()

	// This guard disables unindexed structural search for now.
	if !args.PatternInfo.IsStructuralPat {
		if err := callSearcherOverRepos(searcherRepos, nil); err != nil {
			mu.Lock()
			searchErr = err
			mu.Unlock()
		}
	}

	wg.Wait()
	if searchErr != nil {
		return nil, common, searchErr
	}

	flattened := flattenFileMatches(unflattened, int(args.PatternInfo.FileMatchLimit))
	return flattened, common, nil
}

func flattenFileMatches(unflattened [][]*FileMatchResolver, fileMatchLimit int) []*FileMatchResolver {
	// Return early so we don't have to worry about empty lists in later
	// calculations.
	if len(unflattened) == 0 {
		return nil
	}

	// We pass in a limit to each repository so we may end up with R*limit
	// results where R is the number of repositories we searched. To ensure we
	// have results from all repositories unflattened contains the results per
	// repo. We then want to create an idempontent order of results, but
	// ensuring every repo has atleast one result.
	sort.Slice(unflattened, func(i, j int) bool {
		a, b := unflattened[i][0].uri, unflattened[j][0].uri
		return a > b
	})
	var flattened []*FileMatchResolver
	initialPortion := fileMatchLimit / len(unflattened)
	for _, matches := range unflattened {
		if initialPortion < len(matches) {
			flattened = append(flattened, matches[:initialPortion]...)
		} else {
			flattened = append(flattened, matches...)
		}
	}
	// We now have at most initialPortion from each repo. We add the rest of the
	// results until we hit our limit.
	for _, matches := range unflattened {
		low := initialPortion
		high := low + (fileMatchLimit - len(flattened))
		if high <= len(matches) {
			flattened = append(flattened, matches[low:high]...)
		} else if low < len(matches) {
			flattened = append(flattened, matches[low:]...)
		}
	}
	// Sort again since we constructed flattened by adding more results at the
	// end.
	sort.Slice(flattened, func(i, j int) bool {
		a, b := flattened[i].uri, flattened[j].uri
		return a > b
	})

	return flattened
}
