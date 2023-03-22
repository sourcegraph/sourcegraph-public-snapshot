package graphql

import (
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

func (r *CommitGraphResolver) Stale() bool {
	return r.stale
}

func (r *CommitGraphResolver) UpdatedAt() *gqlutil.DateTime {
	return gqlutil.DateTimeOrNil(r.updatedAt)
}
