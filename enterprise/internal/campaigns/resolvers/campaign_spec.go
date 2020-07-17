package resolvers

import (
	"context"
	"strconv"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	ee "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
)

func marshalCampaignSpecRandID(id string) graphql.ID {
	return relay.MarshalID("CampaignSpec", id)
}

func unmarshalCampaignSpecID(id graphql.ID) (campaignSpecRandID string, err error) {
	err = relay.UnmarshalSpec(id, &campaignSpecRandID)
	return
}

var _ graphqlbackend.CampaignSpecResolver = &campaignSpecResolver{}

type campaignSpecResolver struct {
	store       *ee.Store
	httpFactory *httpcli.Factory

	campaignSpec *campaigns.CampaignSpec
}

func (r *campaignSpecResolver) ID() graphql.ID {
	// 🚨 SECURITY: This needs to be the RandID! We can't expose the
	// sequential, guessable ID.
	return marshalCampaignSpecRandID(r.campaignSpec.RandID)
}

func (r *campaignSpecResolver) OriginalInput() (string, error) {
	return r.campaignSpec.RawSpec, nil
}

func (r *campaignSpecResolver) ParsedInput() (graphqlbackend.JSONValue, error) {
	return graphqlbackend.JSONValue{Value: r.campaignSpec.Spec}, nil
}

func (r *campaignSpecResolver) ChangesetSpecs(ctx context.Context, args *graphqlbackend.ChangesetSpecsConnectionArgs) (graphqlbackend.ChangesetSpecConnectionResolver, error) {
	opts := ee.ListChangesetSpecsOpts{CampaignSpecID: r.campaignSpec.ID}
	if args.First != nil {
		opts.Limit = int(*args.First)
	}
	if args.After != nil {
		id, err := strconv.Atoi(*args.After)
		if err != nil {
			return nil, err
		}
		opts.Cursor = int64(id)
	}

	return &changesetSpecConnectionResolver{
		store:       r.store,
		httpFactory: r.httpFactory,
		opts:        opts,
	}, nil
}

func (r *campaignSpecResolver) Description() graphqlbackend.CampaignDescriptionResolver {
	return &campaignDescriptionResolver{
		name:        r.campaignSpec.Spec.Name,
		description: r.campaignSpec.Spec.Description,
	}
}

func (r *campaignSpecResolver) Creator(ctx context.Context) (*graphqlbackend.UserResolver, error) {
	return graphqlbackend.UserByIDInt32(ctx, r.campaignSpec.UserID)
}

func (r *campaignSpecResolver) Namespace(ctx context.Context) (*graphqlbackend.NamespaceResolver, error) {
	var (
		err error
		n   = &graphqlbackend.NamespaceResolver{}
	)

	if r.campaignSpec.NamespaceUserID != 0 {
		n.Namespace, err = graphqlbackend.UserByIDInt32(ctx, r.campaignSpec.NamespaceUserID)
	} else {
		n.Namespace, err = graphqlbackend.OrgByIDInt32(ctx, r.campaignSpec.NamespaceOrgID)
	}

	if errcode.IsNotFound(err) {
		return nil, nil
	}

	return n, err
}

func (r *campaignSpecResolver) PreviewURL() (string, error) {
	return "/campaigns/new?spec=" + string(r.ID()), nil
}

func (r *campaignSpecResolver) CreatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: r.campaignSpec.CreatedAt}
}

func (r *campaignSpecResolver) ExpiresAt() *graphqlbackend.DateTime {
	return &graphqlbackend.DateTime{Time: r.campaignSpec.ExpiresAt()}
}

func (r *campaignSpecResolver) ViewerCanAdminister(ctx context.Context) (bool, error) {
	return checkSiteAdminOrSameUser(ctx, r.campaignSpec.UserID)
}

type campaignDescriptionResolver struct {
	name, description string
}

func (r *campaignDescriptionResolver) Name() string {
	return r.name
}

func (r *campaignDescriptionResolver) Description() string {
	return r.description
}
