package resolvers

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

const searchJobRepoRevisionIDKind = "SearchJobRepoRevision"

// MarshalSearchJobRepoRevisionID marshals an int64 into a Relay ID for a search job repo revision.
func MarshalSearchJobRepoRevisionID(id int64) graphql.ID {
	return relay.MarshalID(searchJobRepoRevisionIDKind, id)
}

var _ graphqlbackend.SearchJobRepoRevisionResolver = &searchJobRepoRevisionResolver{}

type searchJobRepoRevisionResolver struct {
}

func (e *searchJobRepoRevisionResolver) ID() graphql.ID {
	//TODO implement me
	panic("implement me")
}

func (e *searchJobRepoRevisionResolver) State() string {
	//TODO implement me
	panic("implement me")
}

func (e *searchJobRepoRevisionResolver) Revision() string {
	//TODO implement me
	panic("implement me")
}

func (e *searchJobRepoRevisionResolver) CreatedAt() gqlutil.DateTime {
	//TODO implement me
	panic("implement me")
}

func (e *searchJobRepoRevisionResolver) StartedAt(ctx context.Context) (*gqlutil.DateTime, error) {
	//TODO implement me
	panic("implement me")
}

func (e *searchJobRepoRevisionResolver) FinishedAt(ctx context.Context) (*gqlutil.DateTime, error) {
	//TODO implement me
	panic("implement me")
}

func (e *searchJobRepoRevisionResolver) FailureMessage() *string {
	//TODO implement me
	panic("implement me")
}
