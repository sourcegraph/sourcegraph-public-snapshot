package graphqlbackend

import (
	"context"
	"errors"

	graphql "github.com/neelance/graphql-go"
	"github.com/neelance/graphql-go/relay"

	"sourcegraph.com/sourcegraph/sourcegraph/api"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/internal/localstore"
)

var GraphQLSchema *graphql.Schema

func init() {
	var err error
	GraphQLSchema, err = graphql.ParseSchema(api.Schema, &queryResolver{})
	if err != nil {
		panic(err)
	}
}

type nodeResolver interface {
	ID() graphql.ID
	ToRepository() (*repositoryResolver, bool)
	ToCommit() (*commitResolver, bool)
}

type nodeBase struct{}

func (*nodeBase) ToRepository() (*repositoryResolver, bool) {
	return nil, false
}

func (*nodeBase) ToCommit() (*commitResolver, bool) {
	return nil, false
}

type queryResolver struct{}

func (r *queryResolver) Root() *rootResolver {
	return &rootResolver{}
}

func (r *queryResolver) Node(ctx context.Context, args *struct{ ID graphql.ID }) (nodeResolver, error) {
	switch relay.UnmarshalKind(args.ID) {
	case "Repository":
		return repositoryByID(ctx, args.ID)
	case "Commit":
		return commitByID(ctx, args.ID)
	default:
		return nil, errors.New("invalid id")
	}
}

type rootResolver struct{}

func (r *rootResolver) Repository(ctx context.Context, args *struct{ URI string }) (*repositoryResolver, error) {
	repo, err := localstore.Repos.GetByURI(ctx, args.URI)
	if err != nil {
		return nil, err
	}
	return &repositoryResolver{repo: repo}, nil
}
