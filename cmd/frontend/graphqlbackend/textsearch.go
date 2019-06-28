package graphqlbackend

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"regexp/syntax"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/pkg/errors"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/google/zoekt"
	zoektquery "github.com/google/zoekt/query"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search/query"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/errcode"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/mutablelimiter"
	searchbackend "github.com/sourcegraph/sourcegraph/pkg/search/backend"
	"github.com/sourcegraph/sourcegraph/pkg/trace"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
	"gopkg.in/inconshreveable/log15.v2"
)

var (
	// A global limiter on number of concurrent searcher searches.
	textSearchLimiter = mutablelimiter.New(32)

	searchHTTPClient = &http.Client{
		// nethttp.Transport will propagate opentracing spans
		Transport: &nethttp.Transport{
			RoundTripper: &http.Transport{
				// Default is 2, but we can send many concurrent requests
				MaxIdleConnsPerHost: 500,
			},
		},
	}
)

// A light wrapper around the search service. We implement the service here so
// that we can unmarshal the result directly into graphql resolvers.

// fileMatchResolver is a resolver for the GraphQL type `FileMatch`
type fileMatchResolver struct {
	JPath        string       `json:"Path"`
	JLineMatches []*lineMatch `json:"LineMatches"`
	JLimitHit    bool         `json:"LimitHit"`
	symbols      []*searchSymbolResult
	uri          string
	repo         *types.Repo
	commitID     api.CommitID
	// inputRev is the Git revspec that the user originally requested to search. It is used to
	// preserve the original revision specifier from the user instead of navigating them to the
	// absolute commit ID when they select a result.
	inputRev *string
}

func (fm *fileMatchResolver) Key() string {
	return fm.uri
}

func (fm *fileMatchResolver) File() *gitTreeEntryResolver {
	// NOTE(sqs): Omits other commit fields to avoid needing to fetch them
	// (which would make it slow). This gitCommitResolver will return empty
	// values for all other fields.
	return &gitTreeEntryResolver{
		commit: &gitCommitResolver{
			repo:     &repositoryResolver{repo: fm.repo},
			oid:      gitObjectID(fm.commitID),
			inputRev: fm.inputRev,
		},
		path: fm.JPath,
		stat: createFileInfo(fm.JPath, false),
	}
}

func (fm *fileMatchResolver) Repository() *repositoryResolver {
	return &repositoryResolver{repo: fm.repo}
}

func (fm *fileMatchResolver) Resource() string {
	return fm.uri
}

func (fm *fileMatchResolver) Symbols() []*symbolResolver {
	symbols := make([]*symbolResolver, len(fm.symbols))
	for i, s := range fm.symbols {
		symbols[i] = toSymbolResolver(s.symbol, s.baseURI, s.lang, s.commit)
	}
	return symbols
}

func (fm *fileMatchResolver) LineMatches() []*lineMatch {
	return fm.JLineMatches
}

func (fm *fileMatchResolver) LimitHit() bool {
	return fm.JLimitHit
}

func (fm *fileMatchResolver) ToRepository() (*repositoryResolver, bool) { return nil, false }
func (fm *fileMatchResolver) ToFileMatch() (*fileMatchResolver, bool)   { return fm, true }
func (fm *fileMatchResolver) ToCommitSearchResult() (*commitSearchResultResolver, bool) {
	return nil, false
}

func (fm *fileMatchResolver) searchResultURIs() (string, string) {
	return string(fm.repo.Name), fm.JPath
}

func (fm *fileMatchResolver) resultCount() int32 {
	if l := len(fm.LineMatches()); l > 0 {
		return int32(l)
	}
	return 1 // 1 to count "empty" results like type:path results
}

// LineMatch is the struct used by vscode to receive search results for a line
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

var mockTextSearch func(ctx context.Context, repo gitserver.Repo, commit api.CommitID, p *search.PatternInfo, fetchTimeout time.Duration) (matches []*fileMatchResolver, limitHit bool, err error)

