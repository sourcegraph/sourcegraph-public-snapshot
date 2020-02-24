package resolvers

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	// ee "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
)

const actionJobIDKind = "ActionJob"

func marshalActionJobID(id int64) graphql.ID {
	return relay.MarshalID(actionJobIDKind, id)
}

func unmarshalActionJobID(id graphql.ID) (actionJobID int64, err error) {
	err = relay.UnmarshalSpec(id, &actionJobID)
	return
}

type actionJobResolver struct {
	job campaigns.ActionJob
}

func (r *actionJobResolver) ID() graphql.ID {
	return marshalActionJobID(r.job.ID)
}

func (r *actionJobResolver) Definition() graphqlbackend.ActionDefinitionResolver {
	return &actionDefinitionResolver{}
}

func (r *actionJobResolver) Repository(ctx context.Context) (*graphqlbackend.RepositoryResolver, error) {
	return graphqlbackend.RepositoryByIDInt32(ctx, 1)
}

func (r *actionJobResolver) BaseRevision() string {
	return "master"
}

func (r *actionJobResolver) State() campaigns.ActionJobState {
	return campaigns.ActionJobStatePending
}

func (r *actionJobResolver) Runner() graphqlbackend.RunnerResolver {
	return nil
}

func (r *actionJobResolver) BaseRepository(ctx context.Context) (*graphqlbackend.RepositoryResolver, error) {
	return r.Repository(ctx)
}
func (r *actionJobResolver) Diff() graphqlbackend.ActionJobResolver {
	return nil // return r
}
func (r *actionJobResolver) FileDiffs(ctx context.Context, args *graphqlutil.ConnectionArgs) (graphqlbackend.PreviewFileDiffConnection, error) {
	return nil, nil
}

func (r *actionJobResolver) ExecutionStart() *graphqlbackend.DateTime {
	return nil
}

func (r *actionJobResolver) ExecutionEnd() *graphqlbackend.DateTime {
	return nil
}

func (r *actionJobResolver) Log() *string {
	return nil
}
