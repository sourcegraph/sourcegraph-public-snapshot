package graphql

import (
	"context"
	"time"
)

type CodeIntelligenceCommitGraphResolver interface {
	Stale(ctx context.Context) (bool, error)
	UpdatedAt(ctx context.Context) (*DateTime, error)
}

type CommitGraphResolver struct {
	stale     bool
	updatedAt *time.Time
}

func NewCommitGraphResolver(stale bool, updatedAt *time.Time) CodeIntelligenceCommitGraphResolver {
	return &CommitGraphResolver{
		stale:     stale,
		updatedAt: updatedAt,
	}
}

func (r *CommitGraphResolver) Stale(ctx context.Context) (bool, error) {
	return r.stale, nil
}

func (r *CommitGraphResolver) UpdatedAt(ctx context.Context) (*DateTime, error) {
	return DateTimeOrNil(r.updatedAt), nil
}
