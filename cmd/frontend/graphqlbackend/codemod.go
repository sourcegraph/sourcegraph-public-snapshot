package graphqlbackend

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search/query"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
	"golang.org/x/net/context/ctxhttp"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

type rawCodemodResult struct {
	URI  string `json:"uri"`
	Diff string
}

type args struct {
	matchTemplate     string
	rewriteTemplate   string
	includeFileFilter string
	excludeFileFilter string
}

// codemodResultResolver is a resolver for the GraphQL type `CodemodResult`
type codemodResultResolver struct {
	commit  *GitCommitResolver
	path    string
	fileURL string
	diff    string
	matches []*searchResultMatchResolver
}

func (r *codemodResultResolver) ToRepository() (*RepositoryResolver, bool) { return nil, false }
func (r *codemodResultResolver) ToFileMatch() (*FileMatchResolver, bool)   { return nil, false }
func (r *codemodResultResolver) ToCommitSearchResult() (*commitSearchResultResolver, bool) {
	return nil, false
}

func (r *codemodResultResolver) ToCodemodResult() (*codemodResultResolver, bool) {
	return r, true
}

func (r *codemodResultResolver) searchResultURIs() (string, string) {
	return string(r.commit.repo.repo.Name), r.path
}

func (r *codemodResultResolver) resultCount() int32 {
	return int32(len(r.matches))
}

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
	return r.fileURL
}

func (r *codemodResultResolver) Detail() (*markdownResolver, error) {
	diff, err := diff.ParseFileDiff([]byte(r.diff))
	if err != nil {
		return nil, err
	}
	stat := diff.Stat()
	return &markdownResolver{text: stat.String()}, nil
}

func (r *codemodResultResolver) Matches() []*searchResultMatchResolver {
	return r.matches
}

func (r *codemodResultResolver) Commit() *GitCommitResolver { return r.commit }

func (r *codemodResultResolver) RawDiff() string { return r.diff }

func validateQuery(q *query.Query) (*args, error) {
	matchValues := q.Values(query.FieldDefault)
	var matchTemplates []string
	for _, v := range matchValues {
		if v.String != nil && *v.String != "" {
			matchTemplates = append(matchTemplates, *v.String)
		}
		if v.Regexp != nil || v.Bool != nil {
			return nil, errors.New("this looks like a regex search pattern. Structural search is active because 'replace:' was specified. Please enclose your search string with quotes when using 'replace:'.")
		}
	}
	matchTemplate := strings.Join(matchTemplates, " ")

	replacementValues, _ := q.StringValues(query.FieldReplace)
	var rewriteTemplate string
	if len(replacementValues) > 0 {
		rewriteTemplate = replacementValues[0]
	}

	includeFileFilter, excludeFileFilter := q.RegexpPatterns(query.FieldFile)
	var includeFileFilterText string
	if len(includeFileFilter) > 0 {
		includeFileFilterText = includeFileFilter[0]
		// only file names or files with extensions in the following characterset are allowed
		IsAlphanumericWithPeriod := lazyregexp.New(`^[a-zA-Z0-9_.]+$`).MatchString
		if !IsAlphanumericWithPeriod(includeFileFilterText) {
			return nil, errors.New("the 'file:' filter cannot contain regex when using the 'replace:' filter currently. Only alphanumeric characters or '.'")
		}
	}

	var excludeFileFilterText string
	if len(excludeFileFilter) > 0 {
		excludeFileFilterText = excludeFileFilter[0]
		IsAlphanumericWithPeriod := lazyregexp.New(`^[a-zA-Z_.]+$`).MatchString
		if !IsAlphanumericWithPeriod(includeFileFilterText) {
			return nil, errors.New("the '-file:' filter cannot contain regex when using the 'replace:' filter currently. Only alphanumeric characters or '.'")
		}
	}

	return &args{matchTemplate, rewriteTemplate, includeFileFilterText, excludeFileFilterText}, nil
}

// Calls the codemod backend replacer service for a set of repository revisions.
func performCodemod(ctx context.Context, args *search.Args) ([]SearchResultResolver, *searchResultsCommon, error) {
	cmodArgs, err := validateQuery(args.Query)
	if err != nil {
		return nil, nil, err
	}

	title := fmt.Sprintf("pattern: %+v, replace: %+v, includeFileFilter: %+v, excludeFileFilter: %+v, numRepoRevs: %d", cmodArgs.matchTemplate, cmodArgs.rewriteTemplate, cmodArgs.includeFileFilter, cmodArgs.excludeFileFilter, len(args.Repos))
	tr, ctx := trace.New(ctx, "callCodemod", title)
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		wg          sync.WaitGroup
		mu          sync.Mutex
		unflattened [][]codemodResultResolver
		common      = &searchResultsCommon{}
	)
	for _, repoRev := range args.Repos {
		wg.Add(1)
		repoRev := repoRev // shadow variable so it doesn't change while goroutine is running
		goroutine.Go(func() {
			defer wg.Done()
			results, searchErr := callCodemodInRepo(ctx, repoRev, cmodArgs)
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
				err = errors.Wrapf(searchErr, "failed to call codemod %s", repoRev)
				cancel()
			}
			if len(results) > 0 {
				unflattened = append(unflattened, results)
			}
		})
	}
	wg.Wait()
	if err != nil {
		return nil, nil, err
	}

	var results []SearchResultResolver
	for _, ur := range unflattened {
		for _, resolver := range ur {
			v := resolver
			results = append(results, &v)
		}
	}

	return results, common, nil
}

