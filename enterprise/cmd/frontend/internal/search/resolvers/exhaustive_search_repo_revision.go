package resolvers

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

const exhaustiveSearchRepoRevisionIDKind = "ExhaustiveSearchRepoRevision"

func MarshalExhaustiveSearchRepoRevisionID(id int64) graphql.ID {
	return relay.MarshalID(exhaustiveSearchRepoRevisionIDKind, id)
}

var _ graphqlbackend.ExhaustiveSearchRepoRevisionResolver = &exhaustiveSearchRepoRevisionResolver{}

type exhaustiveSearchRepoRevisionResolver struct {
}

func (e *exhaustiveSearchRepoRevisionResolver) ID() graphql.ID {
	//TODO implement me
	panic("implement me")
}

func (e *exhaustiveSearchRepoRevisionResolver) State() string {
	//TODO implement me
	panic("implement me")
}

func (e *exhaustiveSearchRepoRevisionResolver) Revision() string {
	//TODO implement me
	panic("implement me")
}

func (e *exhaustiveSearchRepoRevisionResolver) CreatedAt() gqlutil.DateTime {
	//TODO implement me
	panic("implement me")
}

func (e *exhaustiveSearchRepoRevisionResolver) StartedAt(ctx context.Context) (*gqlutil.DateTime, error) {
	//TODO implement me
	panic("implement me")
}

func (e *exhaustiveSearchRepoRevisionResolver) FinishedAt(ctx context.Context) (*gqlutil.DateTime, error) {
	//TODO implement me
	panic("implement me")
}

func (e *exhaustiveSearchRepoRevisionResolver) FailureMessage() *string {
	//TODO implement me
	panic("implement me")
}
