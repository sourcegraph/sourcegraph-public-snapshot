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

	"github.com/pkg/errors"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	opentracing "github.com/opentracing/opentracing-go"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api/legacyerr"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/endpoint"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/zoekt"
)

// A light wrapper around the search service. We implement the service here so
// that we can unmarshal the result directly into graphql resolvers.

// patternInfo is the struct used by vscode pass on search queries.
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
	searcherURL := searcherURLs.Get(repo + "@" + commit)
	req, err := http.NewRequest("GET", searcherURL, nil)
	if err != nil {
		return nil, false, err
	}
	req.URL.RawQuery = q.Encode()
	req = req.WithContext(ctx)

	req, ht := nethttp.TraceRequest(opentracing.GlobalTracer(), req,
		nethttp.OperationName("Searcher Client"),
		nethttp.ClientTrace(false))
	defer ht.Finish()

	client := &http.Client{Transport: &nethttp.Transport{}}
	resp, err := client.Do(req)
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
	return e.StatusCode == 400
}

func (e *searcherError) Error() string {
	return e.Message
}

// isBadRequest will check if error or one of its causes is a bad request.
func isBadRequest(err error) bool {
	type badRequester interface {
		BadRequest() bool
	}
	type causer interface {
		Cause() error
	}

	for err != nil {
		if badrequest, ok := err.(badRequester); ok && badrequest.BadRequest() {
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
	} else if errors.Cause(searchErr) == context.DeadlineExceeded {
		common.timedout = append(common.timedout, repoRev.repo.URI)
	} else if searchErr != nil {
		return searchErr
	}
	return nil
}

func zoektSearchHEAD(ctx context.Context, args *repoSearchArgs) ([]*fileMatch, error) {
	// Convert sourcegraph pattern into zoekt query
	pattern := args.query.Pattern
	if !args.query.IsRegExp {
		pattern = regexp.QuoteMeta(pattern)
	}
	pattern = strconv.Quote(pattern)
	query := []string{pattern}

	// zoekt guesses case sensitivity if we don't specify
	if args.query.IsCaseSensitive {
		query = append(query, "case:yes")
	} else {
		query = append(query, "case:no")
	}

	// zoekt also uses regular expressions for file paths
	// TODO PathPatternsAreCaseSensitive
	// TODO whitespace in file path patterns?
	if !args.query.PathPatternsAreRegExps {
		return nil, errors.New("zoekt only supports regex path patterns")
	}
	for _, p := range args.query.IncludePatterns {
		query = append(query, "f:"+p)
	}
	if args.query.ExcludePattern != nil {
		query = append(query, "-f:"+*args.query.ExcludePattern)
	}

	// Tell zoekt which repos to search
	var restrict []zoekt.SearchRequestRestriction
	for _, repoRev := range args.repos {
		// We search HEAD using zoekt
		if repoRev.revSpecsOrDefaultBranch()[0] != "" {
			continue
		}
		// TODO Repo is a substring match, so we can match more
		restrict = append(restrict, zoekt.SearchRequestRestriction{
			Repo:     repoRev.repo.URI,
			Branches: []string{""}, // "" matches all indexed branches
		})
	}

	if len(restrict) == 0 {
		return nil, nil
	}

	zoektCl := &zoekt.Client{Host: zoektHost}
	resp, err := zoektCl.Search(ctx, zoekt.SearchRequest{
		Query:    strings.Join(query, " "),
		Restrict: restrict,
	})
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

var mockSearchRepos func(args *repoSearchArgs) ([]*searchResult, *searchResultsCommon, error)

// searchRepos searches a set of repos for a pattern.
func searchRepos(ctx context.Context, args *repoSearchArgs) ([]*searchResult, *searchResultsCommon, error) {
	if mockSearchRepos != nil {
		return mockSearchRepos(args)
	}

	if err := args.query.validate(); err != nil {
		return nil, nil, &badRequestError{err}
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		err         error
		wg          sync.WaitGroup
		mu          sync.Mutex
		unflattened [][]*fileMatch
		common      = &searchResultsCommon{}
	)
	for _, repoRev := range args.repos {
		if len(repoRev.revs) >= 2 {
			return nil, nil, errMultipleRevsNotSupported
		}

		// We search HEAD using zoekt
		if zoektHost != "" && repoRev.revSpecsOrDefaultBranch()[0] == "" {
			continue
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
				cancel()
			}
			if len(matches) > 0 {
				sort.Slice(matches, func(i, j int) bool {
					a, b := matches[i].uri, matches[j].uri
					return a > b
				})
				unflattened = append(unflattened, matches)
			}
		}(*repoRev)
	}

	if zoektHost != "" {
		wg.Add(1)
		go func() {
			// TODO limitHit, handleRepoSearchResult
			defer wg.Done()
			matches, searchErr := zoektSearchHEAD(ctx, args)
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
	}

	wg.Wait()
	if err != nil {
		return nil, nil, err
	}

	// Return early so we don't have to worry about empty lists in later
	// calculations.
	if len(unflattened) == 0 {
		return nil, common, nil
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
	initialPortion := int(args.query.FileMatchLimit) / len(unflattened)
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
		high := low + (int(args.query.FileMatchLimit) - len(flattened))
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

	return fileMatchesToSearchResults(flattened), common, nil
}

var zoektHost = env.Get("ZOEKT_HOST", "", "host:port of the zoekt instance")
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
}
