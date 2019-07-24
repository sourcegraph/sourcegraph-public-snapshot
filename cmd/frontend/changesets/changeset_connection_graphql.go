package changesets

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

func (GraphQLResolver) ChangesetsForRepository(ctx context.Context, repository *graphqlbackend.RepositoryResolver, arg *graphqlutil.ConnectionArgs) (graphqlbackend.ChangesetConnection, error) {
	changesets := []*gqlChangeset{{title: "Foo"}, {title: "Bar"}}
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
