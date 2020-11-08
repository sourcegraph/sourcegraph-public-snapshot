package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	ee "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
)

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
