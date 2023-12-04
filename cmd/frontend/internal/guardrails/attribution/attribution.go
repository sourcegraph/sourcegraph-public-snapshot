package attribution

import (
	"context"
	"fmt"
	"sync"

	"github.com/sourcegraph/conc/pool"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/guardrails/dotcom"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/client"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// ServiceOpts configures Service.
type ServiceOpts struct {
	// SearchClient is used to find attribution on the local instance.
	SearchClient client.SearchClient

	// SourcegraphDotComClient is a graphql client that is queried if
	// federating out to sourcegraph.com is enabled.
	SourcegraphDotComClient dotcom.Client

	// SourcegraphDotComFederate is true if this instance should also federate
	// to sourcegraph.com.
	SourcegraphDotComFederate bool
}

// Service is for the attribution service which searches for matches on
// snippets of code.
//
// Use NewService to construct this value.
type Service struct {
	ServiceOpts

	operations *operations
}

// NewService returns a service configured with observationCtx.
//
// Note: this registers metrics so should only be called once with the same
// observationCtx.
func NewService(observationCtx *observation.Context, opts ServiceOpts) *Service {
	return &Service{
		operations:  newOperations(observationCtx),
		ServiceOpts: opts,
	}
}

// SnippetAttributions is holds the collection of attributions for a snippet.
type SnippetAttributions struct {
	// RepositoryNames is the list of repository names. We intend on mixing
	// names from both the local instance as well as from sourcegraph.com. So
	// we intentionally use a string since the name may not represent a
	// repository available on this instance.
	//
	// Note: for now this is a simple slice, we likely will expand what is
	// represented here and it will change into a struct capturing more
	// information.
	RepositoryNames []string

	// TotalCount is the total number of repository attributions we found
	// before stopping the search.
	//
	// Note: if we didn't finish searching the full corpus then LimitHit will
	// be true. For filtering use case this means if LimitHit is true you need
	// to be conservative with TotalCount and assume it could be higher.
	TotalCount int

	// LimitHit is true if we stopped searching before looking into the full
	// corpus. If LimitHit is true then it is possible there are more than
	// TotalCount attributions.
	LimitHit bool
}

// SnippetAttribution will search the instances indexed code for code matching
// snippet and return the attribution results.
func (c *Service) SnippetAttribution(ctx context.Context, snippet string, limit int) (result *SnippetAttributions, err error) {
	ctx, traceLogger, endObservation := c.operations.snippetAttribution.With(ctx, &err, observation.Args{
		Attrs: []attribute.KeyValue{
			attribute.Int("snippet.len", len(snippet)),
			attribute.Int("limit", limit),
		},
	})
	defer endObservationWithResult(traceLogger, endObservation, &result)()

	limitHitErr := errors.New("limit hit error")
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	// we massage results in this function and possibly cancel if we can stop
	// looking.
	truncateAtLimit := func(result *SnippetAttributions) {
		if result == nil {
			return
		}
		if limit <= len(result.RepositoryNames) {
			result.LimitHit = true
			result.RepositoryNames = result.RepositoryNames[:limit]
		}
		if result.LimitHit {
			cancel(limitHitErr)
		}
	}

	// TODO(keegancsmith) how should we handle partial errors?
	p := pool.New().WithContext(ctx).WithCancelOnError().WithFirstError()

	//  We don't use NewWithResults since we want local results to come before dotcom
	var local, dotcom *SnippetAttributions

	p.Go(func(ctx context.Context) error {
		var err error
		local, err = c.snippetAttributionLocal(ctx, snippet, limit)
		truncateAtLimit(local)
		return err
	})

	if c.SourcegraphDotComFederate {
		p.Go(func(ctx context.Context) error {
			var err error
			dotcom, err = c.snippetAttributionDotCom(ctx, snippet, limit)
			truncateAtLimit(dotcom)
			return err
		})
	}

	if err := p.Wait(); err != nil && context.Cause(ctx) != limitHitErr {
		return nil, err
	}

	var agg SnippetAttributions
	seen := map[string]struct{}{}
	for _, result := range []*SnippetAttributions{local, dotcom} {
		if result == nil {
			continue
		}

		// Limitation: We just add to TotalCount even though that may mean we
		// overcount (both dotcom and local instance have the repo)
		agg.TotalCount += result.TotalCount
		agg.LimitHit = agg.LimitHit || result.LimitHit
		for _, name := range result.RepositoryNames {
			if _, ok := seen[name]; ok {
				// We have already counted this repo in the above TotalCount
				// increment, so undo that.
				agg.TotalCount--
				continue
			}
			seen[name] = struct{}{}
			agg.RepositoryNames = append(agg.RepositoryNames, name)
		}
	}

	// we call truncateAtLimit on the aggregated result to ensure we only
	// return upto limit. Note this function will call cancel but that is fine
	// since we just return after this.
	truncateAtLimit(&agg)

	return &agg, nil
}

func (c *Service) snippetAttributionLocal(ctx context.Context, snippet string, limit int) (result *SnippetAttributions, err error) {
	ctx, traceLogger, endObservation := c.operations.snippetAttributionLocal.With(ctx, &err, observation.Args{})
	defer endObservationWithResult(traceLogger, endObservation, &result)()

	const (
		version    = "V3"
		searchMode = search.Precise
		protocol   = search.Streaming
	)

	patternType := "literal"
	searchQuery := fmt.Sprintf("type:file select:repo index:only case:yes count:%d content:%q", limit, snippet)

	inputs, err := c.SearchClient.Plan(
		ctx,
		version,
		&patternType,
		searchQuery,
		searchMode,
		protocol,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create search plan")
	}

	// TODO(keegancsmith) Reading the SearchClient code it seems to miss out
	// on some of the observability that we instead add in at a later stage.
	// For example the search dataset in honeycomb will be missing. Will have
	// to follow-up with observability and maybe solve it for all users.
	//
	// Note: In our current API we could just store repo names in seen. But it
	// is safer to rely on searches ranking for result stability than doing
	// something like sorting by name from the map.
	var (
		mu        sync.Mutex
		seen      = map[api.RepoID]struct{}{}
		repoNames []string
		limitHit  bool
	)
	_, err = c.SearchClient.Execute(ctx, streaming.StreamFunc(func(ev streaming.SearchEvent) {
		mu.Lock()
		defer mu.Unlock()

		limitHit = limitHit || ev.Stats.IsLimitHit

		for _, m := range ev.Results {
			repo := m.RepoName()
			if _, ok := seen[repo.ID]; ok {
				continue
			}
			seen[repo.ID] = struct{}{}
			repoNames = append(repoNames, string(repo.Name))
		}
	}), inputs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute search")
	}

	// Note: Our search API is missing total count internally, but Zoekt does
	// expose this. For now we just count what we found.
	totalCount := len(repoNames)
	if len(repoNames) > limit {
		repoNames = repoNames[:limit]
	}

	return &SnippetAttributions{
		RepositoryNames: repoNames,
		TotalCount:      totalCount,
		LimitHit:        limitHit,
	}, nil
}

func (c *Service) snippetAttributionDotCom(ctx context.Context, snippet string, limit int) (result *SnippetAttributions, err error) {
	ctx, traceLogger, endObservation := c.operations.snippetAttributionDotCom.With(ctx, &err, observation.Args{})
	defer endObservationWithResult(traceLogger, endObservation, &result)()

	resp, err := dotcom.SnippetAttribution(ctx, c.SourcegraphDotComClient, snippet, limit)
	if err != nil {
		return nil, err
	}

	var repoNames []string
	for _, node := range resp.SnippetAttribution.Nodes {
		repoNames = append(repoNames, node.RepositoryName)
	}

	return &SnippetAttributions{
		RepositoryNames: repoNames,
		TotalCount:      resp.SnippetAttribution.TotalCount,
		LimitHit:        resp.SnippetAttribution.LimitHit,
	}, nil
}
