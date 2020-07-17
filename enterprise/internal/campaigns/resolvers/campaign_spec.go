package resolvers

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
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
	opts := ee.ListChangesetSpecsOpts{Limit: -1, CampaignSpecID: r.campaignSpec.ID}
	cs, _, err := r.store.ListChangesetSpecs(ctx, opts)
	if err != nil {
		return nil, err
	}

	resolvers := make([]graphqlbackend.ChangesetSpecResolver, 0, len(cs))
	for _, c := range cs {
		resolvers = append(resolvers, &changesetSpecResolver{
			store:         r.store,
			httpFactory:   r.httpFactory,
			changesetSpec: c,
			repoCtx:       ctx,
		})
	}

	return &changesetSpecConnectionResolver{
		resolvers: resolvers,
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

type changesetSpecConnectionResolver struct {
	resolvers []graphqlbackend.ChangesetSpecResolver
}

func (r *changesetSpecConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	// TODO: Implement.
	return int32(len(r.resolvers)), nil
}

func (r *changesetSpecConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	// TODO: Implement.
	return &graphqlutil.PageInfo{}, nil
}

func (r *changesetSpecConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.ChangesetSpecResolver, error) {
	// TODO: Implement.
	return r.resolvers, nil
}
