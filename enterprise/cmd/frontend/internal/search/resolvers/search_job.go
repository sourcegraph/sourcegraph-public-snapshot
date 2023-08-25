package resolvers

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const searchJobIDKind = "SearchJob"

// MarshalSearchJobID marshals an int64 into a Relay ID for a search job.
func MarshalSearchJobID(id int64) graphql.ID {
	return relay.MarshalID(searchJobIDKind, id)
}

var _ graphqlbackend.SearchJobResolver = &searchJobResolver{}

type searchJobResolver struct {
}

func (s searchJobResolver) ID() graphql.ID {
	return "not implemented"
}

func (s searchJobResolver) Query() string {
	return "not implemented"
}

func (s searchJobResolver) State(ctx context.Context) string {
	return "QUEUED"
}

func (s searchJobResolver) Creator(ctx context.Context) (*graphqlbackend.UserResolver, error) {
	return nil, errors.New("not implemented")
}

func (s searchJobResolver) CreatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{}
}

func (s searchJobResolver) StartedAt(ctx context.Context) *gqlutil.DateTime {
	return nil
}

func (s searchJobResolver) FinishedAt(ctx context.Context) *gqlutil.DateTime {
	return nil
}

func (s searchJobResolver) URL(ctx context.Context) (*string, error) {
	return nil, errors.New("not implemented")
}

func (s searchJobResolver) RepoStats(ctx context.Context) (graphqlbackend.SearchJobStatsResolver, error) {
	return nil, errors.New("not implemented")
}
