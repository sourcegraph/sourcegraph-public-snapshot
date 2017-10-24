package graphqlbackend

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"sync"

	"github.com/pkg/errors"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	opentracing "github.com/opentracing/opentracing-go"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api/legacyerr"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/endpoint"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
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

type searchResults struct {
	results  []*fileMatch
	limitHit bool
	cloning  []string
	missing  []string
}

func (sr *searchResults) Results() []*fileMatch {
	return sr.results
}

func (sr *searchResults) LimitHit() bool {
	return sr.limitHit
}

func (sr *searchResults) Cloning() []string {
	if sr.cloning == nil {
		return []string{}
	}
	return sr.cloning
}

func (sr *searchResults) Missing() []string {
	if sr.missing == nil {
		return []string{}
	}
	return sr.missing
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
		return nil, false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, false, err
		}
		return nil, false, fmt.Errorf("non-200 response: code=%d body=%s", resp.StatusCode, string(body))
	}

	r := struct {
		Matches  []*fileMatch
		LimitHit bool
	}{}
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return nil, false, err
	}
	workspace := "git://" + repo + "?" + commit + "#"
	for _, fm := range r.Matches {
		fm.uri = workspace + fm.JPath
	}
	return r.Matches, r.LimitHit, nil
}

func searchRepo(ctx context.Context, repoName, rev string, info *patternInfo) (matches []*fileMatch, limitHit bool, err error) {
	repo, err := localstore.Repos.GetByURI(ctx, repoName)
	if err != nil {
		return nil, false, err
	}
	// 🚨 SECURITY: DO NOT REMOVE THIS CHECK! ResolveRev is responsible for ensuring 🚨
	// the user has permissions to access the repository.
	commit, err := backend.Repos.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{
		Repo: repo.ID,
		Rev:  rev,
	})
	if err != nil {
		return nil, false, err
	}
	return textSearch(ctx, repoName, commit.CommitID, info)
}

type repoSearchArgs struct {
	Query        *patternInfo
	Repositories []*repositoryRevision
}

// repositoryRevision specifies a repository at an (optional) revision. If no revision is
// specified, then the repository's default branch is used.
type repositoryRevision struct {
	Repo string
	Rev  *string
}

func (repoRev *repositoryRevision) String() string {
	if repoRev.hasRev() {
		return repoRev.Repo + "@" + *repoRev.Rev
	}
	return repoRev.Repo
}

func (repoRev *repositoryRevision) hasRev() bool {
	return repoRev.Rev != nil && *repoRev.Rev != ""
}

// SearchRepos searches a set of repos for a pattern.
func (*rootResolver) SearchRepos(ctx context.Context, args *repoSearchArgs) (*searchResults, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		err       error
		wg        sync.WaitGroup
		mu        sync.Mutex
		cloning   []string
		missing   []string
		flattened []*fileMatch
		limitHit  bool
	)
	for _, repoRev := range args.Repositories {
		wg.Add(1)
		go func(repoRev repositoryRevision) {
			defer wg.Done()
			var rev string
			if repoRev.Rev != nil {
				rev = *repoRev.Rev
			}
			matches, repoLimitHit, searchErr := searchRepo(ctx, repoRev.Repo, rev, args.Query)
			if ctx.Err() != nil {
				// Our request has been canceled, we can just ignore searchRepo for this repo.
				return
			}
			mu.Lock()
			defer mu.Unlock()
			limitHit = limitHit || repoLimitHit
			if e, ok := searchErr.(vcs.RepoNotExistError); ok {
				if e.CloneInProgress {
					cloning = append(cloning, repoRev.Repo)
				} else {
					missing = append(missing, repoRev.Repo)
				}
			} else if e, ok := searchErr.(legacyerr.Error); ok && e.Code == legacyerr.NotFound {
				missing = append(missing, repoRev.Repo)
			} else if searchErr == vcs.ErrRevisionNotFound && !repoRev.hasRev() {
				// If we didn't specify an input revision, then the repo is empty and can be ignored.
			} else if searchErr != nil && err == nil {
				err = errors.Wrapf(searchErr, "failed to search %s", repoRev.String())
				cancel()
			}
			if len(matches) > 0 {
				flattened = append(flattened, matches...)
			}
			if len(flattened) > int(args.Query.FileMatchLimit) {
				// We can stop collecting more results.
				cancel()
			}
		}(*repoRev)
	}
	wg.Wait()
	if err != nil {
		return nil, err
	}
	sort.Slice(flattened, func(i, j int) bool {
		a, b := flattened[i].uri, flattened[j].uri
		return a > b
	})
	// We pass in a limit to each repository so we may end up with R*limit results
	// where R is the number of repositories we searched.
	// Clip the results after doing the "relevance" sorting above.
	if len(flattened) > int(args.Query.FileMatchLimit) {
		flattened = flattened[:args.Query.FileMatchLimit]
	}
	return &searchResults{
		results:  flattened,
		limitHit: limitHit,
		cloning:  cloning,
		missing:  missing,
	}, nil
}

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
