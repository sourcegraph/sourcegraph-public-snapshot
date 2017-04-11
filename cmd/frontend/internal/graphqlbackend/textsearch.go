package graphqlbackend

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	opentracing "github.com/opentracing/opentracing-go"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/endpoint"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
)

// A light wrapper around the search service. We implement the service here so
// that we can unmarshal the result directly into graphql resolvers.

// patternInfo is the struct used by vscode pass on search queries.
type patternInfo struct {
	Pattern         string
	IsRegExp        bool
	IsWordMatch     bool
	IsCaseSensitive bool
	MaxResults      int32
	// We do not support IsMultiline
	//IsMultiline     bool
}

type searchResults struct {
	results     []*fileMatch
	hasNextPage bool
}

func (sr *searchResults) Results() []*fileMatch {
	return sr.results
}

func (sr *searchResults) HasNextPage() bool {
	return sr.hasNextPage
}

type fileMatch struct {
	JPath        string       `json:"Path"`
	JLineMatches []*lineMatch `json:"LineMatches"`
	uri          string
}

func (fm *fileMatch) Resource() string {
	return fm.uri
}

func (fm *fileMatch) LineMatches() []*lineMatch {
	return fm.JLineMatches
}

// LineMatch is the struct used by vscode to receive search results for a line
type lineMatch struct {
	JPreview          string    `json:"Preview"`
	JLineNumber       int32     `json:"LineNumber"`
	JOffsetAndLengths [][]int32 `json:"OffsetAndLengths"`
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

// truncateMatches returns a copy of results with at most `limit` matches, and
// whether there was more than `limit` matches in the input slice.
func truncateMatches(results []*fileMatch, limit int) ([]*fileMatch, bool) {
	count := 0
	for i, fm := range results {
		for j, lm := range fm.JLineMatches {
			count += len(lm.JOffsetAndLengths)
			if count > limit {
				lMatch := *lm
				lMatch.JOffsetAndLengths = lMatch.JOffsetAndLengths[:len(lMatch.JOffsetAndLengths)-(count-limit)]
				lineMatches := make([]*lineMatch, j)
				copy(lineMatches, fm.JLineMatches)
				lineMatches = append(lineMatches, &lMatch)

				fMatch := *fm
				fMatch.JLineMatches = lineMatches

				fileMatches := make([]*fileMatch, i)
				copy(fileMatches, results)
				fileMatches = append(fileMatches, &fMatch)
				return fileMatches, true
			}
		}
	}
	return results, false
}

func (r *commitResolver) TextSearch(ctx context.Context, args *struct{ Query *patternInfo }) (*searchResults, error) {
	results, err := textSearch(ctx, r.repo.URI, r.commit.CommitID, args.Query)
	if err != nil {
		return nil, err
	}
	results, limitHit := truncateMatches(results, int(args.Query.MaxResults))
	return &searchResults{results, limitHit}, nil
}

func textSearch(ctx context.Context, repo, commit string, p *patternInfo) ([]*fileMatch, error) {
	if searcherURLs == nil {
		return nil, errors.New("a searcher service has not been configured")
	}
	q := url.Values{
		"Repo":    []string{repo},
		"Commit":  []string{commit},
		"Pattern": []string{p.Pattern},
	}
	if p.IsRegExp {
		q.Set("IsRegExp", "true")
	}
	if p.IsWordMatch {
		q.Set("IsWordMatch", "true")
	}
	if p.IsCaseSensitive {
		q.Set("IsCaseSensitive", "true")
	}
	searcherURL := searcherURLs.Get(repo + "@" + commit)
	req, err := http.NewRequest("GET", searcherURL, nil)
	if err != nil {
		return nil, err
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
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("non-200 response: code=%d body=%s", resp.StatusCode, string(body))
	}

	r := struct {
		Matches []*fileMatch
	}{}
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return nil, err
	}
	workspace := "git://" + repo + "?" + commit + "#"
	for _, fm := range r.Matches {
		fm.uri = workspace + fm.JPath
	}
	return r.Matches, nil
}

func searchRepo(ctx context.Context, repoName string, info *patternInfo) ([]*fileMatch, error) {
	repo, err := localstore.Repos.GetByURI(ctx, repoName)
	if err != nil {
		return nil, err
	}
	// 🚨 SECURITY: DO NOT REMOVE THIS CHECK! ResolveRev is responsible for ensuring 🚨
	// the user has permissions to access the repository.
	commit, err := backend.Repos.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{
		Repo: repo.ID,
	})
	if err != nil {
		return nil, err
	}
	return textSearch(ctx, repoName, commit.CommitID, info)
}

type repoSearchArgs struct {
	Query        *patternInfo
	Repositories []string
}

// SearchRepos searches a set of repos for a pattern.
func (*rootResolver) SearchRepos(ctx context.Context, args *repoSearchArgs) (*searchResults, error) {
	r, err := ParMap(func(repo string) ([]*fileMatch, error) {
		return searchRepo(ctx, repo, args.Query)
	}, args.Repositories)
	if err != nil {
		return nil, err
	}
	results := r.([][]*fileMatch)
	var flattened []*fileMatch
	for _, response := range results {
		flattened = append(flattened, response...)
	}
	sort.Slice(flattened, func(i, j int) bool {
		a, b := len(flattened[i].JLineMatches), len(flattened[j].JLineMatches)
		if a != b {
			return a > b
		}
		return flattened[i].uri > flattened[j].uri
	})
	maxLinesPerFile := 6
	for _, lm := range flattened {
		if len(lm.JLineMatches) > maxLinesPerFile {
			lm.JLineMatches = lm.JLineMatches[:maxLinesPerFile]
		}
	}
	flattened, truncated := truncateMatches(flattened, int(args.Query.MaxResults))
	return &searchResults{flattened, truncated}, nil
}

var searcherURLs *endpoint.Map

func init() {
	searcherURL := env.Get("SEARCHER_URL", "", "searcher server URL (eg http://localhost:3181)")
	if searcherURL == "" {
		return
	}
	var err error
	searcherURLs, err = endpoint.New(searcherURL)
	if err != nil {
		panic(fmt.Sprintf("could not connect to searcher %s: %s", searcherURL, err))
	}
}
