package resolvers

import (
	"context"
	"strconv"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/store"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
)

var _ graphqlbackend.CampaignsConnectionResolver = &campaignsConnectionResolver{}

type campaignsConnectionResolver struct {
	store *store.Store
	opts  store.ListCampaignsOpts

	// cache results because they are used by multiple fields
	once      sync.Once
	campaigns []*campaigns.Campaign
	next      int64
	err       error
}

func (r *campaignsConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.CampaignResolver, error) {
	nodes, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := make([]graphqlbackend.CampaignResolver, 0, len(nodes))
	for _, c := range nodes {
		resolvers = append(resolvers, &campaignResolver{store: r.store, Campaign: c})
	}
	return resolvers, nil
}

func (r *campaignsConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	opts := store.CountCampaignsOpts{
		ChangesetID:      r.opts.ChangesetID,
		State:            r.opts.State,
		InitialApplierID: r.opts.InitialApplierID,
		NamespaceUserID:  r.opts.NamespaceUserID,
		NamespaceOrgID:   r.opts.NamespaceOrgID,
	}
	count, err := r.store.CountCampaigns(ctx, opts)
	return int32(count), err
}

func (r *campaignsConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	_, next, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	if next != 0 {
		return graphqlutil.NextPageCursor(strconv.Itoa(int(next))), nil
	}
	return graphqlutil.HasNextPage(false), nil
}

func (r *campaignsConnectionResolver) compute(ctx context.Context) ([]*campaigns.Campaign, int64, error) {
	r.once.Do(func() {
		r.campaigns, r.next, r.err = r.store.ListCampaigns(ctx, r.opts)
	})
	return r.campaigns, r.next, r.err
}
