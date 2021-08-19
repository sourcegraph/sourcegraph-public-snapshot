// Package searcher provides a client for our just in time text searching
// service "searcher".
package searcher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/opentracing-contrib/go-stdlib/nethttp"

	"github.com/sourcegraph/sourcegraph/cmd/searcher/protocol"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/endpoint"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
)

var (
	searchDoer, _ = httpcli.NewInternalClientFactory("search").Doer()
	MockSearch    func(ctx context.Context, repo api.RepoName, commit api.CommitID, p *search.TextPatternInfo, fetchTimeout time.Duration) (matches []*protocol.FileMatch, limitHit bool, err error)
)

// Search searches repo@commit with p.
func Search(
	ctx context.Context,
	searcherURLs *endpoint.Map,
	repo api.RepoName,
	branch string,
	commit api.CommitID,
	indexed bool,
	p *search.TextPatternInfo,
	fetchTimeout time.Duration,
	indexerEndpoints []string,
	onMatches func([]*protocol.FileMatch),
) (matches []*protocol.FileMatch, limitHit bool, err error) {
	if MockSearch != nil {
		return MockSearch(ctx, repo, commit, p, fetchTimeout)
	}

	tr, ctx := trace.New(ctx, "searcher.client", fmt.Sprintf("%s@%s", repo, commit))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	r := protocol.Request{
		Repo:   repo,
		Commit: commit,
		Branch: branch,
		PatternInfo: protocol.PatternInfo{
			Pattern:                      p.Pattern,
			ExcludePattern:               p.ExcludePattern,
			IncludePatterns:              p.IncludePatterns,
			Languages:                    p.Languages,
			CombyRule:                    p.CombyRule,
			PathPatternsAreRegExps:       true,
			Select:                       p.Select.Root(),
			Limit:                        int(p.FileMatchLimit),
			IsRegExp:                     p.IsRegExp,
			IsStructuralPat:              p.IsStructuralPat,
			IsWordMatch:                  p.IsWordMatch,
			IsCaseSensitive:              p.IsCaseSensitive,
			PathPatternsAreCaseSensitive: p.PathPatternsAreCaseSensitive,
			IsNegated:                    p.IsNegated,
			PatternMatchesContent:        p.PatternMatchesContent,
			PatternMatchesPath:           p.PatternMatchesPath,
		},
		Indexed:          indexed,
		FetchTimeout:     fetchTimeout.String(),
		IndexerEndpoints: indexerEndpoints,
	}

	if deadline, ok := ctx.Deadline(); ok {
		t, err := deadline.MarshalText()
		if err != nil {
			return nil, false, err
		}
		r.Deadline = string(t)
	}
	if onMatches != nil {
		r.Stream = true
	}
	body, err := json.Marshal(r)
	if err != nil {
		return nil, false, err
	}

	// Searcher caches the file contents for repo@commit since it is
	// relatively expensive to fetch from gitserver. So we use consistent
	// hashing to increase cache hits.
	consistentHashKey := string(repo) + "@" + string(commit)
	tr.LazyPrintf("%s", consistentHashKey)

	nodes, err := searcherURLs.Endpoints()
	if err != nil {
		return nil, false, err
	}

	urls, err := searcherURLs.GetN(consistentHashKey, len(nodes))
	if err != nil {
		return nil, false, err
	}

	for attempt := 0; attempt < 2; attempt++ {
		url := urls[attempt%len(urls)]

		tr.LazyPrintf("attempt %d: %s", attempt, url)
		if onMatches != nil {
			limitHit, err = textSearchStream(ctx, url, body, onMatches)
			if err == nil || errcode.IsTimeout(err) {
				return nil, limitHit, err
			}
		} else {
			matches, limitHit, err = textSearch(ctx, url, body)
			if err == nil || errcode.IsTimeout(err) {
				return matches, limitHit, err
			}
		}

		// If we are canceled, return that error.
		if err = ctx.Err(); err != nil {
			return nil, false, err
		}

		// If not temporary or our last attempt then don't try again.
		if !errcode.IsTemporary(err) {
			return nil, false, err
		}

		tr.LazyPrintf("transient error %s", err.Error())
	}

	return nil, false, err
}

func textSearchStream(ctx context.Context, url string, body []byte, cb func([]*protocol.FileMatch)) (bool, error) {
	req, err := http.NewRequest("GET", url, bytes.NewReader(body))
	if err != nil {
		return false, err
	}
	req = req.WithContext(ctx)

	req, ht := nethttp.TraceRequest(ot.GetTracer(ctx), req,
		nethttp.OperationName("Searcher Client"),
		nethttp.ClientTrace(false))
	defer ht.Finish()

	// Do not lose the context returned by TraceRequest
	ctx = req.Context()

	resp, err := searchDoer.Do(req)
	if err != nil {
		// If we failed due to cancellation or timeout (with no partial results in the response
		// body), return just that.
		if ctx.Err() != nil {
			err = ctx.Err()
		}
		return false, errors.Wrap(err, "streaming searcher request failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return false, err
		}
		return false, errors.WithStack(&searcherError{StatusCode: resp.StatusCode, Message: string(body)})
	}

	var ed EventDone
	dec := StreamDecoder{
		OnMatches: cb,
		OnDone: func(e EventDone) {
			ed = e
		},
		OnUnknown: func(event []byte, _ []byte) {
			err = errors.Errorf("unknown event %q", event)
		},
	}
	if err := dec.ReadAll(resp.Body); err != nil {
		return false, err
	}
	if ed.Error != "" {
		return false, errors.New(ed.Error)
	}
	if ed.DeadlineHit {
		err = context.DeadlineExceeded
	}
	return ed.LimitHit, err
}

func textSearch(ctx context.Context, url string, body []byte) ([]*protocol.FileMatch, bool, error) {
	req, err := http.NewRequest("GET", url, bytes.NewReader(body))
	if err != nil {
		return nil, false, err
	}
	req = req.WithContext(ctx)

	req, ht := nethttp.TraceRequest(ot.GetTracer(ctx), req,
		nethttp.OperationName("Searcher Client"),
		nethttp.ClientTrace(false))
	defer ht.Finish()

	// Do not lose the context returned by TraceRequest
	ctx = req.Context()

	resp, err := searchDoer.Do(req)
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
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, false, err
		}
		return nil, false, errors.WithStack(&searcherError{StatusCode: resp.StatusCode, Message: string(body)})
	}

	r := struct {
		Matches     []*protocol.FileMatch
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
