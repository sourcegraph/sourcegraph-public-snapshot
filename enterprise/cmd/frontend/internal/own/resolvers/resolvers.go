package resolvers

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func New() graphqlbackend.OwnResolver {
	return &resolver{}
}

type resolver struct{}

func (r *resolver) GitBlobOwnership(ctx context.Context, blob *graphqlbackend.GitTreeEntryResolver, args graphqlbackend.ListOwnershipArgs) (graphqlbackend.OwnershipConnectionResolver, error) {
	return nil, nil
}

func (r *resolver) PersonOwnerField(person *graphqlbackend.PersonResolver) string {
	return "owner"
}
func (r *resolver) UserOwnerField(user *graphqlbackend.UserResolver) string {
	return "owner"
}
func (r *resolver) TeamOwnerField(team *graphqlbackend.TeamResolver) string {
	return "owner"
}

func (r *resolver) NodeResolvers() map[string]graphqlbackend.NodeByIDFunc {
	return map[string]func(ctx context.Context, id graphql.ID) (graphqlbackend.Node, error){}
}