var ReplacerURL = env.Get("REPLACER_URL", "http://replacer:3185", "replacer server URL")

func toMatchResolver(fileURL string, raw *rawCodemodResult) ([]*searchResultMatchResolver, error) {
	if !strings.Contains(raw.Diff, "@@") {
		return nil, errors.Errorf("Invalid diff does not contain expected @@: %v", raw.Diff)
	}
	strippedDiff := raw.Diff[strings.Index(raw.Diff, "@@"):]

	return []*searchResultMatchResolver{
			{
				url:        fileURL,
				body:       "```diff\n" + strippedDiff + "\n```",
				highlights: nil,
			},
		},
		nil
}

func callCodemodInRepo(ctx context.Context, repoRevs *search.RepositoryRevisions, args *args) (results []codemodResultResolver, err error) {
	tr, ctx := trace.New(ctx, "callCodemodInRepo", fmt.Sprintf("repoRevs: %v, pattern %+v, replace: %+v", repoRevs, args.matchTemplate, args.rewriteTemplate))
	defer func() {
		tr.LazyPrintf("%d results", len(results))
		tr.SetError(err)
		tr.Finish()
	}()

	// For performance, assume repo is cloned in gitserver and do not trigger a repo-updater lookup (this call fails if repo is not on gitserver).
	commit, err := git.ResolveRevision(ctx, repoRevs.GitserverRepo(), nil, repoRevs.Revs[0].RevSpec, &git.ResolveRevisionOptions{NoEnsureRevision: true})
	if err != nil {
		return nil, errors.Wrap(err, "codemod repo lookup failed: it's possible that the repo is not cloned in gitserver. Try force a repo update another way.")
	}

	u, err := url.Parse(ReplacerURL)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("repo", string(repoRevs.Repo.Name))
	q.Set("commit", string(commit))
	q.Set("matchtemplate", args.matchTemplate)
	q.Set("rewritetemplate", args.rewriteTemplate)
	q.Set("fileextension", args.includeFileFilter)
	q.Set("directoryexclude", args.excludeFileFilter)
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

	// TODO(RVT): Use a separate HTTP client here dedicated to codemod,
	// not doing so means codemod and searcher share the same HTTP limits
	// etc. which is fine for now but not if codemod goes in front of users.
	// Once separated please fix cmd/frontend/graphqlbackend/textsearch:50 in #6586
	resp, err := ctxhttp.Do(ctx, searchHTTPClient, req)
	if err != nil {
		// If we failed due to cancellation or timeout (with no partial results in the response body), return just that.
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

	scanner := bufio.NewScanner(resp.Body)
	// TODO(RVT): Remove buffer inefficiency introduced here. Results are
	// line encoded JSON. Malformed JSON can happen, for extremely long
	// lines. Unless we know where this and subsequent malformed lines end,
	// we can't continue decoding the next one. As an inefficient but robust
	// solution, we allow to buffer a very high maximum length line and can
	// skip over very long malformed lines. It is set to 10 * 64K.
	scanner.Buffer(make([]byte, 100), 10*bufio.MaxScanTokenSize)

	for scanner.Scan() {
		var raw *rawCodemodResult
		b := scanner.Bytes()
		if err := scanner.Err(); err != nil {
			log15.Info(fmt.Sprintf("Skipping codemod scanner error (line too long?): %s", err.Error()))
			continue
		}
		if err := json.Unmarshal(b, &raw); err != nil {
			// skip on other decode errors (including e.g., empty
			// responses if dependencies are not installed)
			continue
		}
		fileURL := fileMatchURI(repoRevs.Repo.Name, repoRevs.Revs[0].RevSpec, raw.URI)
		matches, err := toMatchResolver(fileURL, raw)
		if err != nil {
			return nil, err
		}
		results = append(results, codemodResultResolver{
			commit: &GitCommitResolver{
				repo:     &RepositoryResolver{repo: repoRevs.Repo},
				inputRev: &repoRevs.Revs[0].RevSpec,
				oid:      GitObjectID(commit),
			},
			path:    raw.URI,
			fileURL: fileURL,
			diff:    raw.Diff,
			matches: matches,
		})
	}

	return results, nil
}
