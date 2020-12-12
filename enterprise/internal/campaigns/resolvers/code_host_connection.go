package resolvers

import (
	"context"
	"strconv"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/store"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
)

type campaignsCodeHostConnectionResolver struct {
	userID                int32
	onlyWithoutCredential bool
	opts                  store.ListCodeHostsOpts
	limitOffset           db.LimitOffset
	store                 *store.Store

	once          sync.Once
	chs           []*campaigns.CodeHost
	chsPage       []*campaigns.CodeHost
	credsByIDType map[idType]*db.UserCredential
	chsErr        error
}

var _ graphqlbackend.CampaignsCodeHostConnectionResolver = &campaignsCodeHostConnectionResolver{}

func (c *campaignsCodeHostConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	chs, _, _, err := c.compute(ctx)
	if err != nil {
		return 0, err
	}
	return int32(len(chs)), err
}

func (c *campaignsCodeHostConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	chs, _, _, err := c.compute(ctx)
	if err != nil {
		return nil, err
	}

	idx := c.limitOffset.Limit + c.limitOffset.Offset
	if idx < len(chs) {
		return graphqlutil.NextPageCursor(strconv.Itoa(idx)), nil
	}

	return graphqlutil.HasNextPage(false), nil
}

func (c *campaignsCodeHostConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.CampaignsCodeHostResolver, error) {
	_, page, credsByIDType, err := c.compute(ctx)
	if err != nil {
		return nil, err
	}

	nodes := make([]graphqlbackend.CampaignsCodeHostResolver, len(page))
	for i, ch := range page {
		t := idType{
			externalServiceID:   ch.ExternalServiceID,
			externalServiceType: ch.ExternalServiceType,
		}
		cred := credsByIDType[t]
		nodes[i] = &campaignsCodeHostResolver{externalServiceKind: extsvc.TypeToKind(ch.ExternalServiceType), externalServiceURL: ch.ExternalServiceID, credential: cred}
	}

	return nodes, nil
}

func (c *campaignsCodeHostConnectionResolver) compute(ctx context.Context) (all, page []*campaigns.CodeHost, credsByIDType map[idType]*db.UserCredential, err error) {
	c.once.Do(func() {
		// Don't pass c.limitOffset here, as we want all code hosts for the totalCount anyways.
		c.chs, c.chsErr = c.store.ListCodeHosts(ctx, c.opts)
		if c.chsErr != nil {
			return
		}

		// Fetch all user credentials to avoid N+1 per credential resolver.
		creds, _, err := db.UserCredentials.List(ctx, db.UserCredentialsListOpts{Scope: db.UserCredentialScope{Domain: db.UserCredentialDomainCampaigns, UserID: c.userID}})
		if err != nil {
			c.chsErr = err
			return
		}

		c.credsByIDType = make(map[idType]*db.UserCredential)
		for _, cred := range creds {
			t := idType{
				externalServiceID:   cred.ExternalServiceID,
				externalServiceType: cred.ExternalServiceType,
			}
			c.credsByIDType[t] = cred
		}

		if c.onlyWithoutCredential {
			chs := make([]*campaigns.CodeHost, 0)
			for _, ch := range c.chs {
				t := idType{
					externalServiceID:   ch.ExternalServiceID,
					externalServiceType: ch.ExternalServiceType,
				}
				if _, ok := c.credsByIDType[t]; !ok {
					chs = append(chs, ch)
				}
			}
			c.chs = chs
		}

		afterIdx := c.limitOffset.Offset

		// Out of bound means page slice is empty.
		if afterIdx >= len(c.chs) {
			return
		}

		// Prepare page slice based on pagination params.
		limit := c.limitOffset.Limit
		// No limit set: page slice is all from `afterIdx` on.
		if limit <= 0 {
			c.chsPage = c.chs[afterIdx:]
			return
		}
		// If limit + afterIdx exceed slice bounds, cap to limit.
		if limit+afterIdx >= len(c.chs) {
			limit = len(c.chs) - afterIdx
		}
		c.chsPage = c.chs[afterIdx : limit+afterIdx]
	})
	return c.chs, c.chsPage, c.credsByIDType, c.chsErr
}

type idType struct {
	externalServiceID   string
	externalServiceType string
}

type emptyCampaignsCodeHostConnectionResolver struct{}

var _ graphqlbackend.CampaignsCodeHostConnectionResolver = &emptyCampaignsCodeHostConnectionResolver{}

func (c *emptyCampaignsCodeHostConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	return int32(0), nil
}

func (c *emptyCampaignsCodeHostConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(false), nil
}

func (c emptyCampaignsCodeHostConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.CampaignsCodeHostResolver, error) {
	return []graphqlbackend.CampaignsCodeHostResolver{}, nil
}
