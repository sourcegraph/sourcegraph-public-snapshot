package threads

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

func (GraphQLResolver) ThreadsForRepository(ctx context.Context, repositoryID graphql.ID, arg *graphqlutil.ConnectionArgs) (graphqlbackend.ThreadConnection, error) {
	repo, err := graphqlbackend.RepositoryByID(ctx, repositoryID)
	if err != nil {
		return nil, err
	}

	list, err := dbThreads{}.List(ctx, dbThreadsListOptions{
		RepositoryID: repo.DBID(),
	})
	if err != nil {
		return nil, err
	}
	threads := make([]*gqlThread, len(list))
	for i, a := range list {
		threads[i] = &gqlThread{db: a}
	}
	return &threadConnection{arg: arg, threads: threads}, nil
}

type threadConnection struct {
	arg     *graphqlutil.ConnectionArgs
	threads []*gqlThread
}

func (r *threadConnection) Nodes(ctx context.Context) ([]graphqlbackend.Thread, error) {
	threads := r.threads
	if first := r.arg.First; first != nil && len(threads) > int(*first) {
		threads = threads[:int(*first)]
	}

	threads2 := make([]graphqlbackend.Thread, len(threads))
	for i, l := range threads {
		threads2[i] = l
	}
	return threads2, nil
}

func (r *threadConnection) TotalCount(ctx context.Context) (int32, error) {
	return int32(len(r.threads)), nil
}

func (r *threadConnection) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(r.arg.First != nil && int(*r.arg.First) < len(r.threads)), nil
}
