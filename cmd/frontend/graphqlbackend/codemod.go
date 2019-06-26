package graphqlbackend

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search/query"
	"github.com/sourcegraph/sourcegraph/pkg/env"
	"github.com/sourcegraph/sourcegraph/pkg/errcode"
	"github.com/sourcegraph/sourcegraph/pkg/trace"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
	"golang.org/x/net/context/ctxhttp"
)

// codemodResultResolver is a resolver for the GraphQL type `CommitSearchResult`
type codemodResultResolver struct {
	commit  *GitCommitResolver
	path    string
	fileURL string
	diff    string
	matches []*searchResultMatchResolver
}

func (r *codemodResultResolver) Commit() *GitCommitResolver { return r.commit }
func (r *codemodResultResolver) Icon() string {
	return "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' style='width:24px;height:24px' viewBox='0 0 24 24'%3E%3Cpath fill='%23a2b0cd' d='M11,6C12.38,6 13.63,6.56 14.54,7.46L12,10H18V4L15.95,6.05C14.68,4.78 12.93,4 11,4C7.47,4 4.57,6.61 4.08,10H6.1C6.56,7.72 8.58,6 11,6M16.64,15.14C17.3,14.24 17.76,13.17 17.92,12H15.9C15.44,14.28 13.42,16 11,16C9.62,16 8.37,15.44 7.46,14.54L10,12H4V18L6.05,15.95C7.32,17.22 9.07,18 11,18C12.55,18 14,17.5 15.14,16.64L20,21.5L21.5,20L16.64,15.14Z' /%3E%3C/svg%3E"
}
func (r *codemodResultResolver) Label() (*markdownResolver, error) {
	commitURL, err := r.commit.URL()
	if err != nil {
		return nil, err
	}
	text := fmt.Sprintf("[%s](%s) › [%s](%s)", r.commit.repo.Name(), commitURL, r.path, r.fileURL)
	return &markdownResolver{text: text}, nil
}

func (r *codemodResultResolver) URL() string {
	return ""
}

func (r *codemodResultResolver) Detail() (*markdownResolver, error) {
	// diff, err := diff.ParseFileDiff(r.diff)
	// if err != nil {return nil, err}
	// diff.Stat()
	return &markdownResolver{text: strconv.Itoa(len(r.matches))}, nil
}

func (r *codemodResultResolver) Matches() []*searchResultMatchResolver {
	return r.matches
}

func (r *codemodResultResolver) ToRepository() (*RepositoryResolver, bool) { return nil, false }
func (r *codemodResultResolver) ToFileMatch() (*fileMatchResolver, bool)   { return nil, false }
func (r *codemodResultResolver) ToCommitSearchResult() (*commitSearchResultResolver, bool) {
	return nil, false
}
func (r *codemodResultResolver) ToCodemodResult() (*codemodResultResolver, bool) { return r, false }

func (r *codemodResultResolver) searchResultURIs() (string, string) {
	return r.commit.repo.Name(), r.path
}

func (r *codemodResultResolver) resultCount() int32 {
	return int32(len(r.matches))
}

func callCodemod(ctx context.Context, args *search.Args) ([]searchResultResolver, *searchResultsCommon, error) {
	matchValues := args.Query.Values(query.FieldDefault)
	var matchPatterns []string
	for _, v := range matchValues {
		if v.String != nil && *v.String != "" {
			matchPatterns = append(matchPatterns, *v.String)
		}
		if v.Regexp != nil {
			// HACK
			matchPatterns = append(matchPatterns, strings.Replace(v.Regexp.String(), "\\", "", -1))
		}
	}
	matchPattern := strings.Join(matchPatterns, " ")
	replacementValues, _ := args.Query.StringValues(query.FieldReplace)
	var replacementText string
	if len(replacementValues) > 0 {
		replacementText = replacementValues[0]
	}

	var err error
	tr, ctx := trace.New(ctx, "callCodemod", fmt.Sprintf("pattern: %+v, replace: %+v, numRepoRevs: %d", matchPattern, replacementText, len(args.Repos)))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		wg          sync.WaitGroup
		mu          sync.Mutex
		unflattened [][]*codemodResultResolver
		common      = &searchResultsCommon{}
	)
	for _, repoRev := range args.Repos {
		wg.Add(1)
		go func(repoRev search.RepositoryRevisions) {
			defer wg.Done()
			results, searchErr := callCodemodInRepo(ctx, repoRev, matchPattern, replacementText)
			if ctx.Err() == context.Canceled {
				// Our request has been canceled (either because another one of args.repos had a
				// fatal error, or otherwise), so we can just ignore these results.
				return
			}
			repoTimedOut := ctx.Err() == context.DeadlineExceeded
			if searchErr != nil {
				tr.LogFields(otlog.String("repo", string(repoRev.Repo.Name)), otlog.String("searchErr", searchErr.Error()), otlog.Bool("timeout", errcode.IsTimeout(searchErr)), otlog.Bool("temporary", errcode.IsTemporary(searchErr)))
			}
			mu.Lock()
			defer mu.Unlock()
			if fatalErr := handleRepoSearchResult(common, repoRev, false, repoTimedOut, searchErr); fatalErr != nil {
				err = errors.Wrapf(searchErr, "failed to call codemod %s", repoRev.String())
				cancel()
			}
			if len(results) > 0 {
				unflattened = append(unflattened, results)
			}
		}(*repoRev)
	}
	wg.Wait()
	if err != nil {
		return nil, nil, err
	}

	var results []searchResultResolver
	for _, ur := range unflattened {
		for _, r := range ur {
			results = append(results, r)
		}
	}
	return results, common, nil
}

