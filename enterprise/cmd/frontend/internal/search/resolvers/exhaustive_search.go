package resolvers

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

var _ graphqlbackend.ExhaustiveSearchResolver = &exhaustiveSearchResolver{}

type exhaustiveSearchResolver struct {
}

func (e *exhaustiveSearchResolver) ID() graphql.ID {
	//TODO implement me
	panic("implement me")
}

func (e *exhaustiveSearchResolver) Query() string {
	//TODO implement me
	panic("implement me")
}

func (e *exhaustiveSearchResolver) State(ctx context.Context) string {
	//TODO implement me
	panic("implement me")
}

func (e *exhaustiveSearchResolver) Creator(ctx context.Context) (*graphqlbackend.UserResolver, error) {
	//TODO implement me
	panic("implement me")
}

func (e *exhaustiveSearchResolver) CreatedAt() gqlutil.DateTime {
	//TODO implement me
	panic("implement me")
}

func (e *exhaustiveSearchResolver) StartedAt(ctx context.Context) (*gqlutil.DateTime, error) {
	//TODO implement me
	panic("implement me")
}

func (e *exhaustiveSearchResolver) FinishedAt(ctx context.Context) (*gqlutil.DateTime, error) {
	//TODO implement me
	panic("implement me")
}

func (e *exhaustiveSearchResolver) CsvURL(ctx context.Context) (*string, error) {
	//TODO implement me
	panic("implement me")
}

func (e *exhaustiveSearchResolver) RepoStats(ctx context.Context) (*graphqlbackend.ExhaustiveSearchStatsResolver, error) {
	//TODO implement me
	panic("implement me")
}

func (e *exhaustiveSearchResolver) Repositories(ctx context.Context, args *graphqlutil.ConnectionArgs) (graphqlbackend.ExhaustiveSearchRepoConnectionResolver, error) {
	//TODO implement me
	panic("implement me")
}
