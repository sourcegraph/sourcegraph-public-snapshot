package resolvers

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

const exhaustiveSearchRepoIDKind = "ExhaustiveSearchRepo"

func MarshalExhaustiveSearchRepoID(id int64) graphql.ID {
	return relay.MarshalID(exhaustiveSearchRepoIDKind, id)
}

var _ graphqlbackend.ExhaustiveSearchRepoResolver = &exhaustiveSearchRepoResolver{}

type exhaustiveSearchRepoResolver struct {
}

func (e *exhaustiveSearchRepoResolver) ID() graphql.ID {

	//TODO implement me
	panic("implement me")
}

func (e *exhaustiveSearchRepoResolver) State() string {
	//TODO implement me
	panic("implement me")
}

func (e *exhaustiveSearchRepoResolver) Repository(ctx context.Context) *graphqlbackend.RepositoryResolver {
	//TODO implement me
	panic("implement me")
}

func (e *exhaustiveSearchRepoResolver) CreatedAt() gqlutil.DateTime {
	//TODO implement me
	panic("implement me")
}

func (e *exhaustiveSearchRepoResolver) StartedAt(ctx context.Context) (*gqlutil.DateTime, error) {
	//TODO implement me
	panic("implement me")
}

func (e *exhaustiveSearchRepoResolver) FinishedAt(ctx context.Context) (*gqlutil.DateTime, error) {
	//TODO implement me
	panic("implement me")
}

func (e *exhaustiveSearchRepoResolver) FailureMessage() *string {
	//TODO implement me
	panic("implement me")
}

func (e *exhaustiveSearchRepoResolver) Revisions(ctx context.Context, args *graphqlbackend.ExhaustiveSearchRepoRevisionsArgs) (graphqlbackend.ExhaustiveSearchRepoRevisionConnectionResolver, error) {
	//TODO implement me
	panic("implement me")
}
