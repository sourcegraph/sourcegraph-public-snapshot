package reconciler

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/store"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
)

type mockMissingErr struct {
	mockName string
}

func (e mockMissingErr) Error() string {
	return fmt.Sprintf("FakeStore is missing mock for %s", e.mockName)
}

type FakeStore struct {
	GetCampaignMock func(context.Context, store.GetCampaignOpts) (*campaigns.Campaign, error)
}

func (fs *FakeStore) GetCampaign(ctx context.Context, opts store.GetCampaignOpts) (*campaigns.Campaign, error) {
	if fs.GetCampaignMock != nil {
		return fs.GetCampaignMock(ctx, opts)
	}
	return nil, mockMissingErr{"GetCampaign"}
}
