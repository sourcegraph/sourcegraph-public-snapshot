package graphqlbackend

import (
	"context"
)

func (r *RepositoryResolver) Thread(ctx context.Context, arg struct{ Number string }) (Thread, error) {
	return ThreadInRepository(ctx, r.ID(), arg.Number)
}

func (r *RepositoryResolver) Threads(ctx context.Context, arg *ThreadConnectionArgs) (ThreadConnection, error) {
	return ThreadsForRepository(ctx, r.ID(), arg)
}
