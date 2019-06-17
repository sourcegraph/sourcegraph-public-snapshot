package graphqlbackend

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
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

func validateArgs(args *search.Args) (string, string, string, error) {
	matchValues := args.Query.Values(query.FieldDefault)
	var matchTemplates []string
	for _, v := range matchValues {
		if v.String != nil && *v.String != "" {
			matchTemplates = append(matchTemplates, *v.String)
		}
		if v.Regexp != nil || v.Bool != nil {
			return "", "", "", errors.New("This looks like a regex search pattern. Structural search is active because 'replace:' was specified. Please enclose your search string with quotes when using 'replace:'.")
		}
	}
	matchTemplate := strings.Join(matchTemplates, " ")

	replacementValues, _ := args.Query.StringValues(query.FieldReplace)
	var rewriteTemplate string
	if len(replacementValues) > 0 {
		rewriteTemplate = replacementValues[0]
	}

	fileFilter, _ := args.Query.RegexpPatterns(query.FieldFile)
	var fileFilterText string
	if len(fileFilter) > 0 {
		fileFilterText = fileFilter[0]
		// only file names or files with extensions in the following characterset are allowed
		var IsAlphanumericWithPeriod = regexp.MustCompile(`^[a-zA-Z_.]+$`).MatchString
		if !IsAlphanumericWithPeriod(fileFilterText) {
			return matchTemplate, rewriteTemplate, "", errors.New("Note: the 'file:' filter cannot contain regex when using the 'replace:' filter. Only alphanumeric characters or '.'")
		}
	}
	return matchTemplate, rewriteTemplate, fileFilterText, nil
}

func callCodemod(ctx context.Context, args *search.Args) ([]*searchResultResolver, *searchResultsCommon, error) {
	matchTemplate, rewriteTemplate, fileFilter, err := validateArgs(args)
	if err != nil {
		return nil, nil, err
	}

	tr, ctx := trace.New(ctx, "callCodemod", fmt.Sprintf("pattern: %+v, replace: %+v, fileFilter: %+v, numRepoRevs: %d", matchTemplate, rewriteTemplate, fileFilter, len(args.Repos)))
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
			results, searchErr := callCodemodInRepo(ctx, repoRev, matchTemplate, rewriteTemplate, fileFilter)
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

	var results []*searchResultResolver
	for _, ur := range unflattened {
		for _, r := range ur {
			results = append(results, &searchResultResolver{codemod: r})
		}
	}
	return results, common, nil
}

var replacerURL = env.Get("REPLACER_URL", "http://replacer:3185", "replacer server URL")

func toMatchResolver(fileURL string, raw *rawCodemodResult) []*searchResultMatchResolver {
	var highlights []*highlightedRange
	matchBody, matchHighlights := cleanDiffPreview(highlights, raw.Diff)
	_ = matchBody[strings.Index(matchBody, "@@"):]

	return []*searchResultMatchResolver{
		{
			url:        fileURL,
			body:       "```diff\n" + raw.Diff[strings.Index(raw.Diff, "@@"):] + "\n```",
			highlights: matchHighlights,
		},
	}
}

func callCodemodInRepo(ctx context.Context, repoRevs search.RepositoryRevisions, matchPattern, replacementText string, fileFilter string) (results []*codemodResultResolver, err error) {
	tr, ctx := trace.New(ctx, "callCodemodInRepo", fmt.Sprintf("repoRevs: %v, pattern %+v, replace: %+v", repoRevs, matchPattern, replacementText))
	defer func() {
		tr.LazyPrintf("%d results", len(results))
		tr.SetError(err)
		tr.Finish()
	}()

	// For performance, assume repo is cloned in gitserver and do not trigger a repo-updater lookup (this call fails if repo is not on gitserver).
	commit, err := git.ResolveRevision(ctx, repoRevs.GitserverRepo(), nil, repoRevs.Revs[0].RevSpec, &git.ResolveRevisionOptions{NoEnsureRevision: true})
	if err != nil {
		return nil, errors.Wrap(err, "Codemod repo lookup failed: it's possible that the repo is not cloned in gitserver. Try force a repo update another way.")
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
	q.Set("fileextension", fileFilter)
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

	decoder := json.NewDecoder(resp.Body)
	for decoder.More() {
		var raw *rawCodemodResult
		if err := decoder.Decode(&raw); err != nil {
			return nil, errors.Wrap(err, "Replacer response invalid.")
		}

		fileURL := fileMatchURI(repoRevs.Repo.Name, repoRevs.Revs[0].RevSpec, raw.URI)
		matches := toMatchResolver(fileURL, raw)
		result := &codemodResultResolver{
			commit: &gitCommitResolver{
				repo:     &repositoryResolver{repo: repoRevs.Repo},
				inputRev: &repoRevs.Revs[0].RevSpec,
			},
			path:    raw.URI,
			fileURL: fileURL,
			diff:    raw.Diff,
			matches: matches,
		}
		results = append(results, result)
	}
	return results, nil
}
