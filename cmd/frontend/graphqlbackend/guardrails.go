package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

type GuardrailsResolver interface {
	SnippetAttribution(ctx context.Context, args *SnippetAttributionArgs) (SnippetAttributionConnectionResolver, error)
}

type SnippetAttributionArgs struct {
	gqlutil.ConnectionArgs
	Snippet string
}

type SnippetAttributionConnectionResolver interface {
	TotalCount() int32
	LimitHit() bool
	SnippetThreshold() AttributionSnippetThresholdResolver
	PageInfo() *gqlutil.PageInfo
	Nodes() []SnippetAttributionResolver
}

type AttributionSnippetThresholdResolver interface {
	SearchPerformed() bool
	LinesLowerBound() int32
}

type SnippetAttributionResolver interface {
	RepositoryName() string
}
