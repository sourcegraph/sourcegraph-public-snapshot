package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

func (r *RepositoryResolver) Thread(ctx context.Context, arg struct{ Number string }) (Thread, error) {
	return ThreadInRepository(ctx, r.ID(), arg.Number)
}

func (r *RepositoryResolver) Threads(ctx context.Context, arg *graphqlutil.ConnectionArgs) (ThreadConnection, error) {
	return ThreadsForRepository(ctx, r.ID(), arg)
}

func (r *RepositoryResolver) Changeset(ctx context.Context, arg struct{ Number string }) (Changeset, error) {
	return ChangesetInRepository(ctx, r.ID(), arg.Number)
}

func (r *RepositoryResolver) Changesets(ctx context.Context, arg *graphqlutil.ConnectionArgs) (ChangesetConnection, error) {
	return ChangesetsForRepository(ctx, r.ID(), arg)
}

func (r *RepositoryResolver) ThreadOrIssueOrChangeset(ctx context.Context, arg struct{ Number string }) (*ThreadOrIssueOrChangeset, error) {
	return ThreadOrIssueOrChangesetInRepository(ctx, r.ID(), arg.Number)
}

func (r *RepositoryResolver) ThreadOrIssueOrChangesets(ctx context.Context, arg *graphqlutil.ConnectionArgs) (ThreadOrIssueOrChangesetConnection, error) {
	return ThreadOrIssueOrChangesetsForRepository(ctx, r.ID(), arg)
}
