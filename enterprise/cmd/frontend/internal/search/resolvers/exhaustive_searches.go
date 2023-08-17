package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

var _ graphqlbackend.ExhaustiveSearchesResolver = &exhaustiveSearchesResolver{}

type exhaustiveSearchesResolver struct {
}

func (e *exhaustiveSearchesResolver) ValidateExhaustiveSearchQuery(ctx context.Context, args *graphqlbackend.ValidateExhaustiveSearchQueryArgs) (graphqlbackend.ValidateExhaustiveSearchQueryResolver, error) {
	//TODO implement me
	panic("implement me")
}

func (e *exhaustiveSearchesResolver) CreateExhaustiveSearch(ctx context.Context, args *graphqlbackend.CreateExhaustiveSearchArgs) (graphqlbackend.ExhaustiveSearchResolver, error) {
	//TODO implement me
	panic("implement me")
}

func (e *exhaustiveSearchesResolver) CancelExhaustiveSearch(ctx context.Context, args *graphqlbackend.CancelExhaustiveSearchArgs) (*graphqlbackend.EmptyResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (e *exhaustiveSearchesResolver) DeleteExhaustiveSearch(ctx context.Context, args *graphqlbackend.DeleteExhaustiveSearchArgs) (*graphqlbackend.EmptyResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (e *exhaustiveSearchesResolver) ExhaustiveSearch(ctx context.Context, args *graphqlbackend.ExhaustiveSearchArgs) (graphqlbackend.ExhaustiveSearchResolver, error) {
	//TODO implement me
	panic("implement me")
}

func (e *exhaustiveSearchesResolver) ExhaustiveSearches(ctx context.Context, args *graphqlbackend.ExhaustiveSearchesArgs) (graphqlbackend.ExhaustiveSearchesConnectionResolver, error) {
	//TODO implement me
	panic("implement me")
}
