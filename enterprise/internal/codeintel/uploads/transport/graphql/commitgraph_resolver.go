package graphql

import (
	"context"
	"time"

	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

type CommitGraphResolver struct {
	stale     bool
	updatedAt *time.Time
}

func NewCommitGraphResolver(stale bool, updatedAt *time.Time) resolverstubs.CodeIntelligenceCommitGraphResolver {
	return &CommitGraphResolver{
		stale:     stale,
		updatedAt: updatedAt,
	}
}

func (r *CommitGraphResolver) Stale(ctx context.Context) (bool, error) {
	return r.stale, nil
}

func (r *CommitGraphResolver) UpdatedAt(ctx context.Context) (*gqlutil.DateTime, error) {
	return gqlutil.DateTimeOrNil(r.updatedAt), nil
}
