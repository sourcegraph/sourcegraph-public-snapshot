package resolvers

import (
	"context"
	"strconv"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	ee "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
)

type campaignsCodeHostConnectionResolver struct {
	userID      int32
	limitOffset db.LimitOffset
	store       *ee.Store

	once    sync.Once
	chs     []*campaigns.CodeHost
	chsPage []*campaigns.CodeHost
	chsErr  error
}

var _ graphqlbackend.CampaignsCodeHostConnectionResolver = &campaignsCodeHostConnectionResolver{}

func (c *campaignsCodeHostConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	chs, _, err := c.compute(ctx)
	if err != nil {
		return 0, err
	}
	return int32(len(chs)), err
}

func (c *campaignsCodeHostConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	chs, _, err := c.compute(ctx)
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
	_, page, err := c.compute(ctx)
	if err != nil {
		return nil, err
	}

	// Fetch all user credentials to avoid N+1 per credential resolver.
	creds, _, err := db.UserCredentials.List(ctx, db.UserCredentialsListOpts{Scope: db.UserCredentialScope{Domain: db.UserCredentialDomainCampaigns, UserID: c.userID}})
	if err != nil {
		return nil, err
	}

	credsByIDType := make(map[idType]*db.UserCredential)
	for _, cred := range creds {
		t := idType{
			externalServiceID:   cred.ExternalServiceID,
			externalServiceType: cred.ExternalServiceType,
		}
		credsByIDType[t] = cred
	}

	nodes := make([]graphqlbackend.CampaignsCodeHostResolver, len(page))
	for i, ch := range page {
		t := idType{
			externalServiceID:   ch.ExternalServiceID,
			externalServiceType: ch.ExternalServiceType,
		}
		nodes[i] = &campaignsCodeHostResolver{externalServiceKind: extsvc.TypeToKind(ch.ExternalServiceType), externalServiceURL: ch.ExternalServiceID, credential: credsByIDType[t]}
	}

	return nodes, nil
}

func (c *campaignsCodeHostConnectionResolver) compute(ctx context.Context) (all, page []*campaigns.CodeHost, err error) {
	c.once.Do(func() {
		// Don't pass c.limitOffset here, as we want all code hosts for the totalCount anyways.
		c.chs, c.chsErr = c.store.ListCodeHosts(ctx)

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
	return c.chs, c.chsPage, c.chsErr
}

type idType struct {
	externalServiceID   string
	externalServiceType string
}
