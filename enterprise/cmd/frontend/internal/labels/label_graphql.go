package labels

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

// 🚨 SECURITY: TODO!(sqs): there needs to be security checks everywhere here! there are none

// gqlLabel implements the GraphQL type Label.
type gqlLabel struct{ db *dbLabel }

// labelByID looks up and returns the Label with the given GraphQL ID. If no such Label exists, it
// returns a non-nil error.
func labelByID(ctx context.Context, id graphql.ID) (*gqlLabel, error) {
	dbID, err := graphqlbackend.UnmarshalLabelID(id)
	if err != nil {
		return nil, err
	}
	return labelByDBID(ctx, dbID)
}

func (GraphQLResolver) LabelByID(ctx context.Context, id graphql.ID) (graphqlbackend.Label, error) {
	return labelByID(ctx, id)
}

// labelByDBID looks up and returns the Label with the given database ID. If no such Label exists,
// it returns a non-nil error.
func labelByDBID(ctx context.Context, dbID int64) (*gqlLabel, error) {
	v, err := dbLabels{}.GetByID(ctx, dbID)
	if err != nil {
		return nil, err
	}
	return &gqlLabel{db: v}, nil
}

func (v *gqlLabel) ID() graphql.ID {
	return graphqlbackend.MarshalLabelID(v.db.ID)
}

func (v *gqlLabel) Name() string { return v.db.Name }

func (v *gqlLabel) Description() *string { return v.db.Description }

func (v *gqlLabel) Color() string { return v.db.Color }

func (v *gqlLabel) Repository(ctx context.Context) (*graphqlbackend.RepositoryResolver, error) {
	return graphqlbackend.RepositoryByDBID(ctx, api.RepoID(v.db.RepositoryID))
}
