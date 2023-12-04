package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/guardrails/attribution"
)

var _ graphqlbackend.GuardrailsResolver = &GuardrailsResolver{}

type GuardrailsResolver struct {
	AttributionService *attribution.Service
}

func (c *GuardrailsResolver) SnippetAttribution(ctx context.Context, args *graphqlbackend.SnippetAttributionArgs) (graphqlbackend.SnippetAttributionConnectionResolver, error) {
	limit := 5
	if args.First != nil {
		limit = int(*args.First)
	}

	result, err := c.AttributionService.SnippetAttribution(ctx, args.Snippet, limit)
	if err != nil {
		return nil, err
	}

	return snippetAttributionConnectionResolver{result: result}, nil
}

type snippetAttributionConnectionResolver struct {
	result *attribution.SnippetAttributions
}

func (c snippetAttributionConnectionResolver) TotalCount() int32 {
	return int32(c.result.TotalCount)
}
func (c snippetAttributionConnectionResolver) LimitHit() bool {
	return c.result.LimitHit
}
func (c snippetAttributionConnectionResolver) PageInfo() *graphqlutil.PageInfo {
	return graphqlutil.HasNextPage(false)
}
func (c snippetAttributionConnectionResolver) Nodes() []graphqlbackend.SnippetAttributionResolver {
	var nodes []graphqlbackend.SnippetAttributionResolver
	for _, name := range c.result.RepositoryNames {
		nodes = append(nodes, snippetAttributionResolver(name))
	}
	return nodes
}

type snippetAttributionResolver string

func (c snippetAttributionResolver) RepositoryName() string {
	return string(c)
}
