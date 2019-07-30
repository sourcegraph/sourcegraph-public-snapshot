package threadlike

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threadlike/internal"
)

func (GraphQLResolver) ThreadOrIssueOrChangesets(ctx context.Context, arg *graphqlutil.ConnectionArgs) (graphqlbackend.ThreadOrIssueOrChangesetConnection, error) {
	return threadlikesByOptions(ctx, internal.DBThreadsListOptions{}, arg)
}

func (GraphQLResolver) ThreadOrIssueOrChangesetsForRepository(ctx context.Context, repositoryID graphql.ID, arg *graphqlutil.ConnectionArgs) (graphqlbackend.ThreadOrIssueOrChangesetConnection, error) {
	repo, err := graphqlbackend.RepositoryByID(ctx, repositoryID)
	if err != nil {
		return nil, err
	}
	return threadlikesByOptions(ctx, internal.DBThreadsListOptions{
		RepositoryID: repo.DBID(),
	}, arg)
}

func ThreadOrIssueOrChangesetsByIDs(ctx context.Context, ids []int64, arg *graphqlutil.ConnectionArgs) (graphqlbackend.ThreadOrIssueOrChangesetConnection, error) {
	return threadlikesByOptions(ctx, internal.DBThreadsListOptions{
		ThreadIDs: ids,
	}, arg)
}

func threadlikesByOptions(ctx context.Context, options internal.DBThreadsListOptions, arg *graphqlutil.ConnectionArgs) (graphqlbackend.ThreadOrIssueOrChangesetConnection, error) {
	list, err := internal.DBThreads{}.List(ctx, options)
	if err != nil {
		return nil, err
	}
	threadlikes := make([]graphqlbackend.ThreadOrIssueOrChangeset, len(list))
	for i, a := range list {
		threadlikes[i] = newGQLThreadOrIssueOrChangeset(a)
	}
	return &threadOrIssueOrChangesetConnection{arg: arg, threadlikes: threadlikes}, nil
}

type threadOrIssueOrChangesetConnection struct {
	arg         *graphqlutil.ConnectionArgs
	threadlikes []graphqlbackend.ThreadOrIssueOrChangeset
}

func (r *threadOrIssueOrChangesetConnection) Nodes(ctx context.Context) ([]graphqlbackend.ThreadOrIssueOrChangeset, error) {
	threadlikes := r.threadlikes
	if first := r.arg.First; first != nil && len(threadlikes) > int(*first) {
		threadlikes = threadlikes[:int(*first)]
	}

	threadlikes2 := make([]graphqlbackend.ThreadOrIssueOrChangeset, len(threadlikes))
	for i, l := range threadlikes {
		threadlikes2[i] = l
	}
	return threadlikes2, nil
}

func (r *threadOrIssueOrChangesetConnection) TotalCount(ctx context.Context) (int32, error) {
	return int32(len(r.threadlikes)), nil
}

func (r *threadOrIssueOrChangesetConnection) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(r.arg.First != nil && int(*r.arg.First) < len(r.threadlikes)), nil
}
