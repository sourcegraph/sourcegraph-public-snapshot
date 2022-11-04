package graphql

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

type CommitGraphResolver struct {
	stale     bool
	updatedAt *time.Time
}

func NewCommitGraphResolver(stale bool, updatedAt *time.Time) *CommitGraphResolver {
	return &CommitGraphResolver{
		stale:     stale,
		updatedAt: updatedAt,
	}
}

func (r *CommitGraphResolver) Stale(ctx context.Context) (bool, error) {
	return r.stale, nil
}

func (r *CommitGraphResolver) UpdatedAt(ctx context.Context) (*graphqlbackend.DateTime, error) {
	return graphqlbackend.DateTimeOrNil(r.updatedAt), nil
}
