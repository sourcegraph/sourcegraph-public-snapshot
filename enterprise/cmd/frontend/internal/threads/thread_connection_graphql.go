package threads

import (
	"context"
	"sync"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

func (GraphQLResolver) Threads(ctx context.Context, arg *graphqlbackend.ThreadConnectionArgs) (graphqlbackend.ThreadConnection, error) {
	return &threadConnection{opt: threadConnectionArgsToListOptions(arg)}, nil
}

func (GraphQLResolver) ThreadsForRepository(ctx context.Context, repositoryID graphql.ID, arg *graphqlbackend.ThreadConnectionArgs) (graphqlbackend.ThreadConnection, error) {
	repo, err := graphqlbackend.RepositoryByID(ctx, repositoryID)
	if err != nil {
		return nil, err
	}
	opt := threadConnectionArgsToListOptions(arg)
	opt.RepositoryID = repo.DBID()
	return &threadConnection{opt: opt}, nil
}

func ThreadsByIDs(threadIDs []int64, arg *graphqlbackend.ThreadConnectionArgs) graphqlbackend.ThreadConnection {
	opt := threadConnectionArgsToListOptions(arg)
	opt.ThreadIDs = threadIDs
	return &threadConnection{opt: opt}
}

func threadConnectionArgsToListOptions(arg *graphqlbackend.ThreadConnectionArgs) dbThreadsListOptions {
	var opt dbThreadsListOptions
	arg.Set(&opt.LimitOffset)
	if arg.Open != nil && *arg.Open {
		opt.State = string(graphqlbackend.ThreadStateOpen)
	}
	return opt
}

type threadConnection struct {
	opt dbThreadsListOptions

	once    sync.Once
	threads []*dbThread
	err     error
}

func (r *threadConnection) compute(ctx context.Context) ([]*dbThread, error) {
	r.once.Do(func() {
		opt2 := r.opt
		if opt2.LimitOffset != nil {
			tmp := *opt2.LimitOffset
			opt2.LimitOffset = &tmp
			opt2.Limit++ // so we can detect if there is a next page
		}

		r.threads, r.err = dbThreads{}.List(ctx, opt2)
	})
	return r.threads, r.err
}

func (r *threadConnection) Nodes(ctx context.Context) ([]graphqlbackend.Thread, error) {
	dbThreads, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	if r.opt.LimitOffset != nil && len(dbThreads) > r.opt.LimitOffset.Limit {
		dbThreads = dbThreads[:r.opt.LimitOffset.Limit]
	}

	threads := make([]graphqlbackend.Thread, len(dbThreads))
	for i, dbThread := range dbThreads {
		threads[i] = newGQLThread(dbThread)
	}
	return threads, nil
}

func (r *threadConnection) TotalCount(ctx context.Context) (int32, error) {
	count, err := dbThreads{}.Count(ctx, r.opt)
	return int32(count), err
}

func (r *threadConnection) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	threads, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	return graphqlutil.HasNextPage(r.opt.LimitOffset != nil && len(threads) > r.opt.Limit), nil
}