// textSearch searches repo@commit with p.
// Note: the returned matches do not set fileMatch.uri
func textSearch(ctx context.Context, repo gitserver.Repo, commit api.CommitID, p *search.PatternInfo, fetchTimeout time.Duration) (matches []*fileMatchResolver, limitHit bool, err error) {
	if mockTextSearch != nil {
		return mockTextSearch(ctx, repo, commit, p, fetchTimeout)
	}

	tr, ctx := trace.New(ctx, "searcher.client", fmt.Sprintf("%s@%s", repo.Name, commit))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	// Combine IncludePattern and IncludePatterns.
	//
	// NOTE: This makes it easier to (in the future) remove support for
	// IncludePattern from searcher and only have it consult IncludePatterns.
	// We still need to send IncludePattern (because searcher isn't guaranteed
	// to be upgraded yet).
	var includePatterns []string
	if p.IncludePattern != "" {
		includePatterns = append(includePatterns, p.IncludePattern)
	}
	includePatterns = append(includePatterns, p.IncludePatterns...)

	q := url.Values{
		"Repo":            []string{string(repo.Name)},
		"URL":             []string{repo.URL},
		"Commit":          []string{string(commit)},
		"Pattern":         []string{p.Pattern},
		"ExcludePattern":  []string{p.ExcludePattern},
		"IncludePatterns": includePatterns,
		"IncludePattern":  []string{p.IncludePattern},
		"FetchTimeout":    []string{fetchTimeout.String()},
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

		searcherURL, err := SearcherURLs().Get(consistentHashKey, excludedSearchURLs)
		if err != nil {
			return nil, false, err
		}

		// Fallback to a bad host if nothing is left
		if searcherURL == "" {
			tr.LazyPrintf("failed to find endpoint, trying again without excludes")
			searcherURL, err = SearcherURLs().Get(consistentHashKey, nil)
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

func textSearchURL(ctx context.Context, url string) ([]*fileMatchResolver, bool, error) {
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
		Matches     []*fileMatchResolver
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

var mockSearchFilesInRepo func(ctx context.Context, repo *types.Repo, gitserverRepo gitserver.Repo, rev string, info *search.PatternInfo, fetchTimeout time.Duration) (matches []*fileMatchResolver, limitHit bool, err error)

func searchFilesInRepo(ctx context.Context, repo *types.Repo, gitserverRepo gitserver.Repo, rev string, info *search.PatternInfo, fetchTimeout time.Duration) (matches []*fileMatchResolver, limitHit bool, err error) {
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

	shouldBeSearched, err := repoShouldBeSearched(ctx, info, gitserverRepo, commit, fetchTimeout)
	if err != nil {
		return nil, false, err
	}
	if !shouldBeSearched {
		return matches, false, err
	}

	matches, limitHit, err = textSearch(ctx, gitserverRepo, commit, info, fetchTimeout)

	workspace := fileMatchURI(repo.Name, rev, "")
	for _, fm := range matches {
		fm.uri = workspace + fm.JPath
		fm.repo = repo
		fm.commitID = commit
		fm.inputRev = &rev
	}

	return matches, limitHit, err
}

// repoShouldBeSearched determines whether a repository should be searched in, based on whether the repository
// fits in the subset of repositories specified in the query's `repohasfile` and `-repohasfile` flags if they exist.
func repoShouldBeSearched(ctx context.Context, searchPattern *search.PatternInfo, gitserverRepo gitserver.Repo, commit api.CommitID, fetchTimeout time.Duration) (shouldBeSearched bool, err error) {
	shouldBeSearched = true
	flagInQuery := len(searchPattern.FilePatternsReposMustInclude) > 0
	if flagInQuery {
		shouldBeSearched, err = repoHasFilesWithNamesMatching(ctx, true, searchPattern.FilePatternsReposMustInclude, gitserverRepo, commit, fetchTimeout)
		if err != nil {
			return shouldBeSearched, err
		}
	}
	negFlagInQuery := len(searchPattern.FilePatternsReposMustExclude) > 0
	if negFlagInQuery {
		shouldBeSearched, err = repoHasFilesWithNamesMatching(ctx, false, searchPattern.FilePatternsReposMustExclude, gitserverRepo, commit, fetchTimeout)
		if err != nil {
			return shouldBeSearched, err
		}
	}
	return shouldBeSearched, nil
}

// repoHasFilesWithNamesMatching searches in a repository for matches for the patterns in the `repohasfile` or `-repohasfile` flags, and returns
// whether or not the repoShouldBeSearched in or not, based on whether matches were returned.
func repoHasFilesWithNamesMatching(ctx context.Context, include bool, repoHasFileFlag []string, gitserverRepo gitserver.Repo, commit api.CommitID, fetchTimeout time.Duration) (bool, error) {
	for _, pattern := range repoHasFileFlag {
		p := search.PatternInfo{IsRegExp: true, FileMatchLimit: 1, IncludePatterns: []string{pattern}, PathPatternsAreRegExps: true, PathPatternsAreCaseSensitive: false, PatternMatchesContent: true, PatternMatchesPath: true}
		matches, _, err := textSearch(ctx, gitserverRepo, commit, &p, fetchTimeout)
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

func zoektResultCountFactor(numRepos int, query *search.PatternInfo) int {
	// If we're only searching a small number of repositories, return more comprehensive results. This is
	// arbitrary.
	k := 1
	switch {
	case numRepos <= 500:
		k = 2
	case numRepos <= 100:
		k = 3
	case numRepos <= 50:
		k = 5
	case numRepos <= 25:
		k = 8
	case numRepos <= 10:
		k = 10
	case numRepos <= 5:
		k = 100
	}
	if query.FileMatchLimit > defaultMaxSearchResults {
		k = int(float64(k) * 3 * float64(query.FileMatchLimit) / float64(defaultMaxSearchResults))
	}
	return k
}

func zoektSearchOpts(k int, query *search.PatternInfo) zoekt.SearchOptions {
	searchOpts := zoekt.SearchOptions{
		MaxWallTime:            1500 * time.Millisecond,
		ShardMaxMatchCount:     100 * k,
		TotalMaxMatchCount:     100 * k,
		ShardMaxImportantMatch: 15 * k,
		TotalMaxImportantMatch: 25 * k,
		MaxDocDisplayCount:     2 * defaultMaxSearchResults,
	}

	// We want zoekt to return more than FileMatchLimit results since we use
	// the extra results to populate reposLimitHit. Additionally the defaults
	// are very low, so we always want to return at least 2000.
	if query.FileMatchLimit > defaultMaxSearchResults {
		searchOpts.MaxDocDisplayCount = 2 * int(query.FileMatchLimit)
	}
	if searchOpts.MaxDocDisplayCount < 2000 {
		searchOpts.MaxDocDisplayCount = 2000
	}

	if userProbablyWantsToWaitLonger := query.FileMatchLimit > defaultMaxSearchResults; userProbablyWantsToWaitLonger {
		searchOpts.MaxWallTime *= time.Duration(3 * float64(query.FileMatchLimit) / float64(defaultMaxSearchResults))
	}

	return searchOpts
}

func zoektSearchHEAD(ctx context.Context, query *search.PatternInfo, repos []*search.RepositoryRevisions, indexedCommits map[*search.RepositoryRevisions]string, useFullDeadline bool, searcher zoekt.Searcher, searchOpts zoekt.SearchOptions, since func(t time.Time) time.Duration) (fm []*fileMatchResolver, limitHit bool, reposLimitHit map[string]struct{}, err error) {
	if len(repos) == 0 {
		return nil, false, nil, nil
	}

	// Tell zoekt which repos to search
	repoSet := &zoektquery.RepoSet{Set: make(map[string]bool, len(repos))}
	repoMap := make(map[api.RepoName]*search.RepositoryRevisions, len(repos))
	for _, repoRev := range repos {
		repoSet.Set[string(repoRev.Repo.Name)] = true
		repoMap[api.RepoName(strings.ToLower(string(repoRev.Repo.Name)))] = repoRev
	}

	queryExceptRepos, err := queryToZoektQuery(query)
	if err != nil {
		return nil, false, nil, err
	}
	finalQuery := zoektquery.NewAnd(repoSet, queryExceptRepos)

	tr, ctx := trace.New(ctx, "zoekt.Search", fmt.Sprintf("%d %+v", len(repoSet.Set), finalQuery.String()))
	defer func() {
		tr.SetError(err)
		if len(fm) > 0 {
			tr.LazyPrintf("%d file matches", len(fm))
		}
		tr.Finish()
	}()

	if useFullDeadline {
		// If the user manually specified a timeout, allow zoekt to use all of the remaining timeout.
		deadline, _ := ctx.Deadline()
		searchOpts.MaxWallTime = time.Until(deadline)

		// We don't want our context's deadline to cut off zoekt so that we can get the results
		// found before the deadline.
		//
		// We'll create a new context that gets cancelled if the other context is cancelled for any
		// reason other than the deadline being exceeded. This essentially means the deadline for the new context
		// will be `deadline + time for zoekt to cancel + network latency`.
		cNew, cancel := context.WithCancel(context.Background())
		go func(cOld context.Context) {
			<-cOld.Done()
			// cancel the new context if the old one is done for some reason other than the deadline passing.
			if cOld.Err() != context.DeadlineExceeded {
				cancel()
			}
		}(ctx)
		ctx = cNew
		defer cancel()
	}

	// If the query has a `repohasfile` or `-repohasfile` flag, we want to construct a new reposet based
	// on the values passed in to the flag.
	newRepoSet, err := createNewRepoSetWithRepoHasFileInputs(ctx, query, searcher, *repoSet)
	if err != nil {
		return nil, false, nil, err
	}
	finalQuery = zoektquery.NewAnd(newRepoSet, queryExceptRepos)
	tr.LazyPrintf("after repohasfile filters: nRepos=%d query=%v", len(newRepoSet.Set), finalQuery)

	t0 := time.Now()
	resp, err := searcher.Search(ctx, finalQuery, &searchOpts)
	if err != nil {
		return nil, false, nil, err
	}
	if resp.FileCount == 0 && resp.MatchCount == 0 && since(t0) >= searchOpts.MaxWallTime {
		timeoutToTry := longer(2, searchOpts.MaxWallTime)
		err2 := errors.Errorf("no results found before timeout in index search (try timeout:%v)", timeoutToTry)
		return nil, false, nil, err2
	}
	limitHit = resp.FilesSkipped+resp.ShardsSkipped > 0
	// Repositories that weren't fully evaluated because they hit the Zoekt or Sourcegraph file match limits.
	reposLimitHit = make(map[string]struct{})
	if limitHit {
		// Zoekt either did not evaluate some files in repositories, or ignored some repositories altogether.
		// In this case, we can't be sure that we have exhaustive results for _any_ repository. So, all file
		// matches are from repos with potentially skipped matches.
		for _, file := range resp.Files {
			if _, ok := reposLimitHit[file.Repository]; !ok {
				reposLimitHit[file.Repository] = struct{}{}
			}
		}
	}

	if len(resp.Files) == 0 {
		return nil, false, nil, nil
	}

	k := zoektResultCountFactor(len(repos), query)
	maxLineMatches := 25 + k
	maxLineFragmentMatches := 3 + k
	if len(resp.Files) > int(query.FileMatchLimit) {
		// List of files we cut out from the Zoekt response because they exceed the file match limit on the Sourcegraph end.
		// We use this to get a list of repositories that do not have complete results.
		fileMatchesInSkippedRepos := resp.Files[int(query.FileMatchLimit):]
		resp.Files = resp.Files[:int(query.FileMatchLimit)]

		if !limitHit {
			// Zoekt evaluated all files and repositories, but Zoekt returned more file matches
			// than the limit we set on Sourcegraph, so we cut out more results.

			// Generate a list of repositories that had results cut because they exceeded the file match limit set on Sourcegraph.
			for _, file := range fileMatchesInSkippedRepos {
				if _, ok := reposLimitHit[file.Repository]; !ok {
					reposLimitHit[file.Repository] = struct{}{}
				}
			}
		}

		limitHit = true
	}
	matches := make([]*fileMatchResolver, len(resp.Files))
	for i, file := range resp.Files {
		fileLimitHit := false
		if len(file.LineMatches) > maxLineMatches {
			file.LineMatches = file.LineMatches[:maxLineMatches]
			fileLimitHit = true
			limitHit = true
		}
		lines := make([]*lineMatch, 0, len(file.LineMatches))
		for _, l := range file.LineMatches {
			if !l.FileName {
				if len(l.LineFragments) > maxLineFragmentMatches {
					l.LineFragments = l.LineFragments[:maxLineFragmentMatches]
				}
				offsets := make([][2]int32, len(l.LineFragments))
				for k, m := range l.LineFragments {
					offset := utf8.RuneCount(l.Line[:m.LineOffset])
					length := utf8.RuneCount(l.Line[m.LineOffset : m.LineOffset+m.MatchLength])
					offsets[k] = [2]int32{int32(offset), int32(length)}
				}
				lines = append(lines, &lineMatch{
					JPreview:          string(l.Line),
					JLineNumber:       int32(l.LineNumber - 1),
					JOffsetAndLengths: offsets,
				})
			}
		}
		repoRev := repoMap[api.RepoName(strings.ToLower(string(file.Repository)))]
		matches[i] = &fileMatchResolver{
			JPath:        file.FileName,
			JLineMatches: lines,
			JLimitHit:    fileLimitHit,
			uri:          fileMatchURI(repoRev.Repo.Name, "", file.FileName),
			repo:         repoRev.Repo,
			commitID:     api.CommitID(indexedCommits[repoRev]),
		}
	}

	return matches, limitHit, reposLimitHit, nil
}

// Returns a new repoSet which accounts for the `repohasfile` and `-repohasfile` flags that may have been passed in the query.
func createNewRepoSetWithRepoHasFileInputs(ctx context.Context, query *search.PatternInfo, searcher zoekt.Searcher, repoSet zoektquery.RepoSet) (*zoektquery.RepoSet, error) {
	newRepoSet := repoSet.Set
	flagIsInQuery := len(query.FilePatternsReposMustInclude) > 0
	negatedFlagIsInQuery := len(query.FilePatternsReposMustExclude) > 0

	// Construct queries which search for repos containing the files passed into `repohasfile`
	filesToIncludeQueries, err := queryToZoektFileOnlyQueries(query, query.FilePatternsReposMustInclude)
	if err != nil {
		return nil, err
	}

	newSearchOpts := zoekt.SearchOptions{
		ShardMaxMatchCount: 1,
		TotalMaxMatchCount: math.MaxInt32,
		MaxDocDisplayCount: 0,
	}
	newSearchOpts.SetDefaults()

	if flagIsInQuery {
		// Set newRepoSet to an empty map if the `repohasflag` exists.
		newRepoSet = make(map[string]bool)

		for i, q := range filesToIncludeQueries {
			// Execute a new Zoekt search for each file passed in to a `repohasfile` flag.
			includeResp, err := searcher.Search(ctx, q, &newSearchOpts)
			if err != nil {
				return nil, errors.Wrapf(err, "searching for %v", q.String())
			}

			for repoURL := range includeResp.RepoURLs {
				if i == 0 {
					// For the results from the first file query, add each repo that is in the result set to newRepoSet.
					//
					// Only add repoURLs that exist in the original repoSet, since
					// repoSet is already filtered down to repositories that adhere to
					// fit the `repo` filters in the query.
					if repoSet.Set[repoURL] {
						newRepoSet[repoURL] = true
					}
				} else {
					// Then, for all following file queries, if there are repositories already existing in newRepoSet that do not appear in
					// the result set for the current file query, remove them so that we only include repos that have at least
					// one match for each `repohasfile` value in newRepoSet.
					for existing := range newRepoSet {
						if _, ok := includeResp.RepoURLs[existing]; !ok {
							delete(newRepoSet, existing)
						}
					}
				}
			}
		}
	}

	// Construct queries which search for repos containing the files passed into `-repohasfile`
	filesToExcludeQueries, err := queryToZoektFileOnlyQueries(query, query.FilePatternsReposMustExclude)
	if err != nil {
		return nil, err
	}

	if negatedFlagIsInQuery {
		for _, q := range filesToExcludeQueries {
			excludeResp, err := searcher.Search(ctx, q, &newSearchOpts)
			if err != nil {
				return nil, err
			}
			for repoURL := range excludeResp.RepoURLs {
				// For each repo that had a result in the exclude set, if it exists in the repoSet, set the value to false so we don't search over it.
				if newRepoSet[repoURL] {
					delete(newRepoSet, repoURL)
				}
			}
		}
	}

	return &zoektquery.RepoSet{Set: newRepoSet}, nil
}

func noOpAnyChar(re *syntax.Regexp) {
	if re.Op == syntax.OpAnyChar {
		re.Op = syntax.OpAnyCharNotNL
	}
	for _, s := range re.Sub {
		noOpAnyChar(s)
	}
}

func parseRe(pattern string, filenameOnly bool, queryIsCaseSensitive bool) (zoektquery.Q, error) {
	// these are the flags used by zoekt, which differ to searcher.
	re, err := syntax.Parse(pattern, syntax.ClassNL|syntax.PerlX|syntax.UnicodeGroups)
	if err != nil {
		return nil, err
	}
	noOpAnyChar(re)
	// zoekt decides to use its literal optimization at the query parser
	// level, so we check if our regex can just be a literal.
	if re.Op == syntax.OpLiteral {
		return &zoektquery.Substring{
			Pattern:       string(re.Rune),
			CaseSensitive: queryIsCaseSensitive,

			FileName: filenameOnly,
		}, nil
	}
	return &zoektquery.Regexp{
		Regexp:        re,
		CaseSensitive: queryIsCaseSensitive,

		FileName: filenameOnly,
	}, nil
}

func fileRe(pattern string, queryIsCaseSensitive bool) (zoektquery.Q, error) {
	return parseRe(pattern, true, queryIsCaseSensitive)
}

func queryToZoektQuery(query *search.PatternInfo) (zoektquery.Q, error) {
	var and []zoektquery.Q

	if query.IsRegExp {
		q, err := parseRe(query.Pattern, false, query.IsCaseSensitive)
		if err != nil {
			return nil, err
		}
		and = append(and, q)
	} else {
		and = append(and, &zoektquery.Substring{
			Pattern:       query.Pattern,
			CaseSensitive: query.IsCaseSensitive,

			FileName: true,
			Content:  true,
		})
	}

	// zoekt also uses regular expressions for file paths
	// TODO PathPatternsAreCaseSensitive
	// TODO whitespace in file path patterns?
	if !query.PathPatternsAreRegExps {
		return nil, errors.New("zoekt only supports regex path patterns")
	}
	for _, p := range query.IncludePatterns {
		q, err := fileRe(p, query.IsCaseSensitive)
		if err != nil {
			return nil, err
		}
		and = append(and, q)
	}
	if query.ExcludePattern != "" {
		q, err := fileRe(query.ExcludePattern, query.IsCaseSensitive)
		if err != nil {
			return nil, err
		}
		and = append(and, &zoektquery.Not{Child: q})
	}

	return zoektquery.Simplify(zoektquery.NewAnd(and...)), nil
}

// queryToZoektFileOnlyQueries constructs a list of Zoekt queries that search for a file pattern(s).
// `listOfFilePaths` specifies which field on `query` should be the list of file patterns to look for.
//  A separate zoekt query is created for each file path that should be searched.
func queryToZoektFileOnlyQueries(query *search.PatternInfo, listOfFilePaths []string) ([]zoektquery.Q, error) {
	var zoektQueries []zoektquery.Q
	if !query.PathPatternsAreRegExps {
		return nil, errors.New("zoekt only supports regex path patterns")
	}
	for _, p := range listOfFilePaths {
		q, err := fileRe(p, query.IsCaseSensitive)
		if err != nil {
			return nil, err
		}
		zoektQueries = append(zoektQueries, zoektquery.Simplify(q))
	}

	return zoektQueries, nil
}

type zoektBackend interface {
	ListAll(context.Context) (*zoekt.RepoList, error)
}

// zoektIndexedRepos splits the input repo list into two parts: (1) the
// repositories `indexed` by Zoekt and (2) the repositories that are
// `unindexed`.
//
// Additionally, it returns a mapping of `indexed` repositories to the exact
// Git commit of HEAD that is indexed.
func zoektIndexedRepos(ctx context.Context, z *searchbackend.Zoekt, revs []*search.RepositoryRevisions) (indexed, unindexed []*search.RepositoryRevisions, indexedCommits map[*search.RepositoryRevisions]string, err error) {
	revset := make(map[string]*search.RepositoryRevisions, len(revs))
	indexedCount := 0

	for _, r := range revs {
		revset[string(r.Repo.Name)] = r
		// We search HEAD using zoekt
		if len(r.Revs) > 0 {
			// TODO(sqs): search all revspecs
			if r.Revs[0].RevSpec == "" {
				indexedCount++
			} else {
				unindexed = append(unindexed, r)
			}
		}
	}

	// Return early if we don't need to querying zoekt
	if indexedCount == 0 {
		return nil, unindexed, nil, nil
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	resp, err := z.ListAll(ctx)
	if err != nil {
		return nil, revs, nil, err
	}

	indexed = make([]*search.RepositoryRevisions, 0, len(resp.Repos))
	indexedCommits = make(map[*search.RepositoryRevisions]string)

	for _, r := range resp.Repos {
		rev, ok := revset[r.Repository.Name]
		if !ok {
			continue
		}

		indexed = append(indexed, rev)
		for _, branch := range r.Repository.Branches {
			if branch.Name == "HEAD" {
				indexedCommits[rev] = branch.Version
				break
			}
		}

		delete(revset, r.Repository.Name)
	}

	unindexed = make([]*search.RepositoryRevisions, 0, len(revset))
	for _, r := range revset {
		unindexed = append(unindexed, r)
	}

	return indexed, unindexed, indexedCommits, nil
}

var mockSearchFilesInRepos func(args *search.Args) ([]*fileMatchResolver, *searchResultsCommon, error)

// searchFilesInRepos searches a set of repos for a pattern.
func searchFilesInRepos(ctx context.Context, args *search.Args) (res []*fileMatchResolver, common *searchResultsCommon, err error) {
	if mockSearchFilesInRepos != nil {
		return mockSearchFilesInRepos(args)
	}

	tr, ctx := trace.New(ctx, "searchFilesInRepos", fmt.Sprintf("query: %+v, numRepoRevs: %d", args.Pattern, len(args.Repos)))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	common = &searchResultsCommon{partial: make(map[api.RepoName]struct{})}

	var (
		zoektRepos     []*search.RepositoryRevisions
		searcherRepos  []*search.RepositoryRevisions = args.Repos
		indexedCommits map[*search.RepositoryRevisions]string
	)

	if args.Zoekt.Enabled() {
		var err error
		zoektRepos, searcherRepos, indexedCommits, err = zoektIndexedRepos(ctx, args.Zoekt, args.Repos)
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

	common.repos = make([]*types.Repo, len(args.Repos))
	for i, repo := range args.Repos {
		common.repos[i] = repo.Repo
	}

	if args.Pattern.IsEmpty() {
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
		unflattened       [][]*fileMatchResolver
		flattenedSize     int
		overLimitCanceled bool // canceled because we were over the limit
	)

	// addMatches assumes the caller holds mu.
	addMatches := func(matches []*fileMatchResolver) {
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
			if flattenedSize > int(args.Pattern.FileMatchLimit) {
				tr.LazyPrintf("cancel due to result size: %d > %d", flattenedSize, args.Pattern.FileMatchLimit)
				overLimitCanceled = true
				common.limitHit = true
				cancel()
			}
		}
	}

	wg.Add(1)
	go func() {
		// TODO limitHit, handleRepoSearchResult
		defer wg.Done()
		query := args.Pattern
		k := zoektResultCountFactor(len(zoektRepos), query)
		opts := zoektSearchOpts(k, query)
		matches, limitHit, reposLimitHit, searchErr := zoektSearchHEAD(ctx, query, zoektRepos, indexedCommits, args.UseFullDeadline, args.Zoekt.Client, opts, time.Since)
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
		tr.LogFields(otlog.Object("searchErr", searchErr), otlog.Error(err), otlog.Bool("overLimitCanceled", overLimitCanceled))
		if searchErr != nil && err == nil && !overLimitCanceled {
			err = searchErr
			tr.LazyPrintf("cancel indexed search due to error: %v", err)
			cancel()
		}
		addMatches(matches)
	}()

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
		eps, err := SearcherURLs().Endpoints()
		if err != nil {
			return nil, common, err
		}
		textSearchLimiter.SetLimit(len(eps) * 32)
	}

	for _, repoRev := range searcherRepos {
		if len(repoRev.Revs) == 0 {
			continue
		}
		if len(repoRev.Revs) >= 2 {
			return nil, common, errMultipleRevsNotSupported
		}

		// Only reason acquire can fail is if ctx is cancelled. So we can stop
		// looping through searcherRepos.
		limitCtx, limitDone, acquireErr := textSearchLimiter.Acquire(ctx)
		if acquireErr != nil {
			break
		}

		wg.Add(1)
		go func(ctx context.Context, done context.CancelFunc, repoRev search.RepositoryRevisions) {
			defer wg.Done()
			defer done()

			rev := repoRev.RevSpecs()[0] // TODO(sqs): search multiple revs
			matches, repoLimitHit, searchErr := searchFilesInRepo(ctx, repoRev.Repo, repoRev.GitserverRepo(), rev, args.Pattern, fetchTimeout)
			if searchErr != nil {
				tr.LogFields(otlog.String("repo", string(repoRev.Repo.Name)), otlog.String("searchErr", searchErr.Error()), otlog.Bool("timeout", errcode.IsTimeout(searchErr)), otlog.Bool("temporary", errcode.IsTemporary(searchErr)))
				log15.Warn("searchFilesInRepo failed", "error", searchErr, "repo", repoRev.Repo.Name)
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
			// non-diff search reports timeout through searchErr, so pass false for timedOut
			if fatalErr := handleRepoSearchResult(common, repoRev, repoLimitHit, false, searchErr); fatalErr != nil {
				if ctx.Err() == context.Canceled {
					// Our request has been canceled (either because another one of searcherRepos
					// had a fatal error, or otherwise), so we can just ignore these results. We
					// handle this here, not in handleRepoSearchResult, because different callers of
					// handleRepoSearchResult (for different result types) currently all need to
					// handle cancellations differently.
					return
				}
				err = errors.Wrapf(searchErr, "failed to search %s", repoRev.String())
				tr.LazyPrintf("cancel due to error: %v", err)
				cancel()
			}
			addMatches(matches)
		}(limitCtx, limitDone, *repoRev)
	}

	wg.Wait()
	if err != nil {
		return nil, common, err
	}

	flattened := flattenFileMatches(unflattened, int(args.Pattern.FileMatchLimit))
	return flattened, common, nil
}

func flattenFileMatches(unflattened [][]*fileMatchResolver, fileMatchLimit int) []*fileMatchResolver {
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
	var flattened []*fileMatchResolver
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

type semaphore chan struct{}

// Acquire increments the semaphore. Up to cap(sem) can be acquired
// concurrently. If the context is canceled before acquiring the context
// error is returned.
func (sem semaphore) Acquire(ctx context.Context) error {
	select {
	case sem <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Release decrements the semaphore.
func (sem semaphore) Release() {
	<-sem
}
