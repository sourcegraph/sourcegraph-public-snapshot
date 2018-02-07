package graphqlbackend

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
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

	"github.com/google/zoekt"
	zoektquery "github.com/google/zoekt/query"
	zoektrpc "github.com/google/zoekt/rpc"
	"github.com/neelance/parallel"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/endpoint"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/searchquery"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/traceutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	zoektpkg "sourcegraph.com/sourcegraph/sourcegraph/pkg/zoekt"
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
	repo         *types.Repo
	commitID     api.CommitID // or empty for default branch
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
func textSearch(ctx context.Context, repo gitserver.Repo, commit api.CommitID, p *patternInfo) (matches []*fileMatch, limitHit bool, err error) {
	if searcherURLs == nil {
		return nil, false, errors.New("a searcher service has not been configured")
	}

	traceName, ctx := traceutil.TraceName(ctx, "searcher.client")
	tr := trace.New(traceName, fmt.Sprintf("%s@%s", repo.Name, commit))
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
		"Repo":            []string{string(repo.Name)},
		"URL":             []string{repo.URL},
		"Commit":          []string{string(commit)},
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
	searcherURL, err := searcherURLs.Get(string(repo.Name) + "@" + string(commit))
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

var mockSearchFilesInRepo func(ctx context.Context, repo *types.Repo, gitserverRepo gitserver.Repo, rev string, info *patternInfo) (matches []*fileMatch, limitHit bool, err error)

func searchFilesInRepo(ctx context.Context, repo *types.Repo, gitserverRepo gitserver.Repo, rev string, info *patternInfo) (matches []*fileMatch, limitHit bool, err error) {
	if mockSearchFilesInRepo != nil {
		return mockSearchFilesInRepo(ctx, repo, gitserverRepo, rev, info)
	}

	commit, err := backend.Repos.VCS(gitserverRepo).ResolveRevision(ctx, rev, nil)
	if err != nil {
		return nil, false, err
	}

	// We expect textSearch to be fast
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	matches, limitHit, err = textSearch(ctx, gitserverRepo, commit, info)

	var workspace string
	if rev != "" {
		workspace = "git://" + string(repo.URI) + "?" + rev + "#"
	} else {
		workspace = "git://" + string(repo.URI) + "#"
	}
	for _, fm := range matches {
		fm.uri = workspace + fm.JPath
		fm.repo = repo
		fm.commitID = commit
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
//  if fatalErr := handleRepoSearchResult(common, repoRev, limitHit, timedOut, searchErr); fatalErr != nil {
//     err = errors.Wrapf(searchErr, "failed to search %s because foo", ...) // return this error
//     cancel() // cancel any other in-flight operations
//	}
func handleRepoSearchResult(common *searchResultsCommon, repoRev repositoryRevisions, limitHit, timedOut bool, searchErr error) (fatalErr error) {
	common.limitHit = common.limitHit || limitHit
	if e, ok := searchErr.(vcs.RepoNotExistError); ok {
		if e.CloneInProgress {
			common.cloning = append(common.cloning, repoRev.repo.URI)
		} else {
			common.missing = append(common.missing, repoRev.repo.URI)
		}
	} else if errcode.IsNotFound(searchErr) {
		common.missing = append(common.missing, repoRev.repo.URI)
	} else if searchErr == vcs.ErrRevisionNotFound && (len(repoRev.revs) == 0 || len(repoRev.revs) == 1 && repoRev.revs[0].revspec == "") {
		// If we didn't specify an input revision, then the repo is empty and can be ignored.
	} else if errcode.IsTimeout(searchErr) || errcode.IsTemporary(searchErr) || timedOut {
		common.timedout = append(common.timedout, repoRev.repo.URI)
	} else if searchErr != nil {
		return searchErr
	}
	return nil
}

func zoektSearchHEAD(ctx context.Context, query *patternInfo, repos []*repositoryRevisions) (fm []*fileMatch, limitHit bool, err error) {
	if len(repos) == 0 {
		return nil, false, nil
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
		return nil, false, errors.New("zoekt only supports regex path patterns")
	}
	for _, p := range query.IncludePatterns {
		q = append(q, "f:"+p)
	}
	if query.ExcludePattern != nil {
		q = append(q, "-f:"+*query.ExcludePattern)
	}

	// Tell zoekt which repos to search
	repoSet := &zoektquery.RepoSet{Set: make(map[string]bool, len(repos))}
	repoMap := make(map[api.RepoURI]*types.Repo, len(repos))
	for _, repoRev := range repos {
		repoSet.Set[string(repoRev.repo.URI)] = true
		repoMap[api.RepoURI(strings.ToLower(string(repoRev.repo.URI)))] = repoRev.repo
	}

	if len(repoSet.Set) == 0 {
		return nil, false, nil
	}

	queryExceptRepos, err := zoektquery.Parse(strings.Join(q, " "))
	if err != nil {
		return nil, false, err
	}
	finalQuery := zoektquery.NewAnd(repoSet, queryExceptRepos)

	traceName, ctx := traceutil.TraceName(ctx, "zoekt.Search")
	tr := trace.New(traceName, fmt.Sprintf("%d %+v", len(repoSet.Set), finalQuery.String()))
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

	// If we're only searching a small number of repositories, return more comprehensive results. This is
	// arbitrary.
	k := 1
	switch {
	case len(repos) <= 500:
		k = 2
	case len(repos) <= 100:
		k = 3
	case len(repos) <= 50:
		k = 5
	case len(repos) <= 25:
		k = 8
	case len(repos) <= 10:
		k = 10
	case len(repos) <= 5:
		k = 100
	}
	if query.FileMatchLimit > defaultMaxSearchResults {
		k = int(float64(k) * 3 * float64(query.FileMatchLimit) / float64(defaultMaxSearchResults))
	}

	searchOpts := zoekt.SearchOptions{
		MaxWallTime:            1500 * time.Millisecond,
		ShardMaxMatchCount:     100 * k,
		TotalMaxMatchCount:     100 * k,
		ShardMaxImportantMatch: 15 * k,
		TotalMaxImportantMatch: 25 * k,
	}

	if userProbablyWantsToWaitLonger := query.FileMatchLimit > defaultMaxSearchResults; userProbablyWantsToWaitLonger {
		searchOpts.MaxWallTime *= time.Duration(3 * float64(query.FileMatchLimit) / float64(defaultMaxSearchResults))
		tr.LazyPrintf("maxwalltime %s", searchOpts.MaxWallTime)
	}

	ctx, cancel := context.WithTimeout(ctx, searchOpts.MaxWallTime+3*time.Second)
	defer cancel()

	resp, err := zoektCl.Search(ctx, finalQuery, &searchOpts)
	if err != nil {
		return nil, false, err
	}
	limitHit = resp.FilesSkipped > 0

	if len(resp.Files) == 0 {
		return nil, false, nil
	}

	maxLineMatches := 25 + k
	maxLineFragmentMatches := 3 + k
	if len(resp.Files) > int(query.FileMatchLimit) {
		resp.Files = resp.Files[:int(query.FileMatchLimit)]
		limitHit = true
	}
	matches := make([]*fileMatch, len(resp.Files))
	for i, file := range resp.Files {
		if len(file.LineMatches) > maxLineMatches {
			file.LineMatches = file.LineMatches[:maxLineMatches]
		}
		lines := make([]*lineMatch, len(file.LineMatches))
		for j, l := range file.LineMatches {
			if len(l.LineFragments) > maxLineFragmentMatches {
				l.LineFragments = l.LineFragments[:maxLineFragmentMatches]
			}
			offsets := make([][]int32, len(l.LineFragments))
			for k, m := range l.LineFragments {
				offsets[k] = []int32{int32(m.LineOffset), int32(m.MatchLength)}
			}
			lines[j] = &lineMatch{
				JPreview:          string(l.Line),
				JLineNumber:       int32(l.LineNumber - 1),
				JOffsetAndLengths: offsets,
			}
		}
		matches[i] = &fileMatch{
			JPath:        file.FileName,
			JLineMatches: lines,
			uri:          fmt.Sprintf("git://%s#%s", file.Repository, file.FileName),
			repo:         repoMap[api.RepoURI(strings.ToLower(string(file.Repository)))],
			commitID:     "", // zoekt only searches default branch
		}
	}
	return matches, limitHit, nil
}

func zoektIndexedRepos(ctx context.Context, repos []*repositoryRevisions) (indexed, unindexed []*repositoryRevisions, err error) {
	if zoektCache == nil {
		return nil, repos, nil
	}
	for _, repoRev := range repos {
		// We search HEAD using zoekt
		if revspecs := repoRev.revspecs(); len(revspecs) > 0 {
			// TODO(sqs): search all revspecs
			if revspecs[0] == "" {
				indexed = append(indexed, repoRev)
			} else {
				unindexed = append(unindexed, repoRev)
			}
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

	// Everything currently in indexed is at HEAD. Filter out repos which
	// zoekt hasn't indexed yet.
	set := map[string]bool{}
	for _, repo := range resp.Repos {
		set[repo.Repository.Name] = true
	}
	head := indexed
	indexed = indexed[:0]
	for _, repoRev := range head {
		if set[string(repoRev.repo.URI)] {
			indexed = append(indexed, repoRev)
		} else {
			unindexed = append(unindexed, repoRev)
		}
	}

	return indexed, unindexed, nil
}

var mockSearchFilesInRepos func(args *repoSearchArgs) ([]*searchResult, *searchResultsCommon, error)

// searchFilesInRepos searches a set of repos for a pattern.
func searchFilesInRepos(ctx context.Context, args *repoSearchArgs, query searchquery.Query) (res []*searchResult, common *searchResultsCommon, err error) {
	if mockSearchFilesInRepos != nil {
		return mockSearchFilesInRepos(args)
	}

	traceName, ctx := traceutil.TraceName(ctx, "searchFilesInRepos")
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

	common = &searchResultsCommon{}

	common.repos = make([]api.RepoURI, len(args.repos))
	for i, repo := range args.repos {
		common.repos[i] = repo.repo.URI
	}

	if args.query.Pattern == "" {
		// Empty query isn't an error, but it has no results.
		return nil, common, nil
	}

	// Support expzoektonly:yes and expsearcheronly:yes in search query.
	index, _ := query.StringValues(searchquery.FieldIndex)
	if len(index) == 0 && os.Getenv("SEARCH10_INDEX_DEFAULT") != "" && len(args.repos) > 10 {
		index = []string{os.Getenv("SEARCH10_INDEX_DEFAULT")}
	}
	if len(index) > 0 {
		index := index[len(index)-1]
		switch index {
		case "yes", "y", "t", "true":
			// default
			if zoektCache != nil {
				tr.LazyPrintf("%d indexed repos, %d unindexed repos", len(zoektRepos), len(searcherRepos))
			}
		case "only", "o", "force":
			if zoektCache == nil {
				return nil, common, fmt.Errorf("invalid index:%q (indexed search is not enabled)", index)
			}
			if os.Getenv("SEARCH_UNINDEXED_NOMISSING") == "" {
				common.missing = make([]api.RepoURI, len(searcherRepos))
				for i, r := range searcherRepos {
					common.missing[i] = r.repo.URI
				}
			}
			tr.LazyPrintf("index:only, ignoring %d unindexed repos", len(searcherRepos))
			searcherRepos = nil
		case "no", "n", "f", "false":
			tr.LazyPrintf("index:no, bypassing zoekt (using searcher) for %d indexed repos", len(zoektRepos))
			searcherRepos = append(searcherRepos, zoektRepos...)
			zoektRepos = nil
		default:
			return nil, common, fmt.Errorf("invalid index:%q (valid values are: yes, only, no)", index)
		}
	}

	var (
		wg                sync.WaitGroup
		mu                sync.Mutex
		unflattened       [][]*fileMatch
		flattenedSize     int
		overLimitCanceled bool // canceled because we were over the limit
	)

	// addMatches assumes the caller holds mu.
	addMatches := func(matches []*fileMatch) {
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
			if flattenedSize > int(args.query.FileMatchLimit) {
				tr.LazyPrintf("cancel due to result size: %d > %d", flattenedSize, args.query.FileMatchLimit)
				overLimitCanceled = true
				common.limitHit = true
				cancel()
			}
		}
	}

	for _, repoRev := range searcherRepos {
		if len(repoRev.revs) == 0 {
			return nil, common, nil // no revs to search
		}
		if len(repoRev.revs) >= 2 {
			return nil, common, errMultipleRevsNotSupported
		}

		wg.Add(1)
		go func(repoRev repositoryRevisions) {
			defer wg.Done()
			rev := repoRev.revspecs()[0] // TODO(sqs): search multiple revs
			matches, repoLimitHit, searchErr := searchFilesInRepo(ctx, repoRev.repo, repoRev.gitserverRepo, rev, args.query)
			mu.Lock()
			defer mu.Unlock()
			if ctx.Err() == nil {
				common.searched = append(common.searched, repoRev.repo.URI)
			}
			// non-diff search reports timeout through searchErr, so pass false for timedOut
			if fatalErr := handleRepoSearchResult(common, repoRev, repoLimitHit, false, searchErr); fatalErr != nil {
				if ctx.Err() != nil {
					// Our request has been canceled, we can just ignore
					// searchFilesInRepo for this repo. We only check this condition
					// here since handleRepoSearchResult handles deadlines
					// exceeded differently to canceled.
					return
				}
				err = errors.Wrapf(searchErr, "failed to search %s", repoRev.String())
				tr.LazyPrintf("cancel due to error: %v", err)
				cancel()
			}
			addMatches(matches)
		}(*repoRev)
	}

	wg.Add(1)
	go func() {
		// TODO limitHit, handleRepoSearchResult
		defer wg.Done()
		matches, limitHit, searchErr := zoektSearchHEAD(ctx, args.query, zoektRepos)
		mu.Lock()
		defer mu.Unlock()
		if ctx.Err() == nil {
			for _, repo := range zoektRepos {
				common.searched = append(common.searched, repo.repo.URI)
				common.indexed = append(common.indexed, repo.repo.URI)
			}
		}
		if limitHit {
			common.limitHit = true
		}
		if searchErr != nil && err == nil && !overLimitCanceled {
			err = searchErr
			tr.LazyPrintf("cancel indexed search due to error: %v", err)
			cancel()
		}
		addMatches(matches)
	}()

	wg.Wait()
	if err != nil {
		return nil, common, err
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

var zoektCl zoekt.Searcher
var zoektCache *zoektpkg.Cache
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
		zoektCl = zoektrpc.Client(zoektHost)
		zoektCache = &zoektpkg.Cache{Client: zoektCl}
	}
}
