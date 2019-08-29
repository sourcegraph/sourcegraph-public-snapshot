package commitstatuses

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func (GraphQLResolver) CommitStatusContextByID(ctx context.Context, id graphql.ID) (graphqlbackend.CommitStatusContext, error) {
	dbID, err := graphqlbackend.UnmarshalCommitStatusContextID(id)
	if err != nil {
		return nil, err
	}

	// TODO!(sqs): check perms
	db, err := dbCommitStatusContexts{}.GetByID(ctx, dbID)
	if err != nil {
		return nil, err
	}
	return &gqlCommitStatusContext{*db}, nil
}

type gqlCommitStatusContext struct{ db dbCommitStatusContext }

func (v *gqlCommitStatusContext) ID() graphql.ID {
	return graphqlbackend.MarshalCommitStatusContextID(v.db.ID)
}

func (v *gqlCommitStatusContext) DBID() int64 { return v.db.ID }

func (v *gqlCommitStatusContext) Repository(ctx context.Context) (*graphqlbackend.RepositoryResolver, error) {
	return graphqlbackend.RepositoryByDBID(ctx, v.db.RepositoryID)
}

func (v *gqlCommitStatusContext) Commit(ctx context.Context) (*graphqlbackend.GitCommitResolver, error) {
	repository, err := v.Repository(ctx)
	if err != nil {
		return nil, err
	}
	return graphqlbackend.GetGitCommit(ctx, repository, graphqlbackend.GitObjectID(v.db.CommitOID))
}

func (v *gqlCommitStatusContext) Context() string { return v.db.Context }

func (v *gqlCommitStatusContext) State() graphqlbackend.CommitStatusState {
	return graphqlbackend.CommitStatusState(v.db.State)
}

func (v *gqlCommitStatusContext) Description() *string { return v.db.Description }

func (v *gqlCommitStatusContext) TargetURL() *string { return v.db.TargetURL }

func (v *gqlCommitStatusContext) Actor(ctx context.Context) (*graphqlbackend.Actor, error) {
	return v.db.Actor.GQL(ctx)
}

func (v *gqlCommitStatusContext) CreatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{v.db.CreatedAt}
}
