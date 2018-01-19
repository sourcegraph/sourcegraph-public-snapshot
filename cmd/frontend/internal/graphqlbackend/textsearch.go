package graphqlbackend

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/trace"

	"github.com/pkg/errors"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	opentracing "github.com/opentracing/opentracing-go"

	"github.com/neelance/parallel"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api/legacyerr"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/endpoint"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/traceutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/zoekt"
)

var (
	// textSearchLimiter limits the number of open TCP connections created by frontend to searcher.
	textSearchLimiter = parallel.NewRun(500)

	searchHTTPClient = &http.Client{
		// nethttp.Transport will propogate opentracing spans
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

// patternInfo is the struct used by vscode pass on search queries. Keep it in sync with
// pkg/searcher/protocol.PatternInfo.
type patternInfo struct {
	Pattern         string
	IsRegExp        bool
	IsWordMatch     bool
	IsCaseSensitive bool
	FileMatchLimit  int32

	// We do not support IsMultiline
	//IsMultiline     bool
	IncludePattern  *string
	IncludePatterns []string
	ExcludePattern  *string

	PathPatternsAreRegExps       bool
	PathPatternsAreCaseSensitive bool

	PatternMatchesContent bool
	PatternMatchesPath    bool
}

func (p *patternInfo) validate() error {
	if p.IsRegExp {
		if _, err := regexp.Compile(p.Pattern); err != nil {
			return err
		}
	}

	if p.PathPatternsAreRegExps {
		if p.IncludePattern != nil {
			if _, err := regexp.Compile(*p.IncludePattern); err != nil {
				return err
			}
		}
		if p.ExcludePattern != nil {
			if _, err := regexp.Compile(*p.ExcludePattern); err != nil {
				return err
			}
		}
		for _, expr := range p.IncludePatterns {
			if _, err := regexp.Compile(expr); err != nil {
				return err
			}
		}
	}

	return nil
}

type fileMatch struct {
	JPath        string       `json:"Path"`
	JLineMatches []*lineMatch `json:"LineMatches"`
	JLimitHit    bool         `json:"LimitHit"`
	uri          string
}

func (fm *fileMatch) Resource() string {
	return fm.uri
}

func (fm *fileMatch) LineMatches() []*lineMatch {
	return fm.JLineMatches
}

func (fm *fileMatch) LimitHit() bool {
	return fm.JLimitHit
}

func fileMatchesToSearchResults(fms []*fileMatch) []*searchResult {
	results := make([]*searchResult, len(fms))
	for i, fm := range fms {
		results[i] = &searchResult{fileMatch: fm}
	}
	return results
}

// LineMatch is the struct used by vscode to receive search results for a line
type lineMatch struct {
	JPreview          string    `json:"Preview"`
	JLineNumber       int32     `json:"LineNumber"`
	JOffsetAndLengths [][]int32 `json:"OffsetAndLengths"`
	JLimitHit         bool      `json:"LimitHit"`
}

func (lm *lineMatch) Preview() string {
	return lm.JPreview
}

func (lm *lineMatch) LineNumber() int32 {
	return lm.JLineNumber
}

func (lm *lineMatch) OffsetAndLengths() [][]int32 {
	return lm.JOffsetAndLengths
}

func (lm *lineMatch) LimitHit() bool {
	return lm.JLimitHit
}

// textSearch searches repo@commit with p.
// Note: the returned matches do not set fileMatch.uri
func textSearch(ctx context.Context, repo, commit string, p *patternInfo) (matches []*fileMatch, limitHit bool, err error) {
	if searcherURLs == nil {
		return nil, false, errors.New("a searcher service has not been configured")
	}

	traceName, ctx := traceutil.TraceName(ctx, "searcher.client")
	tr := trace.New(traceName, fmt.Sprintf("%s@%s", repo, commit))
	defer func() {
		if err != nil {
			tr.LazyPrintf("error: %v", err)
			tr.SetError()
		}
		tr.Finish()
	}()

	// Combine IncludePattern and IncludePatterns.
	//
	// NOTE: This makes it easier to (in the future) remove support for
	// IncludePattern from searcher and only have it consult IncludePatterns.
	// We still need to send IncludePattern (because searcher isn't guaranteed
	// to be upgraded yet).
	var includePatterns []string
	if p.IncludePattern != nil && *p.IncludePattern != "" {
		includePatterns = append(includePatterns, *p.IncludePattern)
	}
	includePatterns = append(includePatterns, p.IncludePatterns...)

	var s string
	if p.IncludePattern == nil {
		p.IncludePattern = &s
	}
	if p.ExcludePattern == nil {
		p.ExcludePattern = &s
	}
	q := url.Values{
		"Repo":            []string{repo},
		"Commit":          []string{commit},
		"Pattern":         []string{p.Pattern},
		"ExcludePattern":  []string{*p.ExcludePattern},
		"IncludePatterns": includePatterns,
		"IncludePattern":  []string{*p.IncludePattern},
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
	searcherURL, err := searcherURLs.Get(repo + "@" + commit)
	if err != nil {
		return nil, false, err
	}
	req, err := http.NewRequest("GET", searcherURL, nil)
	if err != nil {
		return nil, false, err
	}
	req.URL.RawQuery = q.Encode()
	tr.LazyPrintf("%s", req.URL)
	req = req.WithContext(ctx)

	req, ht := nethttp.TraceRequest(opentracing.GlobalTracer(), req,
		nethttp.OperationName("Searcher Client"),
		nethttp.ClientTrace(false))
	defer ht.Finish()

	// Limit number of outstanding searcher requests
	textSearchLimiter.Acquire()
	defer textSearchLimiter.Release()
	resp, err := searchHTTPClient.Do(req)
	if err != nil {
		// If we failed due to cancellation or timeout, rather return that
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
		Matches  []*fileMatch
		LimitHit bool
	}{}
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return nil, false, errors.Wrap(err, "searcher response invalid")
	}
	return r.Matches, r.LimitHit, nil
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

// isBadRequest will check if error or one of its causes is a bad request.
func isBadRequest(err error) bool {
	type badRequester interface {
		BadRequest() bool
	}
	return isErrorPredicate(err, func(err error) bool {
		badrequest, ok := err.(badRequester)
		return ok && badrequest.BadRequest()
	})
}

// isTemporary will check if error or one of its causes is temporary.
func isTemporary(err error) bool {
	type temporaryer interface {
		Temporary() bool
	}
	return isErrorPredicate(err, func(err error) bool {
		e, ok := err.(temporaryer)
		return ok && e.Temporary()
	})
}

// isTimeout will check if error or one of its causes is a timeout.
func isTimeout(err error) bool {
	type timeouter interface {
		Timeout() bool
	}
	return isErrorPredicate(err, func(err error) bool {
		e, ok := err.(timeouter)
		return ok && e.Timeout()
	})
}

// isErrorPredicate returns true if error or one of its causes returns true
// when passed to p.
func isErrorPredicate(err error, p func(err error) bool) bool {
	type causer interface {
		Cause() error
	}

	for err != nil {
		if p(err) {
			return true
		}
		cause, ok := err.(causer)
		if !ok {
			break
		}
		err = cause.Cause()
	}
	return false
}

var mockSearchRepo func(ctx context.Context, repo *sourcegraph.Repo, rev string, info *patternInfo) (matches []*fileMatch, limitHit bool, err error)

func searchRepo(ctx context.Context, repo *sourcegraph.Repo, rev string, info *patternInfo) (matches []*fileMatch, limitHit bool, err error) {
	if mockSearchRepo != nil {
		return mockSearchRepo(ctx, repo, rev, info)
	}

	commit, err := backend.Repos.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{
		Repo: repo.ID,
		Rev:  rev,
	})
	if err != nil {
		return nil, false, err
	}

	// We expect textSearch to be fast
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	matches, limitHit, err = textSearch(ctx, repo.URI, commit.CommitID, info)

	var workspace string
	if rev != "" {
		workspace = "git://" + repo.URI + "?" + rev + "#"
	} else {
		workspace = "git://" + repo.URI + "#"
	}
	for _, fm := range matches {
		fm.uri = workspace + fm.JPath
	}

	return matches, limitHit, err
}

type repoSearchArgs struct {
	query *patternInfo
	repos []*repositoryRevisions
}

// handleRepoSearchResult handles the limitHit and searchErr returned by a call to searcher or
// gitserver, updating common as to reflect that new information. If searchErr is a fatal error,
// it returns a non-nil error; otherwise, if searchErr == nil or a non-fatal error, it returns a
// nil error.
//
// Callers should use it as follows:
//
//  if fatalErr := handleRepoSearchResult(common, repoRev, limitHit, searchErr); fatalErr != nil {
//     err = errors.Wrapf(searchErr, "failed to search %s because foo", ...) // return this error
//     cancel() // cancel any other in-flight operations
//	}
func handleRepoSearchResult(common *searchResultsCommon, repoRev repositoryRevisions, limitHit bool, searchErr error) (fatalErr error) {
	common.limitHit = common.limitHit || limitHit
	if e, ok := searchErr.(vcs.RepoNotExistError); ok {
		if e.CloneInProgress {
			common.cloning = append(common.cloning, repoRev.repo.URI)
		} else {
			common.missing = append(common.missing, repoRev.repo.URI)
		}
	} else if e, ok := searchErr.(legacyerr.Error); ok && e.Code == legacyerr.NotFound {
		common.missing = append(common.missing, repoRev.repo.URI)
	} else if searchErr == vcs.ErrRevisionNotFound && len(repoRev.revs) == 0 {
		// If we didn't specify an input revision, then the repo is empty and can be ignored.
	} else if isTimeout(searchErr) || isTemporary(searchErr) {
		common.timedout = append(common.timedout, repoRev.repo.URI)
	} else if searchErr != nil {
		return searchErr
	}
	return nil
}

func zoektSearchHEAD(ctx context.Context, query *patternInfo, repos []*repositoryRevisions) (fm []*fileMatch, err error) {
	if len(repos) == 0 {
		return nil, nil
	}

	// Convert sourcegraph pattern into zoekt query
	pattern := query.Pattern
	if !query.IsRegExp {
		pattern = regexp.QuoteMeta(pattern)
	}
	pattern = strconv.Quote(pattern)
	q := []string{pattern}

	// zoekt guesses case sensitivity if we don't specify
	if query.IsCaseSensitive {
		q = append(q, "case:yes")
	} else {
		q = append(q, "case:no")
	}

	// zoekt also uses regular expressions for file paths
	// TODO PathPatternsAreCaseSensitive
	// TODO whitespace in file path patterns?
	if !query.PathPatternsAreRegExps {
		return nil, errors.New("zoekt only supports regex path patterns")
	}
	for _, p := range query.IncludePatterns {
		q = append(q, "f:"+p)
	}
	if query.ExcludePattern != nil {
		q = append(q, "-f:"+*query.ExcludePattern)
	}

	// Tell zoekt which repos to search
	var restrict []zoekt.SearchRequestRestriction
	for _, repoRev := range repos {
		// TODO Repo is a substring match, so we can match more
		restrict = append(restrict, zoekt.SearchRequestRestriction{
			Repo:     repoRev.repo.URI,
			Branches: []string{""}, // "" matches all indexed branches
		})
	}

	if len(restrict) == 0 {
		return nil, nil
	}

	req := zoekt.SearchRequest{
		Query:    strings.Join(q, " "),
		Restrict: restrict,
	}

	traceName, ctx := traceutil.TraceName(ctx, "zoekt.Search")
	tr := trace.New(traceName, fmt.Sprintf("%d %+v", len(req.Restrict), req.Query))
	defer func() {
		if err != nil {
			tr.LazyPrintf("error: %v", err)
			tr.SetError()
		}
		if len(fm) > 0 {
			tr.LazyPrintf("%d file matches", len(fm))
		}
		tr.Finish()
	}()

	resp, err := zoektCl.Search(ctx, req)
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, errors.Errorf("zoekt search failed: %s", *resp.Error)
	}

	if len(resp.Files) == 0 {
		return nil, nil
	}

	matches := make([]*fileMatch, len(resp.Files))
	for i, file := range resp.Files {
		lines := make([]*lineMatch, len(file.Lines))
		for j, l := range file.Lines {
			offsets := make([][]int32, len(l.Matches))
			for k, m := range l.Matches {
				offsets[k] = []int32{int32(m.Start), int32(m.End - m.Start)}
			}
			lines[j] = &lineMatch{
				JPreview:          l.Line,
				JLineNumber:       int32(l.LineNumber - 1),
				JOffsetAndLengths: offsets,
			}
		}
		matches[i] = &fileMatch{
			JPath:        file.FileName,
			JLineMatches: lines,
			uri:          fmt.Sprintf("git://%s#%s", file.Repo, file.FileName),
		}
	}
	return matches, nil
}

func zoektIndexedRepos(ctx context.Context, repos []*repositoryRevisions) (indexed, unindexed []*repositoryRevisions, err error) {
	if zoektCache == nil {
		return nil, repos, nil
	}
	for _, repoRev := range repos {
		// We search HEAD using zoekt
		if repoRev.revSpecsOrDefaultBranch()[0] == "" {
			indexed = append(indexed, repoRev)
		} else {
			unindexed = append(unindexed, repoRev)
		}
	}

	// Return early if we don't need to querying zoekt
	if len(indexed) == 0 {
		return indexed, unindexed, nil
	}

	resp, err := zoektCache.ListAll(ctx)
	if err != nil {
		return nil, nil, err
	}
	if resp.Error != nil {
		return nil, nil, errors.Errorf("zoekt list failed: %s", *resp.Error)
	}

	// Everything currently in indexed is at HEAD. Filter out repos which
	// zoekt hasn't indexed yet.
	set := map[string]bool{}
	for _, repo := range resp.Repos {
		set[repo.Name] = true
	}
	head := indexed
	indexed = indexed[:0]
	for _, repoRev := range head {
		if set[repoRev.repo.URI] {
			indexed = append(indexed, repoRev)
		} else {
			unindexed = append(unindexed, repoRev)
		}
	}

	return indexed, unindexed, nil
}

var mockSearchRepos func(args *repoSearchArgs) ([]*searchResult, *searchResultsCommon, error)

// searchRepos searches a set of repos for a pattern.
func searchRepos(ctx context.Context, args *repoSearchArgs) (res []*searchResult, resCommon *searchResultsCommon, err error) {
	if mockSearchRepos != nil {
		return mockSearchRepos(args)
	}

	traceName, ctx := traceutil.TraceName(ctx, "searchRepos")
	tr := trace.New(traceName, fmt.Sprintf("query: %+v, numRepoRevs: %d", args.query, len(args.repos)))
	defer func() {
		if err != nil {
			tr.LazyPrintf("error: %v", err)
			tr.SetError()
		}
		tr.Finish()
	}()

	if err := args.query.validate(); err != nil {
		return nil, nil, &badRequestError{err}
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	zoektRepos, searcherRepos, err := zoektIndexedRepos(ctx, args.repos)
	if err != nil {
		return nil, nil, err
	}

	var (
		wg            sync.WaitGroup
		mu            sync.Mutex
		unflattened   [][]*fileMatch
		flattenedSize int
		common        = &searchResultsCommon{}
	)
	for _, repoRev := range searcherRepos {
		if len(repoRev.revs) >= 2 {
			return nil, nil, errMultipleRevsNotSupported
		}

		wg.Add(1)
		go func(repoRev repositoryRevisions) {
			defer wg.Done()
			rev := repoRev.revSpecsOrDefaultBranch()[0]
			matches, repoLimitHit, searchErr := searchRepo(ctx, repoRev.repo, rev, args.query)
			mu.Lock()
			defer mu.Unlock()
			if fatalErr := handleRepoSearchResult(common, repoRev, repoLimitHit, searchErr); fatalErr != nil {
				if ctx.Err() != nil {
					// Our request has been canceled, we can just ignore
					// searchRepo for this repo. We only check this condition
					// here since handleRepoSearchResult handles deadlines
					// exceeded differently to canceled.
					return
				}
				err = errors.Wrapf(searchErr, "failed to search %s", repoRev.String())
				tr.LazyPrintf("cancel due to error: %v", err)
				cancel()
			}
			if len(matches) > 0 {
				sort.Slice(matches, func(i, j int) bool {
					a, b := matches[i].uri, matches[j].uri
					return a > b
				})
				unflattened = append(unflattened, matches)
				flattenedSize += len(matches)

				// Stop searching once we have found enough matches. This does
				// lead to potentially unstable result ordering, but is worth
				// it for the performance benefit.
				if flattenedSize > int(args.query.FileMatchLimit) {
					tr.LazyPrintf("cancel due to result size: %d > %d", flattenedSize, args.query.FileMatchLimit)
					cancel()
				}
			}
		}(*repoRev)
	}

	wg.Add(1)
	go func() {
		// TODO limitHit, handleRepoSearchResult
		defer wg.Done()
		matches, searchErr := zoektSearchHEAD(ctx, args.query, zoektRepos)
		mu.Lock()
		defer mu.Unlock()
		if searchErr != nil && err == nil {
			err = searchErr
		}
		if len(matches) > 0 {
			sort.Slice(matches, func(i, j int) bool {
				a, b := matches[i].uri, matches[j].uri
				return a > b
			})
			unflattened = append(unflattened, matches)
		}
	}()

	wg.Wait()
	if err != nil {
		return nil, nil, err
	}

	flattened := flattenFileMatches(unflattened, int(args.query.FileMatchLimit))
	return fileMatchesToSearchResults(flattened), common, nil
}

func flattenFileMatches(unflattened [][]*fileMatch, fileMatchLimit int) []*fileMatch {
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
	var flattened []*fileMatch
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

var zoektCl *zoekt.Client
var zoektCache *zoekt.Cache
var searcherURLs *endpoint.Map

func init() {
	searcherURL := env.Get("SEARCHER_URL", "http://searcher:3181", "searcher server URL")
	if searcherURL == "" {
		return
	}
	var err error
	searcherURLs, err = endpoint.New(searcherURL)
	if err != nil {
		panic(fmt.Sprintf("could not connect to searcher %s: %s", searcherURL, err))
	}

	zoektHost := env.Get("ZOEKT_HOST", "", "host:port of the zoekt instance")
	if zoektHost != "" {
		zoektCl = &zoekt.Client{Host: zoektHost}
		zoektCache = &zoekt.Cache{Client: zoektCl}
	}
}