var replacerURL = env.Get("REPLACER_URL", "http://replacer:3185", "replacer server URL")

func callCodemodInRepo(ctx context.Context, repoRevs search.RepositoryRevisions, matchPattern, replacementText string) (results []*codemodResultResolver, err error) {
	tr, ctx := trace.New(ctx, "callCodemodInRepo", fmt.Sprintf("repoRevs: %v, pattern %+v, replace: %+v", repoRevs, matchPattern, replacementText))
	defer func() {
		tr.LazyPrintf("%d results", len(results))
		tr.SetError(err)
		tr.Finish()
	}()

	// Do not trigger a repo-updater lookup (e.g.,
	// backend.{GitRepo,Repos.ResolveRev}) because that would slow this operation
	// down by a lot (if we're looping over many repos). This means that it'll fail if a
	// repo is not on gitserver.
	commit, err := git.ResolveRevision(ctx, repoRevs.GitserverRepo(), nil, repoRevs.Revs[0].RevSpec, &git.ResolveRevisionOptions{NoEnsureRevision: true})
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(replacerURL)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("repo", string(repoRevs.Repo.Name))
	q.Set("commit", string(commit))
	q.Set("matchtemplate", matchPattern)
	q.Set("rewritetemplate", replacementText)
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	req, ht := nethttp.TraceRequest(opentracing.GlobalTracer(), req,
		nethttp.OperationName("Codemod client"),
		nethttp.ClientTrace(false))
	defer ht.Finish()

	resp, err := ctxhttp.Do(ctx, searchHTTPClient, req)
	if err != nil {
		// If we failed due to cancellation or timeout (with no partial results in the response
		// body), return just that.
		if ctx.Err() != nil {
			err = ctx.Err()
		}
		return nil, errors.Wrap(err, "codemod request failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, errors.WithStack(&searcherError{StatusCode: resp.StatusCode, Message: string(body)})
	}

	type rawCodemodResult struct {
		URI                  string `json:"uri"`
		RewrittenSource      string `json:"rewritten_source"`
		InPlaceSubstitutions []struct {
			Range struct {
				Start struct{ Offset int64 }
				End   struct{ Offset int64 }
			}
			ReplacementContent string `json:"replacement_content"`
			Environment        []struct {
				Value string
			}
		} `json:"in_place_substitutions"`
		Diff string
	}

	computeCodemodResultMatches := func(fileURL string, raw *rawCodemodResult) ([]*searchResultMatchResolver, error) {
		// rawDiff2, highlights, err := git.FilterAndHighlightDiff([]byte(rawDiff), nil, false, pathmatch.All)
		// if err != nil {
		// 	return nil, err
		// }

		seenMatches := map[string]struct{}{}
		matches := func(text string) (hs []*highlightedRange) {
			if _, seen := seenMatches[text]; seen {
				return nil
			}
			seenMatches[text] = struct{}{}
			lines := strings.Split(raw.Diff, "\n")
			for i, line := range lines {
				if line[1] == ' ' {
					continue
				}
				if pos := strings.Index(line, text); pos != -1 {
					hs = append(hs, &highlightedRange{
						line:      int32(i) - 1,
						character: int32(pos),
						length:    int32(len(text)),
					})
				}
			}
			return hs
		}

		var highlights []*highlightedRange
		for _, sub := range raw.InPlaceSubstitutions {
			highlights = append(highlights, matches(sub.ReplacementContent)...)
			for _, e := range sub.Environment {
				highlights = append(highlights, matches(e.Value)...)
			}
		}

		matchBody, matchHighlights := cleanDiffPreview(highlights, raw.Diff)
		_ = matchBody[strings.Index(matchBody, "@@"):]

		return []*searchResultMatchResolver{
			{
				url:        fileURL,
				body:       "```diff\n" + raw.Diff[strings.Index(raw.Diff, "@@"):] + "\n```",
				highlights: matchHighlights,
			},
		}, nil
	}

	var rawResults []*rawCodemodResult
	decoder := json.NewDecoder(resp.Body)
	for decoder.More() {
		var rawResult *rawCodemodResult
		if err := decoder.Decode(&rawResult); err != nil {
			return nil, errors.Wrap(err, "replacer response invalid")
		}
		if len(rawResult.InPlaceSubstitutions) == 0 {
			continue
		}
		rawResults = append(rawResults, rawResult)
	}
	results = make([]*codemodResultResolver, len(rawResults))
	for i, raw := range rawResults {
		fileURL := fileMatchURI(repoRevs.Repo.Name, repoRevs.Revs[0].RevSpec, raw.URI)
		matches, err := computeCodemodResultMatches(fileURL, raw)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("invalid result (%d)", i))
		}
		results[i] = &codemodResultResolver{
			commit: &GitCommitResolver{
				repo:     &RepositoryResolver{repo: repoRevs.Repo},
				inputRev: &repoRevs.Revs[0].RevSpec,
			},
			path:    raw.URI,
			fileURL: fileURL,
			diff:    raw.Diff,
			matches: matches,
		}
	}
	return results, nil
}
