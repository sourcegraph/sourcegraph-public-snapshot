package resolvers

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

// Resolver is the GraphQL resolver of all things related to batch changes.
type Resolver struct {
	logger log.Logger
	db     database.DB
}

// New returns a new Resolver whose store uses the given database
func New(logger log.Logger, db database.DB) graphqlbackend.ExhaustiveSearchesResolver {
	return &Resolver{logger: logger, db: db}
}

var _ graphqlbackend.ExhaustiveSearchesResolver = &Resolver{}

func (r *Resolver) ValidateExhaustiveSearchQuery(ctx context.Context, args *graphqlbackend.ValidateExhaustiveSearchQueryArgs) (graphqlbackend.ValidateExhaustiveSearchQueryResolver, error) {
	//TODO implement me
	panic("implement me")
}

func (r *Resolver) CreateExhaustiveSearch(ctx context.Context, args *graphqlbackend.CreateExhaustiveSearchArgs) (graphqlbackend.ExhaustiveSearchResolver, error) {
	//TODO implement me
	panic("implement me")
}

func (r *Resolver) CancelExhaustiveSearch(ctx context.Context, args *graphqlbackend.CancelExhaustiveSearchArgs) (*graphqlbackend.EmptyResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (r *Resolver) DeleteExhaustiveSearch(ctx context.Context, args *graphqlbackend.DeleteExhaustiveSearchArgs) (*graphqlbackend.EmptyResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (r *Resolver) ExhaustiveSearch(ctx context.Context, args *graphqlbackend.ExhaustiveSearchArgs) (graphqlbackend.ExhaustiveSearchResolver, error) {
	//TODO implement me
	panic("implement me")
}

func (r *Resolver) ExhaustiveSearches(ctx context.Context, args *graphqlbackend.ExhaustiveSearchesArgs) (graphqlbackend.ExhaustiveSearchesConnectionResolver, error) {
	//TODO implement me
	panic("implement me")
}
