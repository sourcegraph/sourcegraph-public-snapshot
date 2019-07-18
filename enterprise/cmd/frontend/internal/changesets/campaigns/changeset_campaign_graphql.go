package campaigns

import (
	"context"
	"path"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/projects"
)

// 🚨 SECURITY: TODO!(sqs): there needs to be security checks everywhere here! there are none

// gqlChangesetCampaign implements the GraphQL type ChangesetCampaign.
type gqlChangesetCampaign struct{ db *dbChangesetCampaign }

// changesetCampaignByID looks up and returns the ChangesetCampaign with the given GraphQL ID. If no such ChangesetCampaign exists, it
// returns a non-nil error.
func changesetCampaignByID(ctx context.Context, id graphql.ID) (*gqlChangesetCampaign, error) {
	dbID, err := unmarshalChangesetCampaignID(id)
	if err != nil {
		return nil, err
	}
	return changesetCampaignByDBID(ctx, dbID)
}

func (GraphQLResolver) ChangesetCampaignByID(ctx context.Context, id graphql.ID) (graphqlbackend.ChangesetCampaign, error) {
	return changesetCampaignByID(ctx, id)
}

// changesetCampaignByDBID looks up and returns the ChangesetCampaign with the given database ID. If no such ChangesetCampaign exists,
// it returns a non-nil error.
func changesetCampaignByDBID(ctx context.Context, dbID int64) (*gqlChangesetCampaign, error) {
	v, err := dbChangesetCampaigns{}.GetByID(ctx, dbID)
	if err != nil {
		return nil, err
	}
	return &gqlChangesetCampaign{db: v}, nil
}

func (v *gqlChangesetCampaign) ID() graphql.ID {
	return marshalChangesetCampaignID(v.db.ID)
}

func marshalChangesetCampaignID(id int64) graphql.ID {
	return relay.MarshalID("ChangesetCampaign", id)
}

func unmarshalChangesetCampaignID(id graphql.ID) (dbID int64, err error) {
	err = relay.UnmarshalSpec(id, &dbID)
	return
}

func (v *gqlChangesetCampaign) Project(ctx context.Context) (graphqlbackend.Project, error) {
	return graphqlbackend.ProjectByDBID(ctx, v.db.ProjectID)
}

func (v *gqlChangesetCampaign) Name() string { return v.db.Name }

func (v *gqlChangesetCampaign) Description() *string { return v.db.Description }

func (v *gqlChangesetCampaign) URL() string {
	return path.Join(projects.URLToProject(v.db.ProjectID), "campaigns", string(v.ID()))
}
