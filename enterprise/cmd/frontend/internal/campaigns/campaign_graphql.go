package campaigns

import (
	"context"
	"path"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

// 🚨 SECURITY: TODO!(sqs): there needs to be security checks everywhere here! there are none

// gqlCampaign implements the GraphQL type Campaign.
type gqlCampaign struct{ db *dbCampaign }

// campaignByID looks up and returns the Campaign with the given GraphQL ID. If no such Campaign exists, it
// returns a non-nil error.
func campaignByID(ctx context.Context, id graphql.ID) (*gqlCampaign, error) {
	dbID, err := unmarshalCampaignID(id)
	if err != nil {
		return nil, err
	}
	return campaignByDBID(ctx, dbID)
}

func (GraphQLResolver) CampaignByID(ctx context.Context, id graphql.ID) (graphqlbackend.Campaign, error) {
	return campaignByID(ctx, id)
}

// campaignByDBID looks up and returns the Campaign with the given database ID. If no such Campaign exists,
// it returns a non-nil error.
func campaignByDBID(ctx context.Context, dbID int64) (*gqlCampaign, error) {
	v, err := dbCampaigns{}.GetByID(ctx, dbID)
	if err != nil {
		return nil, err
	}
	return &gqlCampaign{db: v}, nil
}

func (v *gqlCampaign) ID() graphql.ID {
	return marshalCampaignID(v.db.ID)
}

func marshalCampaignID(id int64) graphql.ID {
	return relay.MarshalID("Campaign", id)
}

func unmarshalCampaignID(id graphql.ID) (dbID int64, err error) {
	err = relay.UnmarshalSpec(id, &dbID)
	return
}

func (v *gqlCampaign) Namespace(ctx context.Context) (*graphqlbackend.NamespaceResolver, error) {
	return graphqlbackend.NamespaceByDBID(ctx, v.db.NamespaceUserID, v.db.NamespaceOrgID)
}

func (v *gqlCampaign) Name() string { return v.db.Name }

func (v *gqlCampaign) Description() *string { return v.db.Description }

func (v *gqlCampaign) URL(ctx context.Context) (string, error) {
	namespace, err := v.Namespace(ctx)
	if err != nil {
		return "", err
	}
	return path.Join(namespace.URL(), "campaigns", string(v.ID())), nil
}
