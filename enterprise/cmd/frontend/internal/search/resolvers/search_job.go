package resolvers

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

const searchJobIDKind = "SearchJob"

// MarshalSearchJobID marshals an int64 into a Relay ID for a search job.
func MarshalSearchJobID(id int64) graphql.ID {
	return relay.MarshalID(searchJobIDKind, id)
}

var _ graphqlbackend.SearchJobResolver = &searchJobResolver{}

type searchJobResolver struct {
}

func (e *searchJobResolver) ID() graphql.ID {
	// TODO implement me
	panic("implement me")
}

func (e *searchJobResolver) Query() string {
	// TODO implement me
	panic("implement me")
}

func (e *searchJobResolver) State(ctx context.Context) string {
	// TODO implement me
	panic("implement me")
}

func (e *searchJobResolver) Creator(ctx context.Context) (*graphqlbackend.UserResolver, error) {
	// TODO implement me
	panic("implement me")
}

func (e *searchJobResolver) CreatedAt() gqlutil.DateTime {
	// TODO implement me
	panic("implement me")
}

func (e *searchJobResolver) StartedAt(ctx context.Context) (*gqlutil.DateTime, error) {
	// TODO implement me
	panic("implement me")
}

func (e *searchJobResolver) FinishedAt(ctx context.Context) (*gqlutil.DateTime, error) {
	// TODO implement me
	panic("implement me")
}

func (e *searchJobResolver) CsvURL(ctx context.Context) (*string, error) {
	// TODO implement me
	panic("implement me")
}

func (e *searchJobResolver) RepoStats(ctx context.Context) (graphqlbackend.SearchJobStatsResolver, error) {
	// TODO implement me
	panic("implement me")
}
