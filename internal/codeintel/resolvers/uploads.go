package resolvers

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

type UploadsServiceResolver interface {
	CommitGraph(ctx context.Context, id graphql.ID) (CodeIntelligenceCommitGraphResolver, error)
}

type CodeIntelligenceCommitGraphResolver interface {
	Stale() bool
	UpdatedAt() *gqlutil.DateTime
}
