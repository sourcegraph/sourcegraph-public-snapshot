package resolvers

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

const searchJobRepoIDKind = "SearchJobRepo"

// MarshalSearchJobRepoID marshals an int64 into a Relay ID for a search job repo.
func MarshalSearchJobRepoID(id int64) graphql.ID {
	return relay.MarshalID(searchJobRepoIDKind, id)
}

var _ graphqlbackend.SearchJobRepoResolver = &searchJobRepoResolver{}

type searchJobRepoResolver struct {
}

func (e *searchJobRepoResolver) ID() graphql.ID {

	//TODO implement me
	panic("implement me")
}

func (e *searchJobRepoResolver) State() string {
	//TODO implement me
	panic("implement me")
}

func (e *searchJobRepoResolver) Repository(ctx context.Context) *graphqlbackend.RepositoryResolver {
	//TODO implement me
	panic("implement me")
}

func (e *searchJobRepoResolver) CreatedAt() gqlutil.DateTime {
	//TODO implement me
	panic("implement me")
}

func (e *searchJobRepoResolver) StartedAt(ctx context.Context) (*gqlutil.DateTime, error) {
	//TODO implement me
	panic("implement me")
}

func (e *searchJobRepoResolver) FinishedAt(ctx context.Context) (*gqlutil.DateTime, error) {
	//TODO implement me
	panic("implement me")
}

func (e *searchJobRepoResolver) FailureMessage() *string {
	//TODO implement me
	panic("implement me")
}

func (e *searchJobRepoResolver) Revisions(ctx context.Context, args *graphqlbackend.SearchJobRepoRevisionsArgs) (graphqlbackend.SearchJobRepoRevisionConnectionResolver, error) {
	//TODO implement me
	panic("implement me")
}
