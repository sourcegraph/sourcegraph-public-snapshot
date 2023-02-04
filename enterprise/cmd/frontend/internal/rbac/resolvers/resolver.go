package resolvers

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

// Resolver is the GraphQL resolver of all things related to batch changes.
type Resolver struct {
	logger log.Logger
	db     database.DB
}

func New(logger log.Logger, db database.DB) gql.RBACResolver {
	return &Resolver{logger: logger, db: db}
}

func (r *Resolver) NodeResolvers() map[string]graphqlbackend.NodeByIDFunc {
	return map[string]graphqlbackend.NodeByIDFunc{
		roleIDKind: func(ctx context.Context, id graphql.ID) (gql.Node, error) {
			return r.roleByID(ctx, id)
		},
		permissionIDKind: func(ctx context.Context, id graphql.ID) (gql.Node, error) {
			return r.permissionByID(ctx, id)
		},
	}
}
