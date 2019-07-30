package changesets

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threadlike/internal"
)

func (GraphQLResolver) Changesets(ctx context.Context, arg *graphqlutil.ConnectionArgs) (graphqlbackend.ChangesetConnection, error) {
	return changesetsByOptions(ctx, internal.DBThreadsListOptions{}, arg)
}

func (GraphQLResolver) ChangesetsForRepository(ctx context.Context, repositoryID graphql.ID, arg *graphqlutil.ConnectionArgs) (graphqlbackend.ChangesetConnection, error) {
	repo, err := graphqlbackend.RepositoryByID(ctx, repositoryID)
	if err != nil {
		return nil, err
	}
	return changesetsByOptions(ctx, internal.DBThreadsListOptions{
		RepositoryID: repo.DBID(),
	}, arg)
}

func changesetsByOptions(ctx context.Context, options internal.DBThreadsListOptions, arg *graphqlutil.ConnectionArgs) (graphqlbackend.ChangesetConnection, error) {
	list, err := internal.DBThreads{}.List(ctx, options)
	if err != nil {
		return nil, err
	}
	changesets := make([]*gqlChangeset, len(list))
	for i, a := range list {
		changesets[i] = newGQLChangeset(a)
	}
	return &changesetConnection{arg: arg, changesets: changesets}, nil
}

type changesetConnection struct {
	arg        *graphqlutil.ConnectionArgs
	changesets []*gqlChangeset
}

func (r *changesetConnection) Nodes(ctx context.Context) ([]graphqlbackend.Changeset, error) {
	changesets := r.changesets
	if first := r.arg.First; first != nil && len(changesets) > int(*first) {
		changesets = changesets[:int(*first)]
	}

	changesets2 := make([]graphqlbackend.Changeset, len(changesets))
	for i, l := range changesets {
		changesets2[i] = l
	}
	return changesets2, nil
}

func (r *changesetConnection) TotalCount(ctx context.Context) (int32, error) {
	return int32(len(r.changesets)), nil
}

func (r *changesetConnection) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(r.arg.First != nil && int(*r.arg.First) < len(r.changesets)), nil
}
