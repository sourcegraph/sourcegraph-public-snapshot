package search

import (
	"context"
	"regexp/syntax"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/zoekt"
	zoektquery "github.com/google/zoekt/query"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/comby"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/backend"
	zoektutil "github.com/sourcegraph/sourcegraph/internal/search/zoekt"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	zoektOnce   sync.Once
	endpointMap atomicEndpoints
	zoektClient zoekt.Streamer
)

func getZoektClient(indexerEndpoints []string) zoekt.Streamer {
	zoektOnce.Do(func() {
		zoektClient = backend.NewMeteredSearcher(
			"", // no hostname means its the aggregator
			&backend.HorizontalSearcher{
				Map:  &endpointMap,
				Dial: backend.ZoektDial,
			},
		)
	})
	endpointMap.Set(indexerEndpoints)
	return zoektClient
}

func HandleFilePathPatterns(query *search.TextPatternInfo) (zoektquery.Q, error) {
	var and []zoektquery.Q

	// Zoekt uses regular expressions for file paths.
	// Unhandled cases: PathPatternsAreCaseSensitive and whitespace in file path patterns.
	for _, p := range query.IncludePatterns {
		q, err := zoektutil.FileRe(p, query.IsCaseSensitive)
		if err != nil {
			return nil, err
		}
		and = append(and, q)
	}
	if query.ExcludePattern != "" {
		q, err := zoektutil.FileRe(query.ExcludePattern, query.IsCaseSensitive)
		if err != nil {
			return nil, err
		}
		and = append(and, &zoektquery.Not{Child: q})
	}

	// For conditionals that happen on a repo we can use type:repo queries. eg
	// (type:repo file:foo) (type:repo file:bar) will match all repos which
	// contain a filename matching "foo" and a filename matchinb "bar".
	//
	// Note: (type:repo file:foo file:bar) will only find repos with a
	// filename containing both "foo" and "bar".
	for _, p := range query.FilePatternsReposMustInclude {
		q, err := zoektutil.FileRe(p, query.IsCaseSensitive)
		if err != nil {
			return nil, err
		}
		and = append(and, &zoektquery.Type{Type: zoektquery.TypeRepo, Child: q})
	}
	for _, p := range query.FilePatternsReposMustExclude {
		q, err := zoektutil.FileRe(p, query.IsCaseSensitive)
		if err != nil {
			return nil, err
		}
		and = append(and, &zoektquery.Not{Child: &zoektquery.Type{Type: zoektquery.TypeRepo, Child: q}})
	}

	return zoektquery.NewAnd(and...), nil
}

func buildQuery(args *search.TextPatternInfo, branchRepos []zoektquery.BranchRepos, filePathPatterns zoektquery.Q, shortcircuit bool) (zoektquery.Q, error) {
	regexString := comby.StructuralPatToRegexpQuery(args.Pattern, shortcircuit)
	if len(regexString) == 0 {
		return &zoektquery.Const{Value: true}, nil
	}
	re, err := syntax.Parse(regexString, syntax.ClassNL|syntax.PerlX|syntax.UnicodeGroups)
	if err != nil {
		return nil, err
	}
	return zoektquery.NewAnd(
		&zoektquery.BranchesRepos{List: branchRepos},
		filePathPatterns,
		&zoektquery.Regexp{
			Regexp:        re,
			CaseSensitive: true,
			Content:       true,
		},
	), nil
}

// zoektSearch searches repositories using zoekt, returning file contents for
// files that match the given pattern.
//
// Timeouts are reported through the context, and as a special case errNoResultsInTimeout
// is returned if no results are found in the given timeout (instead of the more common
// case of finding partial or full results in the given timeout).
func zoektSearch(ctx context.Context, args *search.TextPatternInfo, branchRepos []zoektquery.BranchRepos, endpoints []string, useFullDeadline bool, matches chan []zoekt.FileMatch) (err error) {
	if len(branchRepos) == 0 {
		return nil
	}

	numRepos := 0
	for _, br := range branchRepos {
		numRepos += int(br.Repos.GetCardinality())
	}

	// Choose sensible values for k when we generalize this.
	k := zoektutil.ResultCountFactor(numRepos, args.FileMatchLimit, false)
	searchOpts := zoektutil.SearchOpts(ctx, k, args.FileMatchLimit, nil)
	searchOpts.Whole = true

	// TODO(@camdencheek) TODO(@rvantonder) handle "timeout:..." values in this context.
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
		var cancel context.CancelFunc
		ctx, cancel = contextWithoutDeadline(ctx)
		defer cancel()
	}

	filePathPatterns, err := HandleFilePathPatterns(args)
	if err != nil {
		return err
	}

	q, err := buildQuery(args, branchRepos, filePathPatterns, false)
	if err != nil {
		return err
	}

	client := getZoektClient(endpoints)
	return client.StreamSearch(ctx, q, &searchOpts, backend.ZoektStreamFunc(func(sr *zoekt.SearchResult) {
		matches <- sr.Files
	}))
}

// contextWithoutDeadline returns a context which will cancel if the cOld is
// canceled.
func contextWithoutDeadline(cOld context.Context) (context.Context, context.CancelFunc) {
	cNew, cancel := context.WithCancel(context.Background())

	// Set trace context so we still get spans propagated
	cNew = trace.CopyContext(cNew, cOld)

	// Copy actor from cOld to cNew.
	cNew = actor.WithActor(cNew, actor.FromContext(cOld))

	go func() {
		select {
		case <-cOld.Done():
			// cancel the new context if the old one is done for some reason other than the deadline passing.
			if cOld.Err() != context.DeadlineExceeded {
				cancel()
			}
		case <-cNew.Done():
		}
	}()

	return cNew, cancel
}

// atomicEndpoints allows us to update the endpoints used by our zoekt client.
type atomicEndpoints struct {
	endpoints atomic.Value
}

func (a *atomicEndpoints) Endpoints() ([]string, error) {
	eps := a.endpoints.Load()
	if eps == nil {
		return nil, errors.New("endpoints have not been set")
	}
	return eps.([]string), nil
}

func (a *atomicEndpoints) Set(endpoints []string) {
	if !a.needsUpdate(endpoints) {
		return
	}
	a.endpoints.Store(endpoints)
}

func (a *atomicEndpoints) needsUpdate(endpoints []string) bool {
	old, err := a.Endpoints()
	if err != nil {
		return true
	}
	if len(old) != len(endpoints) {
		return true
	}

	for i := range endpoints {
		if old[i] != endpoints[i] {
			return true
		}
	}

	return false
}
