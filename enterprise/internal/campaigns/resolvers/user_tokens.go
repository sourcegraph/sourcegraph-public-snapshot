package resolvers

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	ee "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
)

const campaignsCredentialIDKind = "CampaignsCredential"

func marshalCampaignsCredentialID(id int64) graphql.ID {
	return relay.MarshalID(campaignsCredentialIDKind, id)
}

func unmarshalCampaignsCredentialID(id graphql.ID) (cid int64, err error) {
	err = relay.UnmarshalSpec(id, &cid)
	return
}

func (r *Resolver) CampaignsCodeHosts(ctx context.Context, args *graphqlbackend.ListCampaignsCodeHostsArgs) (graphqlbackend.CampaignsCodeHostConnectionResolver, error) {
	return &campaignsCodeHostConnectionResolver{args: args, store: r.store}, nil
}

type campaignsCredentialResolver struct {
	credential *db.UserCredential
}

var _ graphqlbackend.CampaignsCredentialResolver = &campaignsCredentialResolver{}

func (c *campaignsCredentialResolver) ID() graphql.ID {
	return marshalCampaignsCredentialID(c.credential.ID)
}

func (c *campaignsCredentialResolver) ExternalServiceKind() string {
	return extsvc.TypeToKind(c.credential.ExternalServiceType)
}

func (c *campaignsCredentialResolver) ExternalServiceURL() string {
	// This is usually the code host URL.
	return c.credential.ExternalServiceID
}

func (c *campaignsCredentialResolver) CreatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: c.credential.CreatedAt}
}

type campaignsCodeHostResolver struct {
	externalServiceKind string
	externalServiceURL  string
	credential          *db.UserCredential
}

var _ graphqlbackend.CampaignsCodeHostResolver = &campaignsCodeHostResolver{}

func (c *campaignsCodeHostResolver) ExternalServiceKind() string {
	return c.externalServiceKind
}

func (c *campaignsCodeHostResolver) ExternalServiceURL() string {
	return c.externalServiceURL
}

func (c *campaignsCodeHostResolver) Credential() graphqlbackend.CampaignsCredentialResolver {
	if c.credential != nil {
		return &campaignsCredentialResolver{credential: c.credential}
	}
	return nil
}

type campaignsCodeHostConnectionResolver struct {
	args  *graphqlbackend.ListCampaignsCodeHostsArgs
	store *ee.Store
}

var _ graphqlbackend.CampaignsCodeHostConnectionResolver = &campaignsCodeHostConnectionResolver{}

func (c *campaignsCodeHostConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	cs, err := c.compute(ctx)
	if err != nil {
		return 0, err
	}
	return int32(len(cs)), err
}

func (c *campaignsCodeHostConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(false), nil
}

func (c *campaignsCodeHostConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.CampaignsCodeHostResolver, error) {
	cs, err := c.compute(ctx)
	if err != nil {
		return nil, err
	}

	nodes := make([]graphqlbackend.CampaignsCodeHostResolver, len(cs))
	for i, ch := range cs {
		// Todo: this is n+1
		cred, err := db.UserCredentials.GetByScope(ctx, db.UserCredentialScope{
			Domain:              db.UserCredentialDomainCampaigns,
			ExternalServiceID:   ch.ExternalServiceID,
			ExternalServiceType: ch.ExternalServiceType,
			UserID:              int32(c.args.UserID),
		})
		if err != nil && !errcode.IsNotFound(err) {
			return nil, err
		}
		nodes[i] = &campaignsCodeHostResolver{externalServiceKind: extsvc.TypeToKind(ch.ExternalServiceType), externalServiceURL: ch.ExternalServiceID, credential: cred}
	}

	return nodes, nil
}

func (c *campaignsCodeHostConnectionResolver) compute(ctx context.Context) ([]*ee.CodeHost, error) {
	return c.store.GetCodeHosts(ctx, ee.GetCodeHostsOpts{})
}
