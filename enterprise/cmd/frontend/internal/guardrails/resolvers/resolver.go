package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/guardrails/attribution"
)

var _ graphqlbackend.GuardrailsResolver = &GuardrailsResolver{}

type GuardrailsResolver struct {
	AttributionService *attribution.Service
}

func (c *GuardrailsResolver) SnippetAttribution(ctx context.Context, args *graphqlbackend.SnippetAttributionArgs) (*graphqlbackend.SnippetAttributionConnectionResolver, error) {
	limit := 5
	if args.First != nil {
		limit = int(*args.First)
	}

	result, err := c.AttributionService.SnippetAttribution(ctx, args.Snippet, limit)
	if err != nil {
		return nil, err
	}

	var nodes []graphqlbackend.SnippetAttributionResolver
	for _, name := range result.RepositoryNames {
		nodes = append(nodes, graphqlbackend.SnippetAttributionResolver{
			RepositoryName: name,
		})
	}

	return &graphqlbackend.SnippetAttributionConnectionResolver{
		TotalCount: result.TotalCount,
		LimitHit:   result.LimitHit,
		PageInfo:   graphqlutil.HasNextPage(false),
		Nodes:      nodes,
	}, nil
}
