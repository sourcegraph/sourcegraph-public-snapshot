package attribution

import (
	"context"

	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/guardrails/dotcom"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search/client"
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

	result, err = c.snippetAttributionDotCom(ctx, snippet, limit)
	if err != nil {
		return nil, err
	}
	return result, nil
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
