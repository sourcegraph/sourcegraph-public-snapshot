package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

type GuardrailsResolver interface {
	SnippetAttribution(ctx context.Context, args *SnippetAttributionArgs) (SnippetAttributionConnectionResolver, error)
}

type SnippetAttributionArgs struct {
	graphqlutil.ConnectionArgs
	Snippet string
}

type SnippetAttributionConnectionResolver interface {
	TotalCount() int32
	LimitHit() bool
	PageInfo() *graphqlutil.PageInfo
	Nodes() []SnippetAttributionResolver
}

type SnippetAttributionResolver interface {
	RepositoryName() string
}
